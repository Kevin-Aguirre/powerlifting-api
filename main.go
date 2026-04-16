package main

import (
	"fmt"
	"net/http"

	"github.com/Kevin-Aguirre/powerlifting-api/api"
	"github.com/Kevin-Aguirre/powerlifting-api/data"
)

func main() {
	ds := data.NewDataStore()

	err := ds.Init(data.DefaultRepoPath, data.DefaultRepoURL, data.DefaultDataPath, data.RefreshInterval)
	if err != nil {
		panic(fmt.Errorf("failed to initialize data: %w", err))
	}

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", api.NewRouter(ds)); err != nil {
		fmt.Println("Server error:", err)
	}
}
