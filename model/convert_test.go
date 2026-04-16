package model

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoundTo2(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{1.006, 1.01},
		{1.004, 1.0},
		{0, 0},
		{100.999, 101.0},
		{-1.555, -1.56},
	}
	for _, tt := range tests {
		got := roundTo2(tt.input)
		if got != tt.expected {
			t.Errorf("roundTo2(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestConvertWeight(t *testing.T) {
	if got := convertWeight(0); got != 0 {
		t.Errorf("convertWeight(0) = %v, want 0", got)
	}
	// 100 kg = 220.46 lbs
	got := convertWeight(100)
	if got != 220.46 {
		t.Errorf("convertWeight(100) = %v, want 220.46", got)
	}
}

func TestConvertToLbs(t *testing.T) {
	if got := ConvertToLbs(0); got != 0 {
		t.Errorf("ConvertToLbs(0) = %v, want 0", got)
	}
	got := ConvertToLbs(100)
	if got != 220.46 {
		t.Errorf("ConvertToLbs(100) = %v, want 220.46", got)
	}
}

func TestParseUnit(t *testing.T) {
	tests := []struct {
		query    string
		expected bool
	}{
		{"unit=lbs", true},
		{"unit=kg", false},
		{"", false},
		{"unit=LBS", false}, // case-sensitive
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
		got := ParseUnit(r)
		if got != tt.expected {
			t.Errorf("ParseUnit(%q) = %v, want %v", tt.query, got, tt.expected)
		}
	}
}

func TestLiftAttempts_ToLbs(t *testing.T) {
	// nil case
	var la *LiftAttempts
	if got := la.ToLbs(); got != nil {
		t.Error("nil LiftAttempts.ToLbs() should return nil")
	}

	la = &LiftAttempts{
		Attempt1: 100,
		Attempt2: 110,
		Attempt3: 120,
		Attempt4: 0,
		Best:     120,
	}
	got := la.ToLbs()
	if got.Attempt1 != 220.46 {
		t.Errorf("Attempt1 = %v, want 220.46", got.Attempt1)
	}
	if got.Attempt4 != 0 {
		t.Errorf("Attempt4 = %v, want 0 (zero stays zero)", got.Attempt4)
	}
	if got.Best != 264.55 {
		t.Errorf("Best = %v, want 264.55", got.Best)
	}
}

func TestPersonalBest_ToLbs(t *testing.T) {
	var pb *PersonalBest
	if got := pb.ToLbs(); got != nil {
		t.Error("nil PersonalBest.ToLbs() should return nil")
	}

	pb = &PersonalBest{
		Squat:    200,
		Bench:    150,
		Deadlift: 250,
		Total:    600,
		Dots:     450.5,
	}
	got := pb.ToLbs()
	if got.Squat != 440.92 {
		t.Errorf("Squat = %v, want 440.92", got.Squat)
	}
	if got.Dots != 450.5 {
		t.Errorf("Dots should remain unchanged, got %v, want 450.5", got.Dots)
	}
}

func TestLifterMeetResult_ToLbs(t *testing.T) {
	var r *LifterMeetResult
	if got := r.ToLbs(); got != nil {
		t.Error("nil LifterMeetResult.ToLbs() should return nil")
	}

	r = &LifterMeetResult{
		Name:         "Test",
		BodyweightKg: 83,
		TotalKg:      600,
		Squat:        &LiftAttempts{Best: 200},
		Bench:        nil,
		Deadlift:     &LiftAttempts{Best: 250},
	}
	got := r.ToLbs()
	if got.Name != "Test" {
		t.Errorf("Name = %v, want Test", got.Name)
	}
	if got.BodyweightKg != convertWeight(83) {
		t.Errorf("BodyweightKg = %v, want %v", got.BodyweightKg, convertWeight(83))
	}
	if got.Bench != nil {
		t.Error("nil Bench should remain nil after conversion")
	}
	if got.Squat.Best != convertWeight(200) {
		t.Errorf("Squat.Best = %v, want %v", got.Squat.Best, convertWeight(200))
	}
}

func TestLifter_ToLbs(t *testing.T) {
	l := &Lifter{
		Name: "Test Lifter",
		PB: map[string]*PersonalBest{
			"Raw": {Squat: 200, Bench: 150, Deadlift: 250, Total: 600, Dots: 400},
		},
		CompetitionResults: []*LifterMeetResult{
			{Name: "Test Lifter", TotalKg: 600},
		},
	}
	got := l.ToLbs()
	if got.Name != "Test Lifter" {
		t.Errorf("Name = %v, want Test Lifter", got.Name)
	}
	if got.PB["Raw"].Dots != 400 {
		t.Errorf("PB Dots should remain unchanged")
	}
	if got.PB["Raw"].Squat != convertWeight(200) {
		t.Errorf("PB Squat = %v, want %v", got.PB["Raw"].Squat, convertWeight(200))
	}
	if len(got.CompetitionResults) != 1 {
		t.Fatalf("CompetitionResults len = %v, want 1", len(got.CompetitionResults))
	}
	if got.CompetitionResults[0].TotalKg != convertWeight(600) {
		t.Errorf("CompetitionResults[0].TotalKg = %v, want %v", got.CompetitionResults[0].TotalKg, convertWeight(600))
	}

	// nil PB
	l2 := &Lifter{Name: "No PB", CompetitionResults: []*LifterMeetResult{}}
	got2 := l2.ToLbs()
	if got2.PB != nil {
		t.Error("nil PB map should remain nil")
	}
}
