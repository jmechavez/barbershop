package http

import (
	"barbershop/internal/domain"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()
	r.Use(h.LoadUser)

	r.Get("/", h.Public)
	r.Post("/pick", h.Pick)
	r.Get("/login", h.LoginForm)
	r.Post("/login", h.LoginSubmit)
	r.Post("/logout", h.Logout)

	// Admin area
	r.Group(func(r chi.Router) {
		r.Use(h.RequireRole(domain.RoleAdmin))
		r.Get("/admin", h.AdminHomePage)
		r.Post("/admin/services", h.AdminCreateService)
		r.Post("/admin/services/{id}", h.AdminUpdateService)
		r.Post("/admin/users", h.AdminCreateUser)

	})

	// Barber area
	r.Group(func(r chi.Router) {
		r.Use(h.RequireLogin)
		r.Get("/barber", h.BarberHome)
	})

	return r
}
