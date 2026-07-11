package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/KronusRodion/task-tracker/internal/ctxkeys"
	"github.com/KronusRodion/task-tracker/internal/domain"
)

func genLoggerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Получаем пользователя из контекста
			userID := "anonymous"
			if userCtx, ok := r.Context().Value(ctxkeys.UserKey).(domain.UserCtx); ok {
				userID = userCtx.ID.String()
			}

			next.ServeHTTP(wrapped, r)

			slog.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"user_id", userID,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
