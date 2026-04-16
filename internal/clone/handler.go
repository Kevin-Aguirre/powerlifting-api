package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/Kevin-Aguirre/powerlifting-api/data"
	"github.com/Kevin-Aguirre/powerlifting-api/model"
	"github.com/go-chi/chi/v5"
)

func GetLifters(ds *data.DataStore) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		fmt.Println("GET /lifters")
		p := model.ParsePagination(r)
		toLbs := model.ParseUnit(r)

		lifters := make([]model.Lifter, 0, len(db.LifterHistory))
		for i := range db.LifterHistory {
			if toLbs {
				lifters = append(lifters, db.LifterHistory[i].ToLbs())
			} else {
				lifters = append(lifters, *db.LifterHistory[i])
			}
		}

		total := len(lifters)
		start := p.Offset
		if start > total {
			start = total
		}
		end := start + p.Limit
		if end > total {
			end = total
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.PaginatedResponse{
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
			Data:   lifters[start:end],
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetLifterNames(ds *data.DataStore) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		fmt.Println("GET /lifters/names")
		p := model.ParsePagination(r)

		names := make([]string, 0, len(db.LifterHistory))
		for name := range db.LifterHistory {
			names = append(names, name)
		}

		total := len(names)
		start := p.Offset
		if start > total {
			start = total
		}
		end := start + p.Limit
		if end > total {
			end = total
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.PaginatedResponse{
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
			Data:   names[start:end],
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func SearchLifters(ds *data.DataStore) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		q := strings.ToLower(r.URL.Query().Get("q"))
		if q == "" {
			http.Error(w, "missing required query parameter: q", http.StatusBadRequest)
			return
		}

		fmt.Println("GET /lifters/search?q=" + q)
		p := model.ParsePagination(r)
		toLbs := model.ParseUnit(r)

		var matches []model.Lifter
		for name, lifter := range db.LifterHistory {
			if strings.Contains(strings.ToLower(name), q) {
				if toLbs {
					matches = append(matches, lifter.ToLbs())
				} else {
					matches = append(matches, *lifter)
				}
			}
		}

		// sort alphabetically for stable results
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Name < matches[j].Name
		})

		total := len(matches)
		start := p.Offset
		if start > total {
			start = total
		}
		end := start + p.Limit
		if end > total {
			end = total
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.PaginatedResponse{
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
			Data:   matches[start:end],
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetLifter(ds *data.DataStore) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
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

		var result interface{}
		if model.ParseUnit(r) {
			converted := lifter.ToLbs()
			result = &converted
		} else {
			result = lifter
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type topLifterEntry struct {
	Name      string             `json:"name"`
	Equipment string             `json:"equipment"`
	PB        *model.PersonalBest `json:"pb"`
}

func GetTopLifters(ds *data.DataStore) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		fmt.Println("GET /lifters/top")
		p := model.ParsePagination(r)
		toLbs := model.ParseUnit(r)

		// optional filters
		sexFilter := r.URL.Query().Get("sex")
		equipFilter := r.URL.Query().Get("equipment")
		wcFilter := r.URL.Query().Get("weightClass")

		var entries []topLifterEntry

		for _, lifter := range db.LifterHistory {
			// check each equipment PB for this lifter
			for equip, pb := range lifter.PB {
				if equipFilter != "" && !strings.EqualFold(equip, equipFilter) {
					continue
				}
				if pb.Dots == 0 {
					continue
				}

				// need to check sex and weightClass from competition results
				if sexFilter != "" || wcFilter != "" {
					matched := false
					for _, cr := range lifter.CompetitionResults {
						if !strings.EqualFold(cr.Equipment, equip) {
							continue
						}
						if sexFilter != "" && !strings.EqualFold(cr.Sex, sexFilter) {
							continue
						}
						if wcFilter != "" && cr.WeightClassKg != wcFilter {
							continue
						}
						matched = true
						break
					}
					if !matched {
						continue
					}
				}

				entryPB := pb
				if toLbs {
					entryPB = pb.ToLbs()
				}
				entries = append(entries, topLifterEntry{
					Name:      lifter.Name,
					Equipment: equip,
					PB:        entryPB,
				})
			}
		}

		// sort by DOTS descending
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].PB.Dots > entries[j].PB.Dots
		})

		total := len(entries)
		start := p.Offset
		if start > total {
			start = total
		}
		end := start + p.Limit
		if end > total {
			end = total
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.PaginatedResponse{
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
			Data:   entries[start:end],
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetRecords(ds *data.DataStore) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		fmt.Println("GET /records")
		toLbs := model.ParseUnit(r)

		// optional filters
		sexFilter := r.URL.Query().Get("sex")
		equipFilter := r.URL.Query().Get("equipment")
		wcFilter := r.URL.Query().Get("weightClass")

		// key: "sex|equipment|weightClass"
		type recordKey struct {
			Sex, Equipment, WeightClass string
		}
		records := make(map[recordKey]*model.Record)

		for _, lifter := range db.LifterHistory {
			for _, cr := range lifter.CompetitionResults {
				if sexFilter != "" && !strings.EqualFold(cr.Sex, sexFilter) {
					continue
				}
				if equipFilter != "" && !strings.EqualFold(cr.Equipment, equipFilter) {
					continue
				}
				if wcFilter != "" && cr.WeightClassKg != wcFilter {
					continue
				}
				if cr.Sex == "" || cr.Equipment == "" || cr.WeightClassKg == "" {
					continue
				}

				rk := recordKey{cr.Sex, cr.Equipment, cr.WeightClassKg}
				rec, exists := records[rk]
				if !exists {
					rec = &model.Record{
						WeightClassKg: cr.WeightClassKg,
						Sex:           cr.Sex,
						Equipment:     cr.Equipment,
					}
					records[rk] = rec
				}

				// check squat
				if cr.Squat != nil && cr.Squat.Best > 0 {
					if rec.Squat == nil || cr.Squat.Best > rec.Squat.Lift {
						rec.Squat = &model.RecordHolder{Lift: cr.Squat.Best, Lifter: cr.Name}
					}
				}
				// check bench
				if cr.Bench != nil && cr.Bench.Best > 0 {
					if rec.Bench == nil || cr.Bench.Best > rec.Bench.Lift {
						rec.Bench = &model.RecordHolder{Lift: cr.Bench.Best, Lifter: cr.Name}
					}
				}
				// check deadlift
				if cr.Deadlift != nil && cr.Deadlift.Best > 0 {
					if rec.Deadlift == nil || cr.Deadlift.Best > rec.Deadlift.Lift {
						rec.Deadlift = &model.RecordHolder{Lift: cr.Deadlift.Best, Lifter: cr.Name}
					}
				}
				// check total
				if cr.TotalKg > 0 {
					if rec.Total == nil || cr.TotalKg > rec.Total.Lift {
						rec.Total = &model.RecordHolder{Lift: cr.TotalKg, Lifter: cr.Name}
					}
				}
			}
		}

		// convert map to sorted slice
		result := make([]model.Record, 0, len(records))
		for _, rec := range records {
			if toLbs {
				converted := *rec
				if converted.Squat != nil {
					converted.Squat = &model.RecordHolder{Lift: model.ConvertToLbs(rec.Squat.Lift), Lifter: rec.Squat.Lifter}
				}
				if converted.Bench != nil {
					converted.Bench = &model.RecordHolder{Lift: model.ConvertToLbs(rec.Bench.Lift), Lifter: rec.Bench.Lifter}
				}
				if converted.Deadlift != nil {
					converted.Deadlift = &model.RecordHolder{Lift: model.ConvertToLbs(rec.Deadlift.Lift), Lifter: rec.Deadlift.Lifter}
				}
				if converted.Total != nil {
					converted.Total = &model.RecordHolder{Lift: model.ConvertToLbs(rec.Total.Lift), Lifter: rec.Total.Lifter}
				}
				result = append(result, converted)
			} else {
				result = append(result, *rec)
			}
		}

		// sort by sex, equipment, weightClass for stable output
		sort.Slice(result, func(i, j int) bool {
			if result[i].Sex != result[j].Sex {
				return result[i].Sex < result[j].Sex
			}
			if result[i].Equipment != result[j].Equipment {
				return result[i].Equipment < result[j].Equipment
			}
			return result[i].WeightClassKg < result[j].WeightClassKg
		})

		p := model.ParsePagination(r)
		total := len(result)
		start := p.Offset
		if start > total {
			start = total
		}
		end := start + p.Limit
		if end > total {
			end = total
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.PaginatedResponse{
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
			Data:   result[start:end],
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetMeets(ds *data.DataStore) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		fmt.Println("GET /meets")
		p := model.ParsePagination(r)

		meets := make([]model.Meet, 0)
		for fed := range db.FederationMeets {
			for meet := range db.FederationMeets[fed] {
				meets = append(meets, *db.FederationMeets[fed][meet])
			}
		}

		total := len(meets)
		start := p.Offset
		if start > total {
			start = total
		}
		end := start + p.Limit
		if end > total {
			end = total
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.PaginatedResponse{
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
			Data:   meets[start:end],
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetMeet(ds *data.DataStore) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		federationNameEncoded := chi.URLParam(r, "federationName")
		federationName, err := url.QueryUnescape(federationNameEncoded)

		if err != nil {
			http.Error(w, "invalid federation name", http.StatusBadRequest)
			return
		}

		federationMeets, exists := db.FederationMeets[federationName]

		if !exists {
			http.Error(w, "federation not found", http.StatusNotFound)
			return
		} else {
			fmt.Println("GET /meets/" + federationName)
		}

		p := model.ParsePagination(r)
		total := len(federationMeets)
		start := p.Offset
		if start > total {
			start = total
		}
		end := start + p.Limit
		if end > total {
			end = total
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.PaginatedResponse{
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
			Data:   federationMeets[start:end],
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetMeetResults(ds *data.DataStore) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		federationNameEncoded := chi.URLParam(r, "federationName")
		federationName, err := url.QueryUnescape(federationNameEncoded)
		if err != nil {
			http.Error(w, "invalid federation name", http.StatusBadRequest)
			return
		}

		meetNameEncoded := chi.URLParam(r, "meetName")
		meetName, err := url.QueryUnescape(meetNameEncoded)
		if err != nil {
			http.Error(w, "invalid meet name", http.StatusBadRequest)
			return
		}

		fmt.Println("GET /meets/" + federationName + "/" + meetName + "/results")

		key := data.MeetKey(federationName, meetName)
		results, exists := db.MeetResults[key]
		if !exists {
			http.Error(w, "meet not found", http.StatusNotFound)
			return
		}

		p := model.ParsePagination(r)
		toLbs := model.ParseUnit(r)

		total := len(results)
		start := p.Offset
		if start > total {
			start = total
		}
		end := start + p.Limit
		if end > total {
			end = total
		}

		page := results[start:end]
		if toLbs {
			converted := make([]*model.LifterMeetResult, len(page))
			for i, r := range page {
				converted[i] = r.ToLbs()
			}
			page = converted
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.PaginatedResponse{
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
			Data:   page,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func GetFederations(ds *data.DataStore) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		fmt.Println("GET /federations")
		p := model.ParsePagination(r)

		federations := make([]string, 0)
		for fed := range db.FederationMeets {
			federations = append(federations, fed)
		}

		total := len(federations)
		start := p.Offset
		if start > total {
			start = total
		}
		end := start + p.Limit
		if end > total {
			end = total
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.PaginatedResponse{
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
			Data:   federations[start:end],
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
