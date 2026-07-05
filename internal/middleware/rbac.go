package middleware

import (
	"context"
	"net/http"

	"github.com/KronusRodion/task-tracker/internal/ctxkeys"
	"github.com/KronusRodion/task-tracker/internal/domain"
	"github.com/google/uuid"
)

type PermissionChecker interface {
	HasPermission(
		ctx context.Context,
		userID uuid.UUID,
		permission string,
	) (bool, error)
}

type RBCMiddleware func(next http.Handler, perms ...string) http.Handler

func GenRbacMiddleware(
	checker PermissionChecker,
) RBCMiddleware {

	return func(next http.Handler, perms ...string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			user, ok := r.Context().Value(ctxkeys.UserKey).(*domain.UserCtx)
			if !ok {
				http.Error(w, "user not authenticated", http.StatusUnauthorized)
				return
			}

			for _, p := range perms {
				ok, err := checker.HasPermission(r.Context(), user.ID, p)
				if err != nil {
					http.Error(w, "permission check failed", http.StatusInternalServerError)
					return
				}
				if !ok {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
