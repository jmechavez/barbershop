package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"barbershop/internal/app"
	"barbershop/internal/domain"
	"barbershop/internal/ports"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Services     ports.ServiceRepository
	Users        ports.UserRepository
	Haircuts     *app.HaircutService
	CashAdvances *app.CashAdvanceService
	Salary       *app.SalaryService
	Auth         *app.AuthService
	Sessions     *app.SessionService
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

	barberName := map[int]string{}
	for _, u := range users {
		barberName[u.ID] = u.FullName
	}

	// Advances: combine recent advances for all barbers.
	// Simple approach: fetch each barber's last 30 days.
	var advances []domain.CashAdvance
	for _, u := range users {
		if u.Role != domain.RoleBarber {
			continue
		}
		as, err := h.CashAdvances.RecentForBarber(r.Context(), u.ID, 30)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		advances = append(advances, as...)
	}

	AdminHome(user, services, users, advances, barberName).Render(r.Context(), w)
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

func (h *Handlers) CounterHome(w http.ResponseWriter, r *http.Request) {
	h.renderCounter(w, r, "")
}

// renderCounter is the shared body of GET /counter and of the POST
// handler's failure path: build the page data and render it.
func (h *Handlers) renderCounter(w http.ResponseWriter, r *http.Request, errorMsg string) {
	admin, _ := CurrentUser(r.Context())

	// Fetch all users, filter to barbers.
	allUsers, err := h.Users.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var barbers []domain.User
	barberName := map[int]string{}
	for _, u := range allUsers {
		if u.Role == domain.RoleBarber {
			barbers = append(barbers, u)
			barberName[u.ID] = u.FullName
		}
	}

	// Fetch services.
	services, err := h.Services.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serviceName := map[int]string{}
	for _, s := range services {
		serviceName[s.ID] = s.Name
	}

	// Fetch today's haircuts.
	haircuts, err := h.Haircuts.TodayAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	CounterPage(admin, barbers, services, haircuts, barberName, serviceName, errorMsg).Render(r.Context(), w)
}

func (h *Handlers) CounterRecordHaircut(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	barberID, err := strconv.Atoi(r.FormValue("barber_id"))
	if err != nil {
		http.Error(w, "bad barber", http.StatusBadRequest)
		return
	}

	serviceID, err := strconv.Atoi(r.FormValue("service_id"))
	if err != nil {
		http.Error(w, "bad service", http.StatusBadRequest)
		return
	}

	// Discount (optional, in pesos)
	discountPesos := 0
	if s := r.FormValue("discount_pesos"); s != "" {
		discountPesos, err = strconv.Atoi(s)
		if err != nil || discountPesos < 0 {
			h.renderCounter(w, r, "invalid discount")
			return
		}
	}
	discountCentavos := discountPesos * 100
	discountReason := r.FormValue("discount_reason")

	// Payments
	payments, err := parsePayments(r)
	if err != nil {
		h.renderCounter(w, r, err.Error())
		return
	}

	// Record.
	if _, err := h.Haircuts.Record(r.Context(), barberID, serviceID, discountCentavos, discountReason, payments); err != nil {
		h.renderCounter(w, r, err.Error())
		return
	}

	http.Redirect(w, r, "/counter", http.StatusSeeOther)
}

// parsePayments reads pay_gcash, pay_maribank, pay_cash from the form.
// Each is a whole-peso amount. Zero amounts are skipped. Any non-zero
// amount becomes a domain.Payment.
func parsePayments(r *http.Request) ([]domain.Payment, error) {
	type entry struct {
		field  string
		method domain.PaymentMethod
	}
	fields := []entry{
		{"pay_gcash", domain.PaymentGCash},
		{"pay_maribank", domain.PaymentMaribank},
		{"pay_cash", domain.PaymentCash},
	}

	var out []domain.Payment
	for _, e := range fields {
		str := r.FormValue(e.field)
		if str == "" {
			continue
		}
		pesos, err := strconv.Atoi(str)
		if err != nil || pesos < 0 {
			return nil, errors.New("invalid payment amount")
		}
		if pesos == 0 {
			continue
		}
		out = append(out, domain.Payment{
			Method:         e.method,
			AmountCentavos: pesos * 100,
		})
	}
	return out, nil
}

func (h *Handlers) MeHome(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	haircuts, err := h.Haircuts.Today(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	services, err := h.Services.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serviceName := map[int]string{}
	for _, s := range services {
		serviceName[s.ID] = s.Name
	}

	advances, err := h.CashAdvances.RecentForBarber(r.Context(), user.ID, 30)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	MePage(user, haircuts, serviceName, advances).Render(r.Context(), w)
}

func (h *Handlers) AdminCreateCashAdvance(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	barberID, err := strconv.Atoi(r.FormValue("barber_id"))
	if err != nil {
		http.Error(w, "bad barber", http.StatusBadRequest)
		return
	}

	amountPesos, err := strconv.Atoi(r.FormValue("amount_pesos"))
	if err != nil || amountPesos <= 0 {
		http.Error(w, "bad amount", http.StatusBadRequest)
		return
	}
	amountCentavos := amountPesos * 100

	note := r.FormValue("note")

	if _, err := h.CashAdvances.Record(r.Context(), barberID, amountCentavos, note); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handlers) MeSalaryPage(w http.ResponseWriter, r *http.Request) {
	user, ok := CurrentUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Only barbers have a salary.
	if user.Role != domain.RoleBarber {
		http.Error(w, "salary is for barbers only", http.StatusForbidden)
		return
	}

	// Read date range from query string.
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	// Default: first day of current month to today.
	now := time.Now()
	if fromStr == "" {
		fromStr = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}
	if toStr == "" {
		toStr = now.Format("2006-01-02")
	}

	from, err := time.ParseInLocation("2006-01-02", fromStr, time.Local)
	if err != nil {
		http.Error(w, "bad from date", http.StatusBadRequest)
		return
	}
	to, err := time.ParseInLocation("2006-01-02", toStr, time.Local)
	if err != nil {
		http.Error(w, "bad to date", http.StatusBadRequest)
		return
	}
	if to.Before(from) {
		http.Error(w, "to date is before from date", http.StatusBadRequest)
		return
	}

	// The salary range is inclusive of the `to` date.
	// We pass `to` as midnight of the day *after*, because the SQL
	// query uses created_at < to.
	toExclusive := to.AddDate(0, 0, 1)

	summary, err := h.Salary.ForRange(r.Context(), user.ID, from, toExclusive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	SalaryPage(user, summary, fromStr, toStr).Render(r.Context(), w)
}
