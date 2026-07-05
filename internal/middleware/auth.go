package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/KronusRodion/task-tracker/internal/ctxkeys"
	"github.com/KronusRodion/task-tracker/internal/domain"
)

type Authenticator interface {
	Authenticate(token string) (*domain.UserCtx, error)
}

func genAuthMiddleware(parser Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token, err := extractToken(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			claims, err := parser.Authenticate(token)
			if err != nil {
				http.Error(w, "invalid access token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				ctxkeys.UserKey,
				&domain.UserCtx{
					ID: claims.ID,
				},
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}

func extractToken(r *http.Request) (string, error) {

	auth := r.Header.Get("Authorization")
	if auth != "" {

		parts := strings.SplitN(auth, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {
			return "", errors.New("invalid authorization header")
		}

		return parts[1], nil
	}

	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return "", errors.New("authorization token not found")
}
