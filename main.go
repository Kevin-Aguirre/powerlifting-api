package main

import (
	"fmt"
	"net/http"
	"github.com/Kevin-Aguirre/powerlifting-api/api"
)

func main() {
	fmt.Println("Starting server on :3333")
	err := http.ListenAndServe(":3333", api.NewRouter())
	if err != nil {
		fmt.Println("Server error:", err)
	}
}