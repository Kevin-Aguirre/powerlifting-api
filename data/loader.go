package data

import (
	"encoding/csv"
	"io/fs"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Kevin-Aguirre/powerlifting-api/model"
)

const (
	dotsMaleA   = -307.75076
	dotsMaleB   = 24.0900756
	dotsMaleC   = -0.1918759221
	dotsMaleD   = 0.0007391293
	dotsMaleE   = -0.000001093
	dotsFemaleA = -57.96288
	dotsFemaleB = 13.6175032
	dotsFemaleC = -0.1126655495
	dotsFemaleD = 0.0005158568
	dotsFemaleE = -0.0000010706

	colHeaderPlace           = "Place"
	colHeaderName            = "Name"
	colHeaderBirthDate       = "BirthDate"
	colHeaderSex             = "Sex"
	colHeaderBirthYear       = "BirthYear"
	colHeaderAge             = "Age"
	colHeaderCountry         = "Country"
	colHeaderState           = "State"
	colHeaderEquipment       = "Equipment"
	colHeaderDivision        = "Division"
	colHeaderBodyweightKg    = "BodyweightKg"
	colHeaderWeightClassKg   = "WeightClassKg"
	colHeaderSquat1Kg        = "Squat1Kg"
	colHeaderSquat2Kg        = "Squat2Kg"
	colHeaderSquat3Kg        = "Squat3Kg"
	colHeaderBest3SquatKg    = "Best3SquatKg"
	colHeaderSquat4Kg        = "Squat4Kg"
	colHeaderBench1Kg        = "Bench1Kg"
	colHeaderBench2Kg        = "Bench2Kg"
	colHeaderBench3Kg        = "Bench3Kg"
	colHeaderBest3BenchKg    = "Best3BenchKg"
	colHeaderBench4Kg        = "Bench4Kg"
	colHeaderDeadlift1Kg     = "Deadlift1Kg"
	colHeaderDeadlift2Kg     = "Deadlift2Kg"
	colHeaderDeadlift3Kg     = "Deadlift3Kg"
	colHeaderBest3DeadliftKg = "Best3DeadliftKg"
	colHeaderDeadlift4Kg     = "Deadlift4Kg"
	colHeaderTotalKg         = "TotalKg"
	colHeaderEvent           = "Event"
	colHeaderTested          = "Tested"

	colHeaderFederation  = "Federation"
	colHeaderDate        = "Date"
	colHeaderMeetCountry = "MeetCountry"
	colHeaderMeetState   = "MeetState"
	colHeaderMeetTown    = "MeetTown"
	colHeaderMeetName    = "MeetName"
	colHeaderRuleSet     = "RuleSet"

	meetEntriesFileName = "entries.csv"
	meetInfoFileName    = "meet.csv"
)

// PrecomputedTopEntry is an indexed entry for the top-lifters leaderboard, sorted by DOTS desc at load time.
type PrecomputedTopEntry struct {
	Name          string
	Equipment     string
	Sex           string
	WeightClassKg string
	PB            *model.PersonalBest
}

type Database struct {
	LifterHistory   map[string]*model.Lifter
	FederationMeets map[string][]*model.Meet
	MeetResults     map[string][]*model.LifterMeetResult

	// Precomputed indexes — built once at load time, read-only during serving.
	TopLifters []*PrecomputedTopEntry // sorted by DOTS desc
	Records    []model.Record        // sorted by sex / equipment / weightClass
}

// MeetKey uniquely identifies a meet by federation and name.
func MeetKey(federation, meetName string) string {
	return federation + "|" + meetName
}

// parsedMeet is the result of parsing a single meet.csv without touching the database.
type parsedMeet struct {
	federation string
	meet       *model.Meet
}

// parsedEntriesFile is the result of parsing a single entries.csv without touching the database.
type parsedEntriesFile struct {
	key     string // MeetKey(federation, meetName), empty if meet info missing
	results []*model.LifterMeetResult
}

func clean(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\uFEFF", ""))
}

func findIndex(headerRow []string, columnName string) int {
	for i, value := range headerRow {
		if clean(value) == clean(columnName) {
			return i
		}
	}
	return -1
}

