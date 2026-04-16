package api

import (
	"encoding/json"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/Kevin-Aguirre/powerlifting-api/internal/clone"
	"github.com/Kevin-Aguirre/powerlifting-api/data"
)

func NewRouter(ds *data.DataStore) http.Handler {
	r := chi.NewRouter()

	// api index
	r.Get("/", apiIndex)

	// lifters
	r.Get("/lifters",                   handlers.GetLifters(ds))
	r.Get("/lifters/names",             handlers.GetLifterNames(ds))
	r.Get("/lifters/search",            handlers.SearchLifters(ds))
	r.Get("/lifters/top",               handlers.GetTopLifters(ds))
	r.Get("/lifters/{lifterName}",      handlers.GetLifter(ds))

	// meets
	r.Get("/meets",                                        handlers.GetMeets(ds))
	r.Get("/meets/{federationName}",                       handlers.GetMeet(ds))
	r.Get("/meets/{federationName}/{meetName}/results",    handlers.GetMeetResults(ds))

	// records
	r.Get("/records",                   handlers.GetRecords(ds))

	// federations
	r.Get("/federations",               handlers.GetFederations(ds))

	return r
}

func apiIndex(w http.ResponseWriter, r *http.Request) {
	index := map[string]interface{}{
		"name":    "Powerlifting API",
		"version": "1.0",
		"endpoints": map[string]string{
			"GET /lifters":                                     "List all lifters (paginated)",
			"GET /lifters/names":                               "List all lifter names (paginated)",
			"GET /lifters/search?q={query}":                    "Search lifters by name (paginated)",
			"GET /lifters/top?sex=M&equipment=Raw&weightClass=83": "Top lifters ranked by DOTS (paginated)",
			"GET /lifters/{lifterName}":                        "Get a single lifter by name",
			"GET /meets":                                       "List all meets (paginated)",
			"GET /meets/{federationName}":                      "List meets by federation (paginated)",
			"GET /meets/{federationName}/{meetName}/results":   "Get all entries for a specific meet (paginated)",
			"GET /records?sex=M&equipment=Raw&weightClass=83":  "All-time records by weight class (paginated)",
			"GET /federations":                                 "List all federation names (paginated)",
		},
		"pagination": map[string]string{
			"limit":   "Number of results per page (default: 50, max: 200)",
			"offset":  "Number of results to skip (default: 0)",
			"example": "/lifters?limit=20&offset=100",
		},
		"unit_conversion": map[string]string{
			"parameter": "unit",
			"values":    "kg (default), lbs",
			"example":   "/lifters/John%20Doe?unit=lbs",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(index)
}

