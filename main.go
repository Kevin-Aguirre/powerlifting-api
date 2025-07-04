package main

import (
	"fmt"
	"os"
	"io"
	"net/http"
	"github.com/go-git/go-git/v6" 
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "This is my website!\n")
}
func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	io.WriteString(w, "Hello, HTTP!\n")
}

func main() {
	http.HandleFunc("/", getRoot)
	http.HandleFunc("/hello", getHello)

	err := http.ListenAndServe(":3333", nil)

	if err != nil {
		fmt.Println(err)

	}

	curdir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}	
	var dataRepo = curdir + "/clone-test"

	_, repoExistsErr := os.Stat(dataRepo)
	if repoExistsErr == nil {
		os.RemoveAll(dataRepo)
	}

	var repoToClone = "https://github.com/Kevin-Aguirre/QuickLook-Frontend"

	_, Newerr := git.PlainClone(dataRepo, &git.CloneOptions{
		URL:      repoToClone,
		Progress: os.Stdout,
	})

	if Newerr != nil {
		fmt.Println(Newerr)
	}

	fmt.Println("all finished")
}
