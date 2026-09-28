package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.Public)
	r.Post("/pick", h.Pick)
	return r
}
