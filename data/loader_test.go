package data

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestMeetKey(t *testing.T) {
	got := MeetKey("USAPL", "2024 Nationals")
	want := "USAPL|2024 Nationals"
	if got != want {
		t.Errorf("MeetKey() = %q, want %q", got, want)
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{" hello ", "hello"},
		{"\uFEFFName", "Name"},
		{"\uFEFF Name \uFEFF", "Name"},
	}
	for _, tt := range tests {
		got := clean(tt.input)
		if got != tt.expected {
			t.Errorf("clean(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFindIndex(t *testing.T) {
	header := []string{"Name", "Age", "Sex"}
	if got := findIndex(header, "Age"); got != 1 {
		t.Errorf("findIndex(Age) = %d, want 1", got)
	}
	if got := findIndex(header, "Missing"); got != -1 {
		t.Errorf("findIndex(Missing) = %d, want -1", got)
	}
	// BOM handling
	headerBOM := []string{"\uFEFFName", "Age"}
	if got := findIndex(headerBOM, "Name"); got != 0 {
		t.Errorf("findIndex with BOM = %d, want 0", got)
	}
}

func TestGetValue(t *testing.T) {
	row := []string{"John", "25", "M"}
	columnsMap := map[string]int{"Name": 0, "Age": 1, "Sex": 2, "Missing": -1}

	if got := getValue(row, columnsMap, "Name"); got != "John" {
		t.Errorf("getValue(Name) = %q, want John", got)
	}
	if got := getValue(row, columnsMap, "Missing"); got != "" {
		t.Errorf("getValue(Missing) = %q, want empty", got)
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"42", 42},
		{"", 0},
		{"abc", 0},
		{"-5", -5},
	}
	for _, tt := range tests {
		got := parseInt(tt.input)
		if got != tt.expected {
			t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"100.5", 100.5},
		{"", 0},
		{"abc", 0},
		{"-50.25", -50.25},
	}
	for _, tt := range tests {
		got := parseFloat(tt.input)
		if got != tt.expected {
			t.Errorf("parseFloat(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestBuildLiftAttempts(t *testing.T) {
	// all zeros → nil
	got := buildLiftAttempts("", "", "", "", "")
	if got != nil {
		t.Error("buildLiftAttempts with all empty should return nil")
	}

	// valid attempts
	got = buildLiftAttempts("100", "110", "120", "", "120")
	if got == nil {
		t.Fatal("buildLiftAttempts should not return nil for valid data")
	}
	if got.Attempt1 != 100 || got.Attempt3 != 120 || got.Best != 120 {
		t.Errorf("unexpected values: %+v", got)
	}
	if got.Attempt4 != 0 {
		t.Errorf("Attempt4 = %v, want 0", got.Attempt4)
	}
}

func TestCalculateDots(t *testing.T) {
	// Male calculation
	dots := CalculateDots(83, 600, "M")
	if dots <= 0 {
		t.Errorf("male DOTS should be positive, got %v", dots)
	}

	// Female calculation
	dots = CalculateDots(63, 400, "F")
	if dots <= 0 {
		t.Errorf("female DOTS should be positive, got %v", dots)
	}

	// Mx returns 0
	if dots := CalculateDots(83, 600, "Mx"); dots != 0 {
		t.Errorf("Mx DOTS = %v, want 0", dots)
	}

	// Unknown sex returns 0
	if dots := CalculateDots(83, 600, ""); dots != 0 {
		t.Errorf("empty sex DOTS = %v, want 0", dots)
	}

	// Zero bodyweight should not panic (denominator could be near-zero)
	dots = CalculateDots(0, 600, "M")
	if math.IsInf(dots, 0) || math.IsNaN(dots) {
		t.Errorf("zero bodyweight should not produce Inf/NaN, got %v", dots)
	}
}

func TestLoadDatabase(t *testing.T) {
	// Create temp directory structure simulating opl-data/meet-data
	tmpDir := t.TempDir()
	fedDir := filepath.Join(tmpDir, "usapl", "2401")
	if err := os.MkdirAll(fedDir, 0755); err != nil {
		t.Fatal(err)
	}

	meetCSV := `Federation,Date,MeetCountry,MeetState,MeetTown,MeetName,RuleSet
USAPL,2024-01-15,USA,CA,Los Angeles,Test Meet,USAPL`
	if err := os.WriteFile(filepath.Join(fedDir, "meet.csv"), []byte(meetCSV), 0644); err != nil {
		t.Fatal(err)
	}

	entriesCSV := `Place,Name,Sex,Equipment,Age,BodyweightKg,WeightClassKg,Squat1Kg,Squat2Kg,Squat3Kg,Best3SquatKg,Bench1Kg,Bench2Kg,Bench3Kg,Best3BenchKg,Deadlift1Kg,Deadlift2Kg,Deadlift3Kg,Best3DeadliftKg,TotalKg,Event,Tested
1,John Doe,M,Raw,25,82.5,83,200,210,220,220,140,150,155,155,240,260,270,270,645,SBD,Yes
2,Jane Smith,F,Raw,30,62.5,63,120,130,135,135,80,85,90,90,150,160,170,170,395,SBD,Yes`
	if err := os.WriteFile(filepath.Join(fedDir, "entries.csv"), []byte(entriesCSV), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := LoadDatabase(tmpDir)
	if err != nil {
		t.Fatalf("LoadDatabase error: %v", err)
	}

	// Check lifters loaded
	if len(db.LifterHistory) != 2 {
		t.Errorf("LifterHistory count = %d, want 2", len(db.LifterHistory))
	}

	john, ok := db.LifterHistory["John Doe"]
	if !ok {
		t.Fatal("John Doe not found in LifterHistory")
	}
	if len(john.CompetitionResults) != 1 {
		t.Errorf("John Doe results = %d, want 1", len(john.CompetitionResults))
	}
	if john.CompetitionResults[0].TotalKg != 645 {
		t.Errorf("John Doe total = %v, want 645", john.CompetitionResults[0].TotalKg)
	}
	if john.CompetitionResults[0].Squat == nil || john.CompetitionResults[0].Squat.Best != 220 {
		t.Error("John Doe squat best should be 220")
	}

	// Check PBs exist
	rawPB, ok := john.PB["Raw"]
	if !ok {
		t.Fatal("John Doe missing Raw PB")
	}
	if rawPB.Squat != 220 || rawPB.Bench != 155 || rawPB.Deadlift != 270 {
		t.Errorf("PB unexpected: %+v", rawPB)
	}
	if rawPB.Dots <= 0 {
		t.Error("John Doe DOTS should be positive")
	}

	// Check federation meets
	usaplMeets := db.FederationMeets["USAPL"]
	if len(usaplMeets) != 1 {
		t.Errorf("USAPL meets = %d, want 1", len(usaplMeets))
	}
	if usaplMeets[0].MeetName != "Test Meet" {
		t.Errorf("meet name = %q, want Test Meet", usaplMeets[0].MeetName)
	}

	// Check meet results indexed
	key := MeetKey("USAPL", "Test Meet")
	results := db.MeetResults[key]
	if len(results) != 2 {
		t.Errorf("meet results = %d, want 2", len(results))
	}
}

func TestLoadDatabase_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := LoadDatabase(tmpDir)
	if err != nil {
		t.Fatalf("LoadDatabase on empty dir should not error, got: %v", err)
	}
	if len(db.LifterHistory) != 0 {
		t.Errorf("empty dir should produce empty db, got %d lifters", len(db.LifterHistory))
	}
}

func TestLoadDatabase_MissingMeetCSV(t *testing.T) {
	// entries.csv without a sibling meet.csv — should still load lifters
	tmpDir := t.TempDir()
	fedDir := filepath.Join(tmpDir, "fed", "meet1")
	os.MkdirAll(fedDir, 0755)

	entriesCSV := `Place,Name,Sex,Equipment,BodyweightKg,WeightClassKg,TotalKg,Event
1,Solo Lifter,M,Raw,90,90,500,SBD`
	os.WriteFile(filepath.Join(fedDir, "entries.csv"), []byte(entriesCSV), 0644)

	db, err := LoadDatabase(tmpDir)
	if err != nil {
		t.Fatalf("LoadDatabase error: %v", err)
	}
	if _, ok := db.LifterHistory["Solo Lifter"]; !ok {
		t.Error("Solo Lifter should be loaded even without meet.csv")
	}
	// No meet results indexed since meet.csv is missing
	if len(db.MeetResults) != 0 {
		t.Errorf("MeetResults should be empty without meet.csv, got %d", len(db.MeetResults))
	}
}
