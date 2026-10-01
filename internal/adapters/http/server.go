package http

import (
	"net/http"

	"barbershop/internal/domain"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()
	r.Use(h.LoadUser)

	// Static files
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// Public routes
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
		r.Post("/admin/cash-advances", h.AdminCreateCashAdvance)
		r.Delete("/admin/services/{id}", h.AdminDeleteService)
		r.Post("/admin/users/{id}/password", h.AdminResetPassword)
		r.Get("/admin/haircuts", h.AdminHaircutsPage)

		// Counter (recording)
		r.Get("/counter", h.CounterHome)
		r.Post("/counter/haircuts", h.CounterRecordHaircut)
	})

	// Logged-in pages
	r.Group(func(r chi.Router) {
		r.Use(h.RequireLogin)
		r.Get("/me", h.MeHome)
		r.Get("/me/salary", h.MeSalaryPage)
	})

	return r
}
