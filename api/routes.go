package api

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/Kevin-Aguirre/powerlifting-api/internal/clone"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", handlers.GetRoot)
	r.Get("/hello", handlers.GetHello)
	return r
}