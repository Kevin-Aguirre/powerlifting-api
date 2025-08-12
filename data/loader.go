package data

import (
	"encoding/csv"
	// "errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Kevin-Aguirre/powerlifting-api/model"
)

const (
	// dots coefficient amounts
	dotsMaleA = -307.75076
	dotsMaleB = 24.0900756
	dotsMaleC = -0.1918759221
	dotsMaleD = 0.0007391293
	dotsMaleE = -0.000001093
	dotsFemaleA = -57.96288
	dotsFemaleB = 13.6175032
	dotsFemaleC = -0.1126655495
	dotsFemaleD = 0.0005158568
	dotsFemaleE = -0.0000010706

	// TODO: determine if this gets used or not 
	// unit conversion rates, constant in case you want more/less precision
	// kgToLbConversionRate = 2.204623

	// column header labels for entries files (who knows if these get changed, better to have them in one place)
	colHeaderPlace = "Place"
	colHeaderName = "Name"
	colHeaderBirthDate = "BirthDate"
	colHeaderSex = "Sex"
	colHeaderBirthYear = "BirthYear"
	colHeaderAge = "Age"
	colHeaderCountry = "Country"
	colHeaderState = "State"
	colHeaderEquipment = "Equipment"
	colHeaderDivision = "Division"
	colHeaderBodyweightKg = "BodyweightKg"
	colHeaderWeightClassKg = "WeightClassKg"
	colHeaderSquat1Kg = "Squat1Kg"
	colHeaderSquat2Kg = "Squat2Kg"
	colHeaderSquat3Kg = "Squat3Kg"
	colHeaderBest3SquatKg = "Best3SquatKg"
	colHeaderSquat4Kg = "Squat4Kg"
	colHeaderBench1Kg = "Bench1Kg"
	colHeaderBench2Kg = "Bench2Kg"
	colHeaderBench3Kg = "Bench3Kg"
	colHeaderBest3BenchKg = "Best3BenchKg"
	colHeaderBench4Kg = "Bench4Kg"
	colHeaderDeadlift1Kg = "Deadlift1Kg"
	colHeaderDeadlift2Kg = "Deadlift2Kg"
	colHeaderDeadlift3Kg = "Deadlift3Kg"
	colHeaderBest3DeadliftKg = "Best3DeadliftKg"
	colHeaderDeadlift4Kg = "Deadlift4Kg"
	colHeaderTotalKg = "TotalKg"
	colHeaderEvent = "Event"
	colHeaderTested = "Tested"

	// column header labels for meet files (also may get changed?)
	colHeaderFederation = "Federation"
	colHeaderDate = "Date"
	colHeaderMeetCountry = "MeetCountry"
	colHeaderMeetState = "MeetState"
	colHeaderMeetTown = "MeetTown"
	colHeaderMeetName = "MeetName"
	colHeaderRuleSet = "RuleSet"

	// file info 
	meetEntriesFileName = "entries.csv"
	meetInfoFileName = "meet.csv"

	// meetsRootFolderName = "meet-data"

)

type Database struct {
	LifterHistory map[string]*model.Lifter
	FederationMeets map[string][]*model.Meet
}

func clean(s string) string {
    return strings.TrimSpace(strings.ReplaceAll(s, "\uFEFF", ""))
}


// findIndex parses a header row in a csv file and given a column name, returns the index if found, else -1.
func findIndex(headerRow []string, columnName string) int {
	for i, value := range headerRow {
		if clean(value) == clean(columnName) {
			return i
		}
	}
	return -1
}

