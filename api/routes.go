package api

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/Kevin-Aguirre/powerlifting-api/internal/clone"
	"github.com/Kevin-Aguirre/powerlifting-api/data"
)

func NewRouter(db *data.Database) http.Handler {

	r := chi.NewRouter()
	r.Get("/lifters", handlers.GetHello(db))
	r.Get("/meets", handlers.GetHello(db))
	r.Get("/federations", handlers.GetHello(db))

	return r
}