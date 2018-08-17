package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/apex/log"
	"github.com/gorilla/mux"
)

func main() {
	addr := ":" + os.Getenv("PORT")
	app := mux.NewRouter()
	app.HandleFunc("/", handleIndex).Methods("GET")
	if err := http.ListenAndServe(addr, app); err != nil {
		log.WithError(err).Fatal("error listening")
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Howdy")
}
