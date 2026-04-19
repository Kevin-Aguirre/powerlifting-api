package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Kevin-Aguirre/powerlifting-api/data"
	"github.com/Kevin-Aguirre/powerlifting-api/model"
)

func GetLifters(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		p := model.ParsePagination(r)
		toLbs := model.ParseUnit(r)
		sp := model.ParseSort(r, []string{"name"}, "name")

		sexFilter := r.URL.Query().Get("sex")
		equipFilter := r.URL.Query().Get("equipment")

		lifters := make([]model.Lifter, 0, len(db.LifterHistory))
		for _, lifter := range db.LifterHistory {
			if sexFilter != "" {
				matched := false
				for _, cr := range lifter.CompetitionResults {
					if strings.EqualFold(cr.Sex, sexFilter) {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
			if equipFilter != "" {
				found := false
				for k := range lifter.PB {
					if strings.EqualFold(k, equipFilter) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			if toLbs {
				lifters = append(lifters, lifter.ToLbs())
			} else {
				lifters = append(lifters, *lifter)
			}
		}

		sort.Slice(lifters, func(i, j int) bool {
			if sp.Order == "desc" {
				return lifters[i].Name > lifters[j].Name
			}
			return lifters[i].Name < lifters[j].Name
		})

		total := len(lifters)
		start, end := paginate(total, p)
		writeJSON(w, model.PaginatedResponse{Total: total, Limit: p.Limit, Offset: p.Offset, Data: lifters[start:end]})
	}
}

func GetLifterNames(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		p := model.ParsePagination(r)
		sp := model.ParseSort(r, []string{"name"}, "name")

		sexFilter := r.URL.Query().Get("sex")
		equipFilter := r.URL.Query().Get("equipment")

		names := make([]string, 0, len(db.LifterHistory))
		for name, lifter := range db.LifterHistory {
			if sexFilter != "" {
				matched := false
				for _, cr := range lifter.CompetitionResults {
					if strings.EqualFold(cr.Sex, sexFilter) {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
			if equipFilter != "" {
				found := false
				for k := range lifter.PB {
					if strings.EqualFold(k, equipFilter) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			names = append(names, name)
		}

		sort.Slice(names, func(i, j int) bool {
			if sp.Order == "desc" {
				return names[i] > names[j]
			}
			return names[i] < names[j]
		})

		total := len(names)
		start, end := paginate(total, p)
		writeJSON(w, model.PaginatedResponse{Total: total, Limit: p.Limit, Offset: p.Offset, Data: names[start:end]})
	}
}

func SearchLifters(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		q := strings.ToLower(r.URL.Query().Get("q"))
		if q == "" {
			http.Error(w, "missing required query parameter: q", http.StatusBadRequest)
			return
		}

		p := model.ParsePagination(r)
		toLbs := model.ParseUnit(r)
		sp := model.ParseSort(r, []string{"name"}, "name")

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

		sort.Slice(matches, func(i, j int) bool {
			if sp.Order == "desc" {
				return matches[i].Name > matches[j].Name
			}
			return matches[i].Name < matches[j].Name
		})

		total := len(matches)
		start, end := paginate(total, p)
		writeJSON(w, model.PaginatedResponse{Total: total, Limit: p.Limit, Offset: p.Offset, Data: matches[start:end]})
	}
}

func GetLifter(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		lifterName, err := url.QueryUnescape(chi.URLParam(r, "lifterName"))
		if err != nil {
			http.Error(w, "invalid lifter name", http.StatusBadRequest)
			return
		}

		lifter, exists := db.LifterHistory[lifterName]
		if !exists {
			http.Error(w, "lifter not found", http.StatusNotFound)
			return
		}

		var result interface{}
		if model.ParseUnit(r) {
			converted := lifter.ToLbs()
			result = &converted
		} else {
			result = lifter
		}
		writeJSON(w, result)
	}
}

func GetLifterStats(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		lifterName, err := url.QueryUnescape(chi.URLParam(r, "lifterName"))
		if err != nil {
			http.Error(w, "invalid lifter name", http.StatusBadRequest)
			return
		}

		lifter, exists := db.LifterHistory[lifterName]
		if !exists {
			http.Error(w, "lifter not found", http.StatusNotFound)
			return
		}

		writeJSON(w, computeCareerStats(lifter, model.ParseUnit(r)))
	}
}

func computeCareerStats(lifter *model.Lifter, toLbs bool) model.CareerStats {
	stats := model.CareerStats{
		Name:          lifter.Name,
		PRProgression: make(map[string][]model.PREntry),
	}

	sorted := make([]*model.LifterMeetResult, len(lifter.CompetitionResults))
	copy(sorted, lifter.CompetitionResults)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].MeetDate < sorted[j].MeetDate
	})

	stats.TotalCompetitions = len(sorted)

	fedSet := make(map[string]bool)
	for _, cr := range sorted {
		if cr.MeetFederation != "" {
			fedSet[cr.MeetFederation] = true
		}
		if cr.MeetDate != "" {
			if stats.FirstCompetition == "" {
				stats.FirstCompetition = cr.MeetDate
			}
			stats.LastCompetition = cr.MeetDate
		}
	}
	for fed := range fedSet {
		stats.Federations = append(stats.Federations, fed)
	}
	sort.Strings(stats.Federations)

	type runningBest struct{ total float64 }
	best := make(map[string]*runningBest)

	for _, cr := range sorted {
		equip := cr.Equipment
		if equip == "" {
			continue
		}
		if cr.TotalKg == 0 {
			continue
		}

		prev, hasPrev := best[equip]
		if hasPrev && cr.TotalKg <= prev.total {
			continue
		}

		squat, bench, deadlift := float64(0), float64(0), float64(0)
		if cr.Squat != nil {
			squat = cr.Squat.Best
		}
		if cr.Bench != nil {
			bench = cr.Bench.Best
		}
		if cr.Deadlift != nil {
			deadlift = cr.Deadlift.Best
		}

		dots := data.CalculateDots(cr.BodyweightKg, cr.TotalKg, cr.Sex)

		entry := model.PREntry{
			Date:     cr.MeetDate,
			Squat:    squat,
			Bench:    bench,
			Deadlift: deadlift,
			Total:    cr.TotalKg,
			Dots:     dots,
		}
		if toLbs {
			entry.Squat = model.ConvertToLbs(entry.Squat)
			entry.Bench = model.ConvertToLbs(entry.Bench)
			entry.Deadlift = model.ConvertToLbs(entry.Deadlift)
			entry.Total = model.ConvertToLbs(entry.Total)
		}

		stats.PRProgression[equip] = append(stats.PRProgression[equip], entry)
		best[equip] = &runningBest{cr.TotalKg}
	}

	return stats
}

func CompareLifters(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		toLbs := model.ParseUnit(r)

		nameA, err := url.QueryUnescape(r.URL.Query().Get("a"))
		if err != nil || nameA == "" {
			http.Error(w, "missing required query parameter: a", http.StatusBadRequest)
			return
		}
		nameB, err := url.QueryUnescape(r.URL.Query().Get("b"))
		if err != nil || nameB == "" {
			http.Error(w, "missing required query parameter: b", http.StatusBadRequest)
			return
		}

		lifterA, okA := db.LifterHistory[nameA]
		lifterB, okB := db.LifterHistory[nameB]

		if !okA {
			http.Error(w, "lifter not found: "+nameA, http.StatusNotFound)
			return
		}
		if !okB {
			http.Error(w, "lifter not found: "+nameB, http.StatusNotFound)
			return
		}

		type entry struct {
			Lifter *model.Lifter  `json:"lifter"`
			Stats  model.CareerStats `json:"stats"`
		}

		toEntry := func(l *model.Lifter) entry {
			var lf model.Lifter
			if toLbs {
				converted := l.ToLbs()
				lf = converted
			} else {
				lf = *l
			}
			return entry{Lifter: &lf, Stats: computeCareerStats(l, toLbs)}
		}

		writeJSON(w, map[string]interface{}{
			"a": toEntry(lifterA),
			"b": toEntry(lifterB),
		})
	}
}

type topLifterEntry struct {
	Name      string              `json:"name"`
	Equipment string              `json:"equipment"`
	PB        *model.PersonalBest `json:"pb"`
}

func GetTopLifters(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		p := model.ParsePagination(r)
		toLbs := model.ParseUnit(r)

		sexFilter := r.URL.Query().Get("sex")
		equipFilter := r.URL.Query().Get("equipment")
		wcFilter := r.URL.Query().Get("weightClass")

		// TopLifters is pre-sorted by DOTS desc — just filter and paginate.
		var entries []topLifterEntry
		for _, e := range db.TopLifters {
			if equipFilter != "" && !strings.EqualFold(e.Equipment, equipFilter) {
				continue
			}
			if sexFilter != "" && !strings.EqualFold(e.Sex, sexFilter) {
				continue
			}
			if wcFilter != "" && e.WeightClassKg != wcFilter {
				continue
			}
			pb := e.PB
			if toLbs {
				pb = pb.ToLbs()
			}
			entries = append(entries, topLifterEntry{Name: e.Name, Equipment: e.Equipment, PB: pb})
		}

		total := len(entries)
		start, end := paginate(total, p)
		writeJSON(w, model.PaginatedResponse{Total: total, Limit: p.Limit, Offset: p.Offset, Data: entries[start:end]})
	}
}

