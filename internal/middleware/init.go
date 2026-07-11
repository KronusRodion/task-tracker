package middleware

import "net/http"

var (
	Auth   func(http.Handler) http.Handler
	Logger func(http.Handler) http.Handler = genLoggerMiddleware()
)

func InitAuth(
	parser Authenticator,
) {

	Auth = genAuthMiddleware(parser)
}
