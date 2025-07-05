package data

import (
	// "errors"
	// "os"
	// "fmt"
	// "path/filepath"
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

	return db, nil
}