func GetRecords(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		toLbs := model.ParseUnit(r)

		sexFilter := r.URL.Query().Get("sex")
		equipFilter := r.URL.Query().Get("equipment")
		wcFilter := r.URL.Query().Get("weightClass")

		// Records is pre-sorted — just filter, optionally convert units, and paginate.
		result := make([]model.Record, 0)
		for _, rec := range db.Records {
			if sexFilter != "" && !strings.EqualFold(rec.Sex, sexFilter) {
				continue
			}
			if equipFilter != "" && !strings.EqualFold(rec.Equipment, equipFilter) {
				continue
			}
			if wcFilter != "" && rec.WeightClassKg != wcFilter {
				continue
			}
			if toLbs {
				converted := rec
				if rec.Squat != nil {
					converted.Squat = &model.RecordHolder{Lift: model.ConvertToLbs(rec.Squat.Lift), Lifter: rec.Squat.Lifter}
				}
				if rec.Bench != nil {
					converted.Bench = &model.RecordHolder{Lift: model.ConvertToLbs(rec.Bench.Lift), Lifter: rec.Bench.Lifter}
				}
				if rec.Deadlift != nil {
					converted.Deadlift = &model.RecordHolder{Lift: model.ConvertToLbs(rec.Deadlift.Lift), Lifter: rec.Deadlift.Lifter}
				}
				if rec.Total != nil {
					converted.Total = &model.RecordHolder{Lift: model.ConvertToLbs(rec.Total.Lift), Lifter: rec.Total.Lifter}
				}
				result = append(result, converted)
			} else {
				result = append(result, rec)
			}
		}

		p := model.ParsePagination(r)
		total := len(result)
		start, end := paginate(total, p)
		writeJSON(w, model.PaginatedResponse{Total: total, Limit: p.Limit, Offset: p.Offset, Data: result[start:end]})
	}
}

