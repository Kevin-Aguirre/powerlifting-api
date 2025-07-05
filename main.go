package main

import (
	"fmt"
	"net/http"

	"github.com/Kevin-Aguirre/powerlifting-api/api"
	"github.com/Kevin-Aguirre/powerlifting-api/data"
)

func main() {
	fmt.Println("Loading powerlifting data...")
	var dataFolderPath = "/home/kevin/dev/Projects/powerlifting-api/opl-data/meet-data"
	db, db_err := data.LoadDatabase(dataFolderPath)
	if db_err != nil {
		panic(fmt.Errorf("Failed to load data: %w", db_err))
	}

	fmt.Println("Successfully loaded data")
	fmt.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", api.NewRouter(db))
	if err != nil {
		fmt.Println("Server error:", err)
	}
}