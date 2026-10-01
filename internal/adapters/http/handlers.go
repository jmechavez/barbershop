package http

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
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
	fmt.Fprintf(w, "%s", domain.FormatCentavos(total))
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

	// Send the user to the page they'll actually use.
	if user.Role == domain.RoleAdmin {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/me", http.StatusSeeOther)
	}
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

	// Collect recent advances, then keep only the newest 5 overall.
	var allAdvances []domain.CashAdvance
	for _, u := range users {
		if u.Role != domain.RoleBarber {
			continue
		}
		as, err := h.CashAdvances.RecentForBarber(r.Context(), u.ID, 30)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		allAdvances = append(allAdvances, as...)
	}

	sort.Slice(allAdvances, func(i, j int) bool {
		return allAdvances[i].TakenAt.After(allAdvances[j].TakenAt)
	})
	if len(allAdvances) > 5 {
		allAdvances = allAdvances[:5]
	}

	AdminHome(user, services, users, allAdvances, barberName).Render(r.Context(), w)
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

func (h *Handlers) renderCounter(w http.ResponseWriter, r *http.Request, errorMsg string) {
	admin, _ := CurrentUser(r.Context())

	allUsers, err := h.Users.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var barbers []domain.User
	barberName := map[int]string{}
	for _, u := range allUsers {
		barberName[u.ID] = u.FullName
		if u.Role == domain.RoleBarber {
			barbers = append(barbers, u)
		}
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

func (h *Handlers) AdminDeleteService(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if err := h.Services.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// htmx will remove the row from the page.
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) AdminResetPassword(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad user id", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	newPassword := r.FormValue("new_password")
	if len(newPassword) < 6 {
		http.Error(w, "password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	if err := h.Auth.ResetPassword(r.Context(), userID, newPassword); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handlers) AdminHaircutsPage(w http.ResponseWriter, r *http.Request) {
	admin, _ := CurrentUser(r.Context())

	// All users (to build the barber filter dropdown and name map).
	allUsers, err := h.Users.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var barbers []domain.User
	barberName := map[int]string{}
	for _, u := range allUsers {
		barberName[u.ID] = u.FullName
		if u.Role == domain.RoleBarber {
			barbers = append(barbers, u)
		}
	}

	// Services (for name display).
	services, err := h.Services.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serviceName := map[int]string{}
	for _, s := range services {
		serviceName[s.ID] = s.Name
	}

	// Date range and barber filter from query string.
	now := time.Now()
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	barberStr := r.URL.Query().Get("barber_id")

	if fromStr == "" {
		fromStr = now.Format("2006-01-02")
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
	toExclusive := to.AddDate(0, 0, 1)

	selectedBarberID := 0
	if barberStr != "" {
		selectedBarberID, _ = strconv.Atoi(barberStr)
	}

	// Fetch haircuts.
	var haircuts []domain.Haircut
	if selectedBarberID > 0 {
		haircuts, err = h.Haircuts.RangeForBarber(r.Context(), selectedBarberID, from, toExclusive)
	} else {
		haircuts, err = h.Haircuts.Range(r.Context(), from, toExclusive)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Newest first, always — regardless of what the repo returned.
	sort.Slice(haircuts, func(i, j int) bool {
		return haircuts[i].CreatedAt.After(haircuts[j].CreatedAt)
	})

	HaircutsPage(
		admin, barbers, haircuts,
		barberName, serviceName,
		fromStr, toStr, selectedBarberID,
	).Render(r.Context(), w)
}

func (h *Handlers) AdminCashAdvancesPage(w http.ResponseWriter, r *http.Request) {
	admin, _ := CurrentUser(r.Context())

	// Barbers for the filter dropdown.
	allUsers, err := h.Users.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var barbers []domain.User
	barberName := map[int]string{}
	for _, u := range allUsers {
		barberName[u.ID] = u.FullName
		if u.Role == domain.RoleBarber {
			barbers = append(barbers, u)
		}
	}

	// Date range from query string.
	now := time.Now()
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	barberStr := r.URL.Query().Get("barber_id")

	if fromStr == "" {
		fromStr = now.AddDate(0, 0, -30).Format("2006-01-02")
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
	toExclusive := to.AddDate(0, 0, 1)

	selectedBarberID := 0
	if barberStr != "" {
		selectedBarberID, _ = strconv.Atoi(barberStr)
	}

	// Fetch advances. If a barber is selected, use that one only.
	// Otherwise, gather from all barbers.
	var advances []domain.CashAdvance
	if selectedBarberID > 0 {
		as, err := h.CashAdvances.ListRange(r.Context(), selectedBarberID, from, toExclusive)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		advances = as
	} else {
		for _, b := range barbers {
			as, err := h.CashAdvances.ListRange(r.Context(), b.ID, from, toExclusive)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			advances = append(advances, as...)
		}
	}

	// Newest first.
	sort.Slice(advances, func(i, j int) bool {
		return advances[i].TakenAt.After(advances[j].TakenAt)
	})

	total := domain.SumAdvances(advances)

	CashAdvancesPage(
		admin, barbers, advances,
		barberName, fromStr, toStr, selectedBarberID, total,
	).Render(r.Context(), w)
}
