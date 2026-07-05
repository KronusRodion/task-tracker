package middleware

import (
	"log"
	"net/http"
)

func genLoggerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("Begin request: ", r.URL)
			next.ServeHTTP(w, r)
		})
	}
}
