package http

import (
	"context"
	"net/http"

	"barbershop/internal/domain"
)

type contextKey int

const userContextKey contextKey = iota

func CurrentUser(ctx context.Context) (domain.User, bool) {
	u, ok := ctx.Value(userContextKey).(domain.User)
	return u, ok
}

func (h *Handlers) LoadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := sessionToken(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := h.Sessions.UserForToken(r.Context(), token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