func GetMeets(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		p := model.ParsePagination(r)
		sp := model.ParseSort(r, []string{"date", "meetName", "federation", "country"}, "date")

		countryFilter := r.URL.Query().Get("country")
		fromFilter := r.URL.Query().Get("from")
		toFilter := r.URL.Query().Get("to")

		meets := make([]model.Meet, 0)
		for fed := range db.FederationMeets {
			for _, meet := range db.FederationMeets[fed] {
				if countryFilter != "" && !strings.EqualFold(meet.MeetCountry, countryFilter) {
					continue
				}
				if fromFilter != "" && meet.Date < fromFilter {
					continue
				}
				if toFilter != "" && meet.Date > toFilter {
					continue
				}
				meets = append(meets, *meet)
			}
		}

		sort.Slice(meets, func(i, j int) bool {
			var less bool
			switch sp.Field {
			case "meetName":
				less = meets[i].MeetName < meets[j].MeetName
			case "federation":
				less = meets[i].Federation < meets[j].Federation
			case "country":
				less = meets[i].MeetCountry < meets[j].MeetCountry
			default:
				less = meets[i].Date < meets[j].Date
			}
			if sp.Order == "desc" {
				return !less
			}
			return less
		})

		total := len(meets)
		start, end := paginate(total, p)
		writeJSON(w, model.PaginatedResponse{Total: total, Limit: p.Limit, Offset: p.Offset, Data: meets[start:end]})
	}
}

