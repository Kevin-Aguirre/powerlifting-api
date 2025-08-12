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
		} else {
			fmt.Println("GET /lifters/" + lifterName)
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(lifter); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}
	}
}


func GetMeets(db *data.Database) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		fmt.Println("GET /meets")

		// Collect Names 
		meets := make([]model.Meet, 0)
		for fed := range db.FederationMeets {
			for meet := range db.FederationMeets[fed] {
				meets = append(
					meets, 
					*db.FederationMeets[fed][meet],
				)

			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(meets); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}
	}
}

func GetMeet(db *data.Database) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		federationNameEncoded := chi.URLParam(r, "federationName")
		federationName, err := url.QueryUnescape(federationNameEncoded)

		if err != nil {
			http.Error(w, "invalid federation name", http.StatusBadRequest)
			return 
		}

		federationMeets, exists := db.LifterHistory[federationName]

		if !exists {
			http.Error(w, "federation not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(federationMeets); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}
	}
}
