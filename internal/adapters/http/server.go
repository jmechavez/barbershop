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

	// Admin area (admin only)
	r.Group(func(r chi.Router) {
		r.Use(h.RequireRole(domain.RoleAdmin))

		// Admin home
		r.Get("/admin", h.AdminHomePage)

		// Services
		r.Post("/admin/services", h.AdminCreateService)
		r.Post("/admin/services/{id}", h.AdminUpdateService)
		r.Delete("/admin/services/{id}", h.AdminDeleteService)

		// Users
		r.Post("/admin/users", h.AdminCreateUser)
		r.Post("/admin/users/{id}/password", h.AdminResetPassword)

		// Cash advances
		r.Post("/admin/cash-advances", h.AdminCreateCashAdvance)
		r.Get("/admin/cash-advances", h.AdminCashAdvancesPage)

		// Haircuts
		r.Get("/admin/haircuts", h.AdminHaircutsPage)

		// Salaries
		r.Get("/admin/salaries", h.AdminSalariesPage)

		// Counter
		r.Get("/counter", h.CounterHome)
		r.Post("/counter/haircuts", h.CounterRecordHaircut)
		r.Post("/counter/attendance", h.CounterSaveAttendance)
	})

	// Logged-in pages (any role)
	r.Group(func(r chi.Router) {
		r.Use(h.RequireLogin)
		r.Get("/me", h.MeHome)
		r.Get("/me/salary", h.MeSalaryPage)
	})

	return r
}
