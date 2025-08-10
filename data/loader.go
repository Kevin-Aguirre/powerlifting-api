package data

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"github.com/Kevin-Aguirre/powerlifting-api/model"
)

// type Database struct {
// 	// FederationMeets map[string][]*model.Meet
// 	LifterHistory map[string]*model.Lifter
// }

type Database struct {
	LifterHistory map[string]*model.Lifter
}

// findIndex parses a header row in a csv file and given a column name, returns the index if found, else -1.
func findIndex(headerRow []string, columnName string) int {
	for i, value := range headerRow {
		if value == columnName {
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
func computeColumnsMap(records [][]string) map[string]int {
	columnsMap := make(map[string]int)
	possibleColumns := []string {
		"Place",
		"Name",
		"BirthDate",
		"Sex",
		"BirthYear",
		"Age",
		"Country",
		"State",
		"Equipment",
		"Division",
		"BodyweightKg",
		"WeightClassKg",
		"Squat1Kg",
		"Squat2Kg",
		"Squat3Kg",
		"Best3SquatKg",
		"Squat4Kg",
		"Bench1Kg",
		"Bench2Kg",
		"Bench3Kg",
		"Best3BenchKg",
		"Bench4Kg",
		"Deadlift1Kg",
		"Deadlift2Kg",
		"Deadlift3Kg",
		"Best3DeadliftKg",
		"Deadlift4Kg",
		"TotalKg",
		"Event",
		"Tested",
	}

	for _, column := range possibleColumns {
		index := findIndex(records[0], column)
		columnsMap[column] = index
	}

	return columnsMap
}

func LoadDatabase(root string) (*Database, error) {
	// create Database object 
	db := &Database {
		// FederationMeets: make(map[string][]*model.Meet),
		LifterHistory: make(map[string]*model.Lifter),
	}

	err := filepath.WalkDir(root, func(currPath string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", currPath, err)
			return err
		}

		splitDir := "meet-data"
		relativePath := strings.Split(currPath, splitDir)[1]
		// pathArr := strings.Split(relativePath, "/")

		if (!strings.Contains(relativePath, "entries.csv")) {
			return nil
		}

		// attempt to open file
		file, err := os.Open(currPath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return nil
		}
		defer file.Close()

		// read lines 
		reader := csv.NewReader(file)
		reader.LazyQuotes = true
		records, err := reader.ReadAll()

		if err != nil {
			fmt.Println("Could not read lines:", err)
			return nil
		}

		columnsMap := computeColumnsMap(records)

		// iterate through non-header rows 
		for _, row := range records[1:] {
			// extract lifter name 
			lifterName := row[columnsMap["Name"]]

			// construct LifterMeetResult for each row 
			lifterResult := &model.LifterMeetResult{
				Place: 			getValue(row, columnsMap, "Place"),
				Name:           getValue(row, columnsMap, "Name"),
				BirthDate:      getValue(row, columnsMap, "BirthDate"),
				Sex:            getValue(row, columnsMap, "Sex"),
				BirthYear:      getValue(row, columnsMap, "BirthYear"),
				Age:            getValue(row, columnsMap, "Age"),
				Country:        getValue(row, columnsMap, "Country"),
				State:          getValue(row, columnsMap, "State"),
				Equipment:      getValue(row, columnsMap, "Equipment"),
				Division:       getValue(row, columnsMap, "Division"),
				BodyweightKg:   getValue(row, columnsMap, "BodyweightKg"),
				WeightClassKg:  getValue(row, columnsMap, "WeightClassKg"),
				Squat1Kg:       getValue(row, columnsMap, "Squat1Kg"),
				Squat2Kg:       getValue(row, columnsMap, "Squat2Kg"),
				Squat3Kg:       getValue(row, columnsMap, "Squat3Kg"),
				Best3SquatKg:   getValue(row, columnsMap, "Best3SquatKg"),
				Squat4Kg:       getValue(row, columnsMap, "Squat4Kg"),
				Bench1Kg:       getValue(row, columnsMap, "Bench1Kg"),
				Bench2Kg:       getValue(row, columnsMap, "Bench2Kg"),
				Bench3Kg:       getValue(row, columnsMap, "Bench3Kg"),
				Best3BenchKg:   getValue(row, columnsMap, "Best3BenchKg"),
				Bench4Kg:       getValue(row, columnsMap, "Bench4Kg"),
				Deadlift1Kg:    getValue(row, columnsMap, "Deadlift1Kg"),
				Deadlift2Kg:    getValue(row, columnsMap, "Deadlift2Kg"),
				Deadlift3Kg:    getValue(row, columnsMap, "Deadlift3Kg"),
				Best3DeadliftKg:getValue(row, columnsMap, "Best3DeadliftKg"),
				Deadlift4Kg:    getValue(row, columnsMap, "Deadlift4Kg"),
				TotalKg:        getValue(row, columnsMap, "TotalKg"),
				Event:          getValue(row, columnsMap, "Event"),
				Tested:         getValue(row, columnsMap, "Tested"),
			}

			if _, exists := db.LifterHistory[lifterName]; !exists {
				db.LifterHistory[lifterName] = &model.Lifter{
					Name: lifterName,
					PB: make(map[string]*model.PersonalBest),
					CompetitionResults: []*model.LifterMeetResult{},
				}
			}
			db.LifterHistory[lifterName].CompetitionResults = append(
				db.LifterHistory[lifterName].CompetitionResults, 
				lifterResult,
			)	
		}
		
		return nil
	})
	
	
	if err != nil {
		fmt.Printf("Error walking the directory: %v\n", err)
	}

	return db, nil
}