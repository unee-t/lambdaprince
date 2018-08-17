package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

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
	var out []byte
	path, err := exec.LookPath("./static/hello")
	if err == nil {
		out, err = exec.Command(path).CombinedOutput()
		log.Infof("out: %s", out)
		if err != nil {
			log.WithError(err).Warnf("hello failed: %s", out)
		}
	}
	log.Infof("out here: %s", out)
	fmt.Fprintf(w, string(out))
}
