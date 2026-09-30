package http

import (
	"fmt"
	"net/http"
	"strconv"

	"barbershop/internal/app"
	"barbershop/internal/domain"
	"barbershop/internal/ports"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Services ports.ServiceRepository
	Users    ports.UserRepository
	Haircuts *app.HaircutService
	Auth     *app.AuthService
	Sessions *app.SessionService
}

func (h *Handlers) Public(w http.ResponseWriter, r *http.Request) {
	services, err := h.Services.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, loggedIn := CurrentUser(r.Context())
	PublicPage(services, user, loggedIn).Render(r.Context(), w)
}

func (h *Handlers) Pick(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var picked []domain.Service
	for _, v := range r.Form["service"] {
		n, _ := strconv.Atoi(v)
		picked = append(picked, domain.Service{PriceCentavos: n})
	}

	total := domain.TotalCentavos(picked)
	fmt.Fprintf(w, "Total: %s", domain.FormatCentavos(total))
}

func (h *Handlers) LoginForm(w http.ResponseWriter, r *http.Request) {
	LoginPage("").Render(r.Context(), w)
}

func (h *Handlers) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	fmt.Printf("DEBUG: form email=%q password=%q\n", email, password)

	user, err := h.Auth.Authenticate(r.Context(), email, password)
	if err != nil {
		fmt.Printf("DEBUG: Authenticate failed: %v\n", err)
		msg := "Invalid email or password."
		LoginPage(msg).Render(r.Context(), w)
		return
	}

	fmt.Printf("DEBUG: Authenticate OK user=%d\n", user.ID)

	token, err := h.Sessions.Start(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setSessionCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	token := sessionToken(r)
	if err := h.Sessions.Stop(r.Context(), token); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	clearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) AdminHomePage(w http.ResponseWriter, r *http.Request) {
	user, _ := CurrentUser(r.Context())

	services, err := h.Services.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	users, err := h.Users.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	AdminHome(user, services, users).Render(r.Context(), w)
}

func (h *Handlers) AdminUpdateService(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	priceStr := r.FormValue("price_pesos")
	pricePesos, err := strconv.Atoi(priceStr)
	if err != nil || pricePesos < 0 {
		http.Error(w, "bad price", http.StatusBadRequest)
		return
	}
	priceCentavos := pricePesos * 100

	if err := h.Services.Update(r.Context(), id, name, priceCentavos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	services, err := h.Services.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, s := range services {
		if s.ID == id {
			ServiceRow(s).Render(r.Context(), w)
			return
		}
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func (h *Handlers) AdminCreateService(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	priceStr := r.FormValue("price_pesos")
	pricePesos, err := strconv.Atoi(priceStr)
	if err != nil || pricePesos < 0 {
		http.Error(w, "bad price", http.StatusBadRequest)
		return
	}
	priceCentavos := pricePesos * 100

	if _, err := h.Services.Create(r.Context(), name, priceCentavos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handlers) AdminCreateUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	fullName := r.FormValue("full_name")
	password := r.FormValue("password")
	roleStr := r.FormValue("role")
	floorStr := r.FormValue("daily_floor_pesos")

	if email == "" || fullName == "" || password == "" {
		http.Error(w, "email, full name and password required", http.StatusBadRequest)
		return
	}

	role := domain.Role(roleStr)
	if !role.Valid() {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	floorPesos, err := strconv.Atoi(floorStr)
	if err != nil || floorPesos < 0 {
		http.Error(w, "bad floor", http.StatusBadRequest)
		return
	}
	floorCentavos := floorPesos * 100

	if _, err := h.Auth.CreateUser(r.Context(), email, fullName, password, role, floorCentavos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
