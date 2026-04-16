package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kevin-Aguirre/powerlifting-api/data"
	"github.com/Kevin-Aguirre/powerlifting-api/model"
	"github.com/go-chi/chi/v5"
)

// helper to build a DataStore with test data
func testDataStore() *data.DataStore {
	ds := data.NewDataStore()
	db := &data.Database{
		LifterHistory: map[string]*model.Lifter{
			"John Doe": {
				Name: "John Doe",
				PB: map[string]*model.PersonalBest{
					"Raw": {Squat: 220, Bench: 155, Deadlift: 270, Total: 645, Dots: 450},
				},
				CompetitionResults: []*model.LifterMeetResult{
					{
						Name:          "John Doe",
						Sex:           "M",
						Equipment:     "Raw",
						BodyweightKg:  83,
						WeightClassKg: "83",
						TotalKg:       645,
						Squat:         &model.LiftAttempts{Best: 220},
						Bench:         &model.LiftAttempts{Best: 155},
						Deadlift:      &model.LiftAttempts{Best: 270},
						Event:         "SBD",
					},
				},
			},
			"Jane Smith": {
				Name: "Jane Smith",
				PB: map[string]*model.PersonalBest{
					"Raw": {Squat: 135, Bench: 90, Deadlift: 170, Total: 395, Dots: 420},
				},
				CompetitionResults: []*model.LifterMeetResult{
					{
						Name:          "Jane Smith",
						Sex:           "F",
						Equipment:     "Raw",
						BodyweightKg:  63,
						WeightClassKg: "63",
						TotalKg:       395,
						Squat:         &model.LiftAttempts{Best: 135},
						Bench:         &model.LiftAttempts{Best: 90},
						Deadlift:      &model.LiftAttempts{Best: 170},
						Event:         "SBD",
					},
				},
			},
		},
		FederationMeets: map[string][]*model.Meet{
			"USAPL": {
				{Federation: "USAPL", MeetName: "Test Meet", Date: "2024-01-15", MeetCountry: "USA"},
			},
		},
		MeetResults: map[string][]*model.LifterMeetResult{
			"USAPL|Test Meet": {
				{Name: "John Doe", TotalKg: 645},
				{Name: "Jane Smith", TotalKg: 395},
			},
		},
	}
	ds.SetForTest(db)
	return ds
}

// helper to create a chi router context with URL params
func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		rctx = chi.NewRouteContext()
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	}
	rctx.URLParams.Add(key, value)
	return r
}

type paginatedJSON struct {
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	Data   json.RawMessage `json:"data"`
}

func TestGetLifters(t *testing.T) {
	ds := testDataStore()
	handler := GetLifters(ds)

	r := httptest.NewRequest(http.MethodGet, "/lifters", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp paginatedJSON
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
}

func TestGetLifters_Lbs(t *testing.T) {
	ds := testDataStore()
	handler := GetLifters(ds)

	r := httptest.NewRequest(http.MethodGet, "/lifters?unit=lbs", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestGetLifterNames(t *testing.T) {
	ds := testDataStore()
	handler := GetLifterNames(ds)

	r := httptest.NewRequest(http.MethodGet, "/lifters/names", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	var resp paginatedJSON
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
}

func TestSearchLifters(t *testing.T) {
	ds := testDataStore()
	handler := SearchLifters(ds)

	// missing q param
	r := httptest.NewRequest(http.MethodGet, "/lifters/search", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing q: status = %d, want 400", w.Code)
	}

	// valid search
	r = httptest.NewRequest(http.MethodGet, "/lifters/search?q=john", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp paginatedJSON
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 1 {
		t.Errorf("total = %d, want 1", resp.Total)
	}

	// no matches
	r = httptest.NewRequest(http.MethodGet, "/lifters/search?q=zzzzz", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 0 {
		t.Errorf("total = %d, want 0", resp.Total)
	}
}

func TestGetLifter(t *testing.T) {
	ds := testDataStore()
	handler := GetLifter(ds)

	// found
	r := httptest.NewRequest(http.MethodGet, "/lifters/John%20Doe", nil)
	r = withURLParam(r, "lifterName", "John%20Doe")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	// not found
	r = httptest.NewRequest(http.MethodGet, "/lifters/Nobody", nil)
	r = withURLParam(r, "lifterName", "Nobody")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestGetTopLifters(t *testing.T) {
	ds := testDataStore()
	handler := GetTopLifters(ds)

	// unfiltered
	r := httptest.NewRequest(http.MethodGet, "/lifters/top", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp paginatedJSON
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}

	// filter by sex
	r = httptest.NewRequest(http.MethodGet, "/lifters/top?sex=M", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 1 {
		t.Errorf("sex=M total = %d, want 1", resp.Total)
	}

	// filter by equipment
	r = httptest.NewRequest(http.MethodGet, "/lifters/top?equipment=Equipped", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 0 {
		t.Errorf("equipment=Equipped total = %d, want 0", resp.Total)
	}
}

func TestGetRecords(t *testing.T) {
	ds := testDataStore()
	handler := GetRecords(ds)

	r := httptest.NewRequest(http.MethodGet, "/records", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	// with filters
	r = httptest.NewRequest(http.MethodGet, "/records?sex=M&equipment=Raw&weightClass=83", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp paginatedJSON
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 1 {
		t.Errorf("filtered total = %d, want 1", resp.Total)
	}
}

func TestGetMeets(t *testing.T) {
	ds := testDataStore()
	handler := GetMeets(ds)

	r := httptest.NewRequest(http.MethodGet, "/meets", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp paginatedJSON
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 1 {
		t.Errorf("total = %d, want 1", resp.Total)
	}
}

func TestGetMeet(t *testing.T) {
	ds := testDataStore()
	handler := GetMeet(ds)

	// found
	r := httptest.NewRequest(http.MethodGet, "/meets/USAPL", nil)
	r = withURLParam(r, "federationName", "USAPL")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	// not found
	r = httptest.NewRequest(http.MethodGet, "/meets/FAKE", nil)
	r = withURLParam(r, "federationName", "FAKE")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestGetMeetResults(t *testing.T) {
	ds := testDataStore()
	handler := GetMeetResults(ds)

	// found
	r := httptest.NewRequest(http.MethodGet, "/meets/USAPL/Test%20Meet/results", nil)
	r = withURLParam(r, "federationName", "USAPL")
	r = withURLParam(r, "meetName", "Test%20Meet")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp paginatedJSON
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}

	// not found
	r = httptest.NewRequest(http.MethodGet, "/meets/USAPL/FakeMeet/results", nil)
	r = withURLParam(r, "federationName", "USAPL")
	r = withURLParam(r, "meetName", "FakeMeet")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestGetFederations(t *testing.T) {
	ds := testDataStore()
	handler := GetFederations(ds)

	r := httptest.NewRequest(http.MethodGet, "/federations", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp paginatedJSON
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 1 {
		t.Errorf("total = %d, want 1", resp.Total)
	}
}

func TestPagination_Integration(t *testing.T) {
	ds := testDataStore()
	handler := GetLifters(ds)

	// limit=1
	r := httptest.NewRequest(http.MethodGet, "/lifters?limit=1", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	var resp paginatedJSON
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
	if resp.Limit != 1 {
		t.Errorf("limit = %d, want 1", resp.Limit)
	}

	// offset beyond total
	r = httptest.NewRequest(http.MethodGet, "/lifters?offset=999", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
}
