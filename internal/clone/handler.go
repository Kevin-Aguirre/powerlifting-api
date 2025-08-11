package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	// "sort"

	"github.com/Kevin-Aguirre/powerlifting-api/data"
	"github.com/Kevin-Aguirre/powerlifting-api/model"
	"github.com/go-chi/chi/v5"
)

func GetLifters(db *data.Database) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET /lifters")

		// Collect Names 
		lifters := make([]model.Lifter, 0, len(db.LifterHistory))
		for i := range db.LifterHistory {
			lifters = append(lifters, *db.LifterHistory[i])
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(lifters); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}
	}
}

func GetLifter(db *data.Database) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		lifterNameEncoded := chi.URLParam(r, "lifterName")
		lifterName, err := url.QueryUnescape(lifterNameEncoded)

		if err != nil {
			http.Error(w, "invalid lifter name", http.StatusBadRequest)
			return 
		}

		lifter, exists := db.LifterHistory[lifterName]

		if !exists {
			http.Error(w, "lifter not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(lifter); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}
	}
}