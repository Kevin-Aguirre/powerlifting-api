package data

import (
	"encoding/csv"
	"fmt"
	// "io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	// "strings"

	"github.com/Kevin-Aguirre/powerlifting-api/model"
)

type Database struct {
	// FederationMeets map[string][]*model.Meet
	LifterHistory map[string][]*model.Lifter
}

func LoadDatabase(root string) (*Database, error) {
	// create Database object 
	db := &Database {
		// FederationMeets: make(map[string][]*model.Meet),
		LifterHistory: make(map[string][]*model.Lifter),
	}
	
	minColumns := 1000
	minColPath := ""

	maxColumns := 0 
	maxColPath := ""

	err := filepath.WalkDir(root, func(currPath string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", currPath, err)
			return err
		}

		if strings.Contains(currPath, "entries.csv") {
			file, err := os.Open(currPath)
			if err != nil {
				fmt.Println("Error opening file:", err)
				return nil
			}
			defer file.Close()
	
			reader := csv.NewReader(file)
			reader.LazyQuotes = true
			records, err := reader.ReadAll()

			if err != nil {
				fmt.Println("Error opening file:", err)
				fmt.Println(currPath)
				return nil
			}

			if len(records[0]) < minColumns {
				minColumns = len(records[0])
				minColPath = currPath
			}

			if len(records[0]) > maxColumns {
				maxColumns = len(records[0])
				maxColPath = currPath
			}
		}
		
		// splitDir := "meet-data"
		// relativePath := strings.Split(currPath, splitDir)[1]
		// pathArr := strings.Split(relativePath, "/")
		// fmt.Println(pathArr)

		return nil
	})

	fmt.Println(minColumns)
	fmt.Println(minColPath)
	fmt.Println(maxColumns)
	fmt.Println(maxColPath)

	if err != nil {
		fmt.Printf("Error walking the directory: %v\n", err)
	}

	return db, nil
}