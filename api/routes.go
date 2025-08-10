package api

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/Kevin-Aguirre/powerlifting-api/internal/clone"
	"github.com/Kevin-Aguirre/powerlifting-api/data"
)

func NewRouter(db *data.Database) http.Handler {
	r := chi.NewRouter()
	r.Get("/lifters", handlers.GetLifters(db))
	r.Get("/lifters/{lifterName}", handlers.GetLifter(db))

	return r
}