func getValue(row []string, columnsMap map[string]int, column string) string {
	idx := columnsMap[column]
	if idx == -1 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func computeEntriesColumnsMap(records [][]string) map[string]int {
	columnsMap := make(map[string]int)
	for _, column := range []string{
		colHeaderPlace, colHeaderName, colHeaderBirthDate, colHeaderSex,
		colHeaderBirthYear, colHeaderAge, colHeaderCountry, colHeaderState,
		colHeaderEquipment, colHeaderDivision, colHeaderBodyweightKg, colHeaderWeightClassKg,
		colHeaderSquat1Kg, colHeaderSquat2Kg, colHeaderSquat3Kg, colHeaderBest3SquatKg, colHeaderSquat4Kg,
		colHeaderBench1Kg, colHeaderBench2Kg, colHeaderBench3Kg, colHeaderBest3BenchKg, colHeaderBench4Kg,
		colHeaderDeadlift1Kg, colHeaderDeadlift2Kg, colHeaderDeadlift3Kg, colHeaderBest3DeadliftKg, colHeaderDeadlift4Kg,
		colHeaderTotalKg, colHeaderEvent, colHeaderTested,
	} {
		columnsMap[column] = findIndex(records[0], column)
	}
	return columnsMap
}

func computeMeetColumnsMap(records [][]string) map[string]int {
	columnsMap := make(map[string]int)
	for _, column := range []string{
		colHeaderFederation, colHeaderDate, colHeaderMeetCountry,
		colHeaderMeetState, colHeaderMeetTown, colHeaderMeetName, colHeaderRuleSet,
	} {
		columnsMap[column] = findIndex(records[0], column)
	}
	return columnsMap
}

func getFederationMeetInfo(row []string, columnsMap map[string]int) *model.Meet {
	return &model.Meet{
		Federation:  getValue(row, columnsMap, colHeaderFederation),
		Date:        getValue(row, columnsMap, colHeaderDate),
		MeetCountry: getValue(row, columnsMap, colHeaderMeetCountry),
		MeetState:   getValue(row, columnsMap, colHeaderMeetState),
		MeetTown:    getValue(row, columnsMap, colHeaderMeetTown),
		MeetName:    getValue(row, columnsMap, colHeaderMeetName),
		RuleSet:     getValue(row, columnsMap, colHeaderRuleSet),
	}
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) {
		return 0
	}
	return v
}

func buildLiftAttempts(a1, a2, a3, a4, best string) *model.LiftAttempts {
	la := &model.LiftAttempts{
		Attempt1: parseFloat(a1),
		Attempt2: parseFloat(a2),
		Attempt3: parseFloat(a3),
		Attempt4: parseFloat(a4),
		Best:     parseFloat(best),
	}
	if la.Attempt1 == 0 && la.Attempt2 == 0 && la.Attempt3 == 0 && la.Attempt4 == 0 && la.Best == 0 {
		return nil
	}
	return la
}

func getLifterMeetResult(row []string, columnsMap map[string]int) *model.LifterMeetResult {
	return &model.LifterMeetResult{
		Place:         getValue(row, columnsMap, colHeaderPlace),
		Name:          getValue(row, columnsMap, colHeaderName),
		BirthDate:     getValue(row, columnsMap, colHeaderBirthDate),
		Sex:           getValue(row, columnsMap, colHeaderSex),
		BirthYear:     parseInt(getValue(row, columnsMap, colHeaderBirthYear)),
		Age:           parseFloat(getValue(row, columnsMap, colHeaderAge)),
		Country:       getValue(row, columnsMap, colHeaderCountry),
		State:         getValue(row, columnsMap, colHeaderState),
		Equipment:     getValue(row, columnsMap, colHeaderEquipment),
		Division:      getValue(row, columnsMap, colHeaderDivision),
		BodyweightKg:  parseFloat(getValue(row, columnsMap, colHeaderBodyweightKg)),
		WeightClassKg: getValue(row, columnsMap, colHeaderWeightClassKg),
		Squat: buildLiftAttempts(
			getValue(row, columnsMap, colHeaderSquat1Kg),
			getValue(row, columnsMap, colHeaderSquat2Kg),
			getValue(row, columnsMap, colHeaderSquat3Kg),
			getValue(row, columnsMap, colHeaderSquat4Kg),
			getValue(row, columnsMap, colHeaderBest3SquatKg),
		),
		Bench: buildLiftAttempts(
			getValue(row, columnsMap, colHeaderBench1Kg),
			getValue(row, columnsMap, colHeaderBench2Kg),
			getValue(row, columnsMap, colHeaderBench3Kg),
			getValue(row, columnsMap, colHeaderBench4Kg),
			getValue(row, columnsMap, colHeaderBest3BenchKg),
		),
		Deadlift: buildLiftAttempts(
			getValue(row, columnsMap, colHeaderDeadlift1Kg),
			getValue(row, columnsMap, colHeaderDeadlift2Kg),
			getValue(row, columnsMap, colHeaderDeadlift3Kg),
			getValue(row, columnsMap, colHeaderDeadlift4Kg),
			getValue(row, columnsMap, colHeaderBest3DeadliftKg),
		),
		TotalKg: parseFloat(getValue(row, columnsMap, colHeaderTotalKg)),
		Event:   getValue(row, columnsMap, colHeaderEvent),
		Tested:  getValue(row, columnsMap, colHeaderTested),
	}
}

func CalculateDots(bodyweightKgs, totalKgs float64, gender string) float64 {
	var a, b, c, d, e float64
	switch gender {
	case "M":
		a, b, c, d, e = dotsMaleA, dotsMaleB, dotsMaleC, dotsMaleD, dotsMaleE
	case "F":
		a, b, c, d, e = dotsFemaleA, dotsFemaleB, dotsFemaleC, dotsFemaleD, dotsFemaleE
	default:
		return 0
	}
	x := bodyweightKgs
	result := (500 * totalKgs) / (a + b*x + c*x*x + d*x*x*x + e*x*x*x*x)
	if math.IsInf(result, 0) || math.IsNaN(result) {
		return 0
	}
	return result
}

func ensureLifterExists(db *Database, lifterName string) {
	if _, exists := db.LifterHistory[lifterName]; !exists {
		db.LifterHistory[lifterName] = &model.Lifter{
			Name:               lifterName,
			PB:                 make(map[string]*model.PersonalBest),
			CompetitionResults: []*model.LifterMeetResult{},
		}
	}
}

func ensureFederationExists(db *Database, federationName string) {
	if _, exists := db.FederationMeets[federationName]; !exists {
		db.FederationMeets[federationName] = make([]*model.Meet, 0)
	}
}

func handleCompetitionResultsUpdate(db *Database, lifterResult *model.LifterMeetResult, lifterName string) {
	db.LifterHistory[lifterName].CompetitionResults = append(
		db.LifterHistory[lifterName].CompetitionResults,
		lifterResult,
	)
}

func handleFederationMeetUpdate(db *Database, meetInfo *model.Meet, federationName string) {
	db.FederationMeets[federationName] = append(db.FederationMeets[federationName], meetInfo)
}

func handlePBUpdate(db *Database, lifterResult *model.LifterMeetResult, lifterName string) {
	squat := float64(0)
	if lifterResult.Squat != nil {
		squat = lifterResult.Squat.Best
	}
	bench := float64(0)
	if lifterResult.Bench != nil {
		bench = lifterResult.Bench.Best
	}
	deadlift := float64(0)
	if lifterResult.Deadlift != nil {
		deadlift = lifterResult.Deadlift.Best
	}

	total := squat + bench + deadlift
	var currDots float64
	if lifterResult.BodyweightKg > 0 {
		currDots = CalculateDots(lifterResult.BodyweightKg, total, lifterResult.Sex)
	}

	candidate := &model.PersonalBest{
		Squat:    squat,
		Bench:    bench,
		Deadlift: deadlift,
		Total:    total,
		Dots:     currDots,
	}

	existing, ok := db.LifterHistory[lifterName].PB[lifterResult.Equipment]
	if !ok || currDots > existing.Dots {
		db.LifterHistory[lifterName].PB[lifterResult.Equipment] = candidate
	}
}

// readMeetInfo reads the meet.csv sibling and returns federation, meetName, and date.
func readMeetInfo(entriesPath string) (federation, meetName, date string) {
	meetPath := filepath.Join(filepath.Dir(entriesPath), meetInfoFileName)
	file, err := os.Open(meetPath)
	if err != nil {
		return "", "", ""
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil || len(records) < 2 {
		return "", "", ""
	}

	columnsMap := computeMeetColumnsMap(records)
	row := records[1]
	return getValue(row, columnsMap, colHeaderFederation),
		getValue(row, columnsMap, colHeaderMeetName),
		getValue(row, columnsMap, colHeaderDate)
}

// parseMeetFileRaw parses a meet.csv and returns its data without touching any database.
func parseMeetFileRaw(path string) (parsedMeet, bool) {
	file, err := os.Open(path)
	if err != nil {
		slog.Error("error opening file", "path", path, "error", err)
		return parsedMeet{}, false
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil || len(records) < 2 {
		return parsedMeet{}, false
	}

	columnsMap := computeMeetColumnsMap(records)
	row := records[1]
	federation := getValue(row, columnsMap, colHeaderFederation)
	meet := getFederationMeetInfo(row, columnsMap)
	return parsedMeet{federation: federation, meet: meet}, federation != ""
}

// parseEntriesFileRaw parses an entries.csv and returns all results without touching any database.
func parseEntriesFileRaw(path string) (parsedEntriesFile, bool) {
	file, err := os.Open(path)
	if err != nil {
		slog.Error("error opening file", "path", path, "error", err)
		return parsedEntriesFile{}, false
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		slog.Error("could not read lines", "path", path, "error", err)
		return parsedEntriesFile{}, false
	}
	if len(records) < 1 {
		return parsedEntriesFile{}, false
	}

	columnsMap := computeEntriesColumnsMap(records)
	federation, meetName, meetDate := readMeetInfo(path)
	var key string
	if federation != "" && meetName != "" {
		key = MeetKey(federation, meetName)
	}

	results := make([]*model.LifterMeetResult, 0, len(records)-1)
	for _, row := range records[1:] {
		r := getLifterMeetResult(row, columnsMap)
		r.MeetFederation = federation
		r.MeetName = meetName
		r.MeetDate = meetDate
		results = append(results, r)
	}

	return parsedEntriesFile{key: key, results: results}, true
}

// parallelParse runs fn concurrently over paths using numWorkers goroutines.
// The parse phase (I/O) runs in parallel; results are returned for serial merging by the caller.
func parallelParse[T any](paths []string, numWorkers int, fn func(string) (T, bool)) []T {
	if len(paths) == 0 {
		return nil
	}

	ch := make(chan string, len(paths))
	for _, p := range paths {
		ch <- p
	}
	close(ch)

	type item struct {
		val T
		ok  bool
	}
	resultCh := make(chan item, len(paths))

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range ch {
				v, ok := fn(p)
				resultCh <- item{v, ok}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var out []T
	for it := range resultCh {
		if it.ok {
			out = append(out, it.val)
		}
	}
	return out
}

func LoadDatabase(root string) (*Database, error) {
	db := &Database{
		FederationMeets: make(map[string][]*model.Meet),
		LifterHistory:   make(map[string]*model.Lifter),
		MeetResults:     make(map[string][]*model.LifterMeetResult),
	}

	var entryPaths, meetPaths []string
	err := filepath.WalkDir(root, func(currPath string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("error accessing path", "path", currPath, "error", err)
			return err
		}
		switch filepath.Base(currPath) {
		case meetEntriesFileName:
			entryPaths = append(entryPaths, currPath)
		case meetInfoFileName:
			meetPaths = append(meetPaths, currPath)
		}
		return nil
	})
	if err != nil {
		slog.Error("error walking the directory", "root", root, "error", err)
	}

	numWorkers := runtime.NumCPU()

	// Parse meet files in parallel (I/O), merge serially (memory).
	for _, pm := range parallelParse(meetPaths, numWorkers, parseMeetFileRaw) {
		ensureFederationExists(db, pm.federation)
		handleFederationMeetUpdate(db, pm.meet, pm.federation)
	}

	// Parse entries files in parallel (I/O), merge serially (memory).
	for _, pe := range parallelParse(entryPaths, numWorkers, parseEntriesFileRaw) {
		for _, r := range pe.results {
			if r.Name == "" {
				continue
			}
			ensureLifterExists(db, r.Name)
			handleCompetitionResultsUpdate(db, r, r.Name)
			handlePBUpdate(db, r, r.Name)
			if pe.key != "" {
				db.MeetResults[pe.key] = append(db.MeetResults[pe.key], r)
			}
		}
	}

	db.BuildIndexes()
	return db, nil
}

// BuildIndexes computes the TopLifters and Records sorted slices from the loaded data.
// It is called automatically by LoadDatabase and should also be called when constructing
// a Database manually (e.g. in tests).
func (db *Database) BuildIndexes() {
	db.buildTopLifters()
	db.buildRecords()
}

func (db *Database) buildTopLifters() {
	db.TopLifters = db.TopLifters[:0]
	for _, lifter := range db.LifterHistory {
		for equip, pb := range lifter.PB {
			if pb.Dots == 0 {
				continue
			}
			var sex, wc string
			for _, cr := range lifter.CompetitionResults {
				if strings.EqualFold(cr.Equipment, equip) {
					sex = cr.Sex
					wc = cr.WeightClassKg
					break
				}
			}
			db.TopLifters = append(db.TopLifters, &PrecomputedTopEntry{
				Name:          lifter.Name,
				Equipment:     equip,
				Sex:           sex,
				WeightClassKg: wc,
				PB:            pb,
			})
		}
	}
	sort.Slice(db.TopLifters, func(i, j int) bool {
		return db.TopLifters[i].PB.Dots > db.TopLifters[j].PB.Dots
	})
}

func (db *Database) buildRecords() {
	type recordKey struct{ Sex, Equipment, WeightClass string }
	records := make(map[recordKey]*model.Record)

	for _, lifter := range db.LifterHistory {
		for _, cr := range lifter.CompetitionResults {
			if cr.Sex == "" || cr.Equipment == "" || cr.WeightClassKg == "" {
				continue
			}
			rk := recordKey{cr.Sex, cr.Equipment, cr.WeightClassKg}
			rec, exists := records[rk]
			if !exists {
				rec = &model.Record{WeightClassKg: cr.WeightClassKg, Sex: cr.Sex, Equipment: cr.Equipment}
				records[rk] = rec
			}
			if cr.Squat != nil && cr.Squat.Best > 0 {
				if rec.Squat == nil || cr.Squat.Best > rec.Squat.Lift {
					rec.Squat = &model.RecordHolder{Lift: cr.Squat.Best, Lifter: cr.Name}
				}
			}
			if cr.Bench != nil && cr.Bench.Best > 0 {
				if rec.Bench == nil || cr.Bench.Best > rec.Bench.Lift {
					rec.Bench = &model.RecordHolder{Lift: cr.Bench.Best, Lifter: cr.Name}
				}
			}
			if cr.Deadlift != nil && cr.Deadlift.Best > 0 {
				if rec.Deadlift == nil || cr.Deadlift.Best > rec.Deadlift.Lift {
					rec.Deadlift = &model.RecordHolder{Lift: cr.Deadlift.Best, Lifter: cr.Name}
				}
			}
			if cr.TotalKg > 0 {
				if rec.Total == nil || cr.TotalKg > rec.Total.Lift {
					rec.Total = &model.RecordHolder{Lift: cr.TotalKg, Lifter: cr.Name}
				}
			}
		}
	}

	db.Records = make([]model.Record, 0, len(records))
	for _, rec := range records {
		db.Records = append(db.Records, *rec)
	}
	sort.Slice(db.Records, func(i, j int) bool {
		if db.Records[i].Sex != db.Records[j].Sex {
			return db.Records[i].Sex < db.Records[j].Sex
		}
		if db.Records[i].Equipment != db.Records[j].Equipment {
			return db.Records[i].Equipment < db.Records[j].Equipment
		}
		return db.Records[i].WeightClassKg < db.Records[j].WeightClassKg
	})
}
