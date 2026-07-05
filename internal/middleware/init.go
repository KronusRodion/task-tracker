package middleware

import "net/http"

var (
	Auth   func(http.Handler) http.Handler
	Logger func(http.Handler) http.Handler
	Rbac   RBCMiddleware
)

func Init(
	parser Authenticator,
	checker PermissionChecker,
) {

	Auth = genAuthMiddleware(parser)
	Logger = genLoggerMiddleware()
	Rbac = GenRbacMiddleware(checker)
}
