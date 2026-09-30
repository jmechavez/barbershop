package http

import (
	"net/http"
)

// RequireLogin returns middleware that redirects to /login if there's
// no current user. It assumes LoadUser ran first.
func (h *Handlers) RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := CurrentUser(r.Context()); !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
