package api

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/Kevin-Aguirre/powerlifting-api/internal/clone"
	"github.com/Kevin-Aguirre/powerlifting-api/data"
)

func NewRouter(db *data.Database) http.Handler {
	r := chi.NewRouter()

	/*						GET Methods						*/
	// lifters
	r.Get("/lifters", 					handlers.GetLifters(db))
	r.Get("/lifters/names", 			handlers.GetLifterNames(db))
	r.Get("/lifters/{lifterName}", 		handlers.GetLifter(db))
	
	// meets
	r.Get("/meets", 					handlers.GetMeets(db))
	r.Get("/meets/{federationName}",	handlers.GetMeet(db))
	
	// federations 
	r.Get("/federations", 				handlers.GetFederations(db))

	return r
}
