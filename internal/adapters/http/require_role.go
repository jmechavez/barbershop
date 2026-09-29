package http

import (
	"net/http"

	"barbershop/internal/domain"
)

// RequireRole returns middleware that rejects requests whose current
// user does not have the given role. It assumes LoadUser ran first.
func (h *Handlers) RequireRole(role domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := CurrentUser(r.Context())
			if !ok || user.Role != role {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
