package handlers

import (
	"fmt"
	"io"
	"net/http"
)

func GetRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got / request")
	io.WriteString(w, "This is my website!\n")
}

func GetHello(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got /hello request")
	io.WriteString(w, "Hello, HTTP!\n")
}