func GetMeet(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		federationName, err := url.QueryUnescape(chi.URLParam(r, "federationName"))
		if err != nil {
			http.Error(w, "invalid federation name", http.StatusBadRequest)
			return
		}

		federationMeets, exists := db.FederationMeets[federationName]
		if !exists {
			http.Error(w, "federation not found", http.StatusNotFound)
			return
		}

		p := model.ParsePagination(r)
		sp := model.ParseSort(r, []string{"date", "meetName", "country"}, "date")

		fromFilter := r.URL.Query().Get("from")
		toFilter := r.URL.Query().Get("to")
		countryFilter := r.URL.Query().Get("country")

		meets := make([]*model.Meet, 0, len(federationMeets))
		for _, meet := range federationMeets {
			if fromFilter != "" && meet.Date < fromFilter {
				continue
			}
			if toFilter != "" && meet.Date > toFilter {
				continue
			}
			if countryFilter != "" && !strings.EqualFold(meet.MeetCountry, countryFilter) {
				continue
			}
			meets = append(meets, meet)
		}

		sort.Slice(meets, func(i, j int) bool {
			var less bool
			switch sp.Field {
			case "meetName":
				less = meets[i].MeetName < meets[j].MeetName
			case "country":
				less = meets[i].MeetCountry < meets[j].MeetCountry
			default:
				less = meets[i].Date < meets[j].Date
			}
			if sp.Order == "desc" {
				return !less
			}
			return less
		})

		total := len(meets)
		start, end := paginate(total, p)
		writeJSON(w, model.PaginatedResponse{Total: total, Limit: p.Limit, Offset: p.Offset, Data: meets[start:end]})
	}
}

func GetMeetResults(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		federationName, err := url.QueryUnescape(chi.URLParam(r, "federationName"))
		if err != nil {
			http.Error(w, "invalid federation name", http.StatusBadRequest)
			return
		}
		meetName, err := url.QueryUnescape(chi.URLParam(r, "meetName"))
		if err != nil {
			http.Error(w, "invalid meet name", http.StatusBadRequest)
			return
		}

		key := data.MeetKey(federationName, meetName)
		results, exists := db.MeetResults[key]
		if !exists {
			http.Error(w, "meet not found", http.StatusNotFound)
			return
		}

		p := model.ParsePagination(r)
		toLbs := model.ParseUnit(r)

		total := len(results)
		start, end := paginate(total, p)
		page := results[start:end]

		if toLbs {
			converted := make([]*model.LifterMeetResult, len(page))
			for i, res := range page {
				converted[i] = res.ToLbs()
			}
			writeJSON(w, model.PaginatedResponse{Total: total, Limit: p.Limit, Offset: p.Offset, Data: converted})
			return
		}

		writeJSON(w, model.PaginatedResponse{Total: total, Limit: p.Limit, Offset: p.Offset, Data: page})
	}
}

func GetFederations(ds *data.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := ds.DB()
		p := model.ParsePagination(r)
		sp := model.ParseSort(r, []string{"name"}, "name")

		federations := make([]string, 0, len(db.FederationMeets))
		for fed := range db.FederationMeets {
			federations = append(federations, fed)
		}

		sort.Slice(federations, func(i, j int) bool {
			if sp.Order == "desc" {
				return federations[i] > federations[j]
			}
			return federations[i] < federations[j]
		})

		total := len(federations)
		start, end := paginate(total, p)
		writeJSON(w, model.PaginatedResponse{Total: total, Limit: p.Limit, Offset: p.Offset, Data: federations[start:end]})
	}
}

func paginate(total int, p model.PaginationParams) (start, end int) {
	start = p.Offset
	if start > total {
		start = total
	}
	end = start + p.Limit
	if end > total {
		end = total
	}
	return
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