// getValue accepts a csv data row, a columnsMap, and a column name. 
// return the value of the csv row given the index stored in columnMap
func getValue(row []string, columnsMap map[string]int, column string) string {
	idx := columnsMap[column]
	if idx == -1 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

// given a csv file as a slice of string slices, generates map of column name to index.
func computeEntriesColumnsMap(records [][]string) map[string]int {
	columnsMap := make(map[string]int)
	possibleColumns := []string {
		colHeaderPlace,
		colHeaderName,
		colHeaderBirthDate,
		colHeaderSex,
		colHeaderBirthYear,
		colHeaderAge,
		colHeaderCountry,
		colHeaderState,
		colHeaderEquipment,
		colHeaderDivision,
		colHeaderBodyweightKg,
		colHeaderWeightClassKg,
		colHeaderSquat1Kg,
		colHeaderSquat2Kg,
		colHeaderSquat3Kg,
		colHeaderBest3SquatKg,
		colHeaderSquat4Kg,
		colHeaderBench1Kg,
		colHeaderBench2Kg,
		colHeaderBench3Kg,
		colHeaderBest3BenchKg,
		colHeaderBench4Kg,
		colHeaderDeadlift1Kg,
		colHeaderDeadlift2Kg,
		colHeaderDeadlift3Kg,
		colHeaderBest3DeadliftKg,
		colHeaderDeadlift4Kg,
		colHeaderTotalKg,
		colHeaderEvent,
		colHeaderTested,
	}

	for _, column := range possibleColumns {
		index := findIndex(records[0], column)
		columnsMap[column] = index
	}

	return columnsMap
}

// given a csv file as a slice of string slices, generates map of column name to index.
func computeMeetColumnsMap(records [][]string) map[string]int {
	columnsMap := make(map[string]int)
	possibleColumns := []string {
		colHeaderFederation,
		colHeaderDate,
		colHeaderMeetCountry,
		colHeaderMeetState,
		colHeaderMeetTown,
		colHeaderMeetName,
		colHeaderRuleSet,
	}

	for _, column := range possibleColumns {
		index := findIndex(records[0], column)
		columnsMap[column] = index
	}

	return columnsMap
}

func getFederationMeetInfo(row []string, columnsMap map[string]int) (*model.Meet) {
	meetResult := &model.Meet{
		Federation: getValue(row, columnsMap, colHeaderFederation),
		Date: getValue(row, columnsMap, colHeaderDate),
		MeetCountry: getValue(row, columnsMap, colHeaderMeetCountry),
		MeetState: getValue(row, columnsMap, colHeaderMeetState),
		MeetTown: getValue(row, columnsMap, colHeaderMeetTown),
		MeetName: getValue(row, columnsMap, colHeaderMeetName),
		RuleSet: getValue(row, columnsMap, colHeaderRuleSet),
	}
	return meetResult
}

// handles creating a LifterMeetResult object given a row of a csv file 
func getLifterMeetResult(row []string, columnsMap map[string]int) (*model.LifterMeetResult) {
	lifterResult := &model.LifterMeetResult{
		Place: 			getValue(row, columnsMap, colHeaderPlace),
		Name:           getValue(row, columnsMap, colHeaderName),
		BirthDate:      getValue(row, columnsMap, colHeaderBirthDate),
		Sex:            getValue(row, columnsMap, colHeaderSex),
		BirthYear:      getValue(row, columnsMap, colHeaderBirthYear),
		Age:            getValue(row, columnsMap, colHeaderAge),
		Country:        getValue(row, columnsMap, colHeaderCountry),
		State:          getValue(row, columnsMap, colHeaderState),
		Equipment:      getValue(row, columnsMap, colHeaderEquipment),
		Division:       getValue(row, columnsMap, colHeaderDivision),
		BodyweightKg:   getValue(row, columnsMap, colHeaderBodyweightKg),
		WeightClassKg:  getValue(row, columnsMap, colHeaderWeightClassKg),
		Squat1Kg:       getValue(row, columnsMap, colHeaderSquat1Kg),
		Squat2Kg:       getValue(row, columnsMap, colHeaderSquat2Kg),
		Squat3Kg:       getValue(row, columnsMap, colHeaderSquat3Kg),
		Best3SquatKg:   getValue(row, columnsMap, colHeaderBest3SquatKg),
		Squat4Kg:       getValue(row, columnsMap, colHeaderSquat4Kg),
		Bench1Kg:       getValue(row, columnsMap, colHeaderBench1Kg),
		Bench2Kg:       getValue(row, columnsMap, colHeaderBench2Kg),
		Bench3Kg:       getValue(row, columnsMap, colHeaderBench3Kg),
		Best3BenchKg:   getValue(row, columnsMap, colHeaderBest3BenchKg),
		Bench4Kg:       getValue(row, columnsMap, colHeaderBench4Kg),
		Deadlift1Kg:    getValue(row, columnsMap, colHeaderDeadlift1Kg),
		Deadlift2Kg:    getValue(row, columnsMap, colHeaderDeadlift2Kg),
		Deadlift3Kg:    getValue(row, columnsMap, colHeaderDeadlift3Kg),
		Best3DeadliftKg:getValue(row, columnsMap, colHeaderBest3DeadliftKg),
		Deadlift4Kg:    getValue(row, columnsMap, colHeaderDeadlift4Kg),
		TotalKg:        getValue(row, columnsMap, colHeaderTotalKg),
		Event:          getValue(row, columnsMap, colHeaderEvent),
		Tested:         getValue(row, columnsMap, colHeaderTested),
	}
	return lifterResult
}


// TODO: use these or not 
// func lbsToKg(n float64) float64 {
// 	return n / kgToLbConversionRate
// }

// func kgToLbs(n float64) float64 {
// 	return n * kgToLbConversionRate
// }

func calculateDots(
	bodyweightKgs float64,
	totalKgs float64,  
	gender string,
) float64 {
	var a, b, c, d, e float64
	switch gender {
		case "M":
			a = dotsMaleA
			b = dotsMaleB
			c = dotsMaleC
			d = dotsMaleD
			e = dotsMaleE
		case "F":
			a = dotsFemaleA
			b = dotsFemaleB
			c = dotsFemaleC
			d = dotsFemaleD
			e = dotsFemaleE
		// im not sure how to account for Mx or missing sex field, safest option to leave it blank
		case "Mx":
		default:
			return 0
	}

	x := bodyweightKgs
	numerator := 500 * totalKgs
	denominator := a + b*x + c*x*x + d*x*x*x + e*x*x*x*x
	return numerator / denominator

}

func getBestSquat(meetResult *model.LifterMeetResult) float64 {
	squat, err := strconv.ParseFloat(meetResult.Best3SquatKg, 64)
	if err != nil {
		return 0
	}
	return squat
}

func getBestBench(meetResult *model.LifterMeetResult) float64 {
	squat, err := strconv.ParseFloat(meetResult.Best3BenchKg, 64)
	if err != nil {
		return 0
	}
	return squat
}

func ensureLifterExists(db *Database, lifterName string) {
	if _, exists := db.LifterHistory[lifterName]; !exists {
		db.LifterHistory[lifterName] = &model.Lifter{
			Name: lifterName,
			PB: make(map[string]*model.PersonalBest),
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
	db.FederationMeets[federationName] = append(
		db.FederationMeets[federationName],
		meetInfo,
	)
}

func handlePBUpdate(db *Database, lifterResult *model.LifterMeetResult, lifterName string) {
	computeDotsFlag := true 
	s := getBestSquat(lifterResult)
	b := getBestBench(lifterResult)
	d := getBestDeadlift(lifterResult)
	lifterWeightKg, err := strconv.ParseFloat(lifterResult.BodyweightKg, 64)
	if err != nil {
		computeDotsFlag = false
	}
	
	total := s + b + d

	var currDots float64
	if computeDotsFlag {
		currDots = calculateDots(
			lifterWeightKg,
			total,
			lifterResult.Sex,
		)
	} else {
		currDots = 0
	}
	
	rowAsPB := &model.PersonalBest{
		Squat: s,
		Bench: b,
		Deadlift: d,
		Total: total,
		Dots: currDots,
	}

	if _, exists := db.LifterHistory[lifterName].PB[lifterResult.Equipment]; !exists {
		db.LifterHistory[lifterName].PB[lifterResult.Equipment] = rowAsPB
	} else {
		prevDots := db.LifterHistory[lifterName].PB[lifterResult.Equipment].Dots

		if (currDots > prevDots) {
			db.LifterHistory[lifterName].PB[lifterResult.Equipment] = rowAsPB
		}
	}
}

func getBestDeadlift(meetResult *model.LifterMeetResult) float64 {
	squat, err := strconv.ParseFloat(meetResult.Best3DeadliftKg, 64)
	if err != nil {
		return 0
	}
	return squat
}

// TODO: this never returns an error 
func handleEntriesFile(db *Database, currPath string) error {
	// attempt to open file
	file, err := os.Open(currPath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	// read csv lines 
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()

	// check if failed to read lines
	if err != nil {
		fmt.Println("Could not read lines:", err)
		return nil
	}

	// map column header labels to their indices
	columnsMap := computeEntriesColumnsMap(records)

	// iterate through non-header rows 
	for _, row := range records[1:] {
		lifterName := row[columnsMap[colHeaderName]]
		lifterResult := getLifterMeetResult(row, columnsMap)
		
		ensureLifterExists(db, lifterName)
		handleCompetitionResultsUpdate(db, lifterResult, lifterName)
		handlePBUpdate(db, lifterResult, lifterName)
	}
	return nil
}

// TODO: maybe modularize this csv file code 
func handleMeetFile(db *Database, currPath string) error {
	// attempt to open file
	file, err := os.Open(currPath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	// read csv lines 
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	
	// check if failed to read lines
	if err != nil {
		fmt.Println("Could not read lines:", err)
		return nil
	}

	// map column header labels to their indices
	columnsMap := computeMeetColumnsMap(records)

	row := records[1]
	idx := columnsMap[colHeaderFederation]
	federationName := row[idx]
	meetInfo := getFederationMeetInfo(row, columnsMap)
	ensureFederationExists(db, federationName)
	handleFederationMeetUpdate(db, meetInfo, federationName)

	return nil
}

func LoadDatabase(root string) (*Database, error) {
	// create Database object 
	db := &Database {
		FederationMeets: make(map[string][]*model.Meet),
		LifterHistory: make(map[string]*model.Lifter),
	}

	err := filepath.WalkDir(root, func(currPath string, d fs.DirEntry, err error) error {
		// try to access path 
		if err != nil {	
			fmt.Printf("Error accessing path %q: %v\n", currPath, err)
			return err
		}

		// get relative path (not full path)
		arr := strings.Split(currPath, "/")
		fileName := arr[len(arr)-1]

		switch fileName {
			case meetEntriesFileName:
				err := handleEntriesFile(db, currPath)
				if err != nil {
					fmt.Println("Error handling entries file: ", err)
				}
				return nil
			case meetInfoFileName:
				err := handleMeetFile(db, currPath)
				if err != nil {
					fmt.Println("Error handling meet file: ", err)
				}
				return nil
			default: 
				return nil
		}
	})

	if err != nil {
		fmt.Printf("Error walking the directory: %v\n", err)
	}

	return db, nil
}