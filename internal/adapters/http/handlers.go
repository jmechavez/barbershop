package http

import (
	"fmt"
	"net/http"
	"strconv"

	"barbershop/internal/domain"
	"barbershop/internal/ports"
)

type Handlers struct {
	Services ports.ServiceRepository
}

func (h *Handlers) Public(w http.ResponseWriter, r *http.Request) {
	services, err := h.Services.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	PublicPage(services).Render(r.Context(), w)
}

// Pick receives the checked prices from htmx and returns the total.
// It calls domain.TotalCentavos — the HTTP layer does not do math.
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
