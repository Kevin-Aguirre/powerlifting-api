package handlers

import (
	"fmt"
	"io"
	"net/http"
	"github.com/Kevin-Aguirre/powerlifting-api/data"
)

func GetHello(db *data.Database) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		fmt.Println("got /hello request")
		io.WriteString(w, "Hello, HTTP!\n")
	}
}
