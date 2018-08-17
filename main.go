package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
	"github.com/tj/go/http/response"
)

func main() {
	addr := ":" + os.Getenv("PORT")
	app := mux.NewRouter()
	app.HandleFunc("/", handleIndex).Methods("GET")
	app.HandleFunc("/", handlePost).Methods("POST")
	if err := http.ListenAndServe(addr, app); err != nil {
		log.WithError(err).Fatal("error listening")
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	var out []byte
	path, err := exec.LookPath("./prince/bin/prince")
	if err == nil {
		out, err = exec.Command(path, "--version").CombinedOutput()
		log.Infof("out: %s", out)
		if err != nil {
			log.WithError(err).Warnf("hello failed: %s", out)
		}
	}
	log.Infof("out here: %s", out)
	fmt.Fprintf(w, string(out))
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	type Input struct {
		URL string `json:"document_url"`
	}

	var input Input

	ctx := log.WithFields(log.Fields{
		"method": r.Method,
		"url":    r.URL.String(),
		"input":  input,
		"route":  "handlePost",
	})

	err := decoder.Decode(&input)
	if err != nil {
		ctx.WithError(err).Fatal("failed to read input")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check it's a valid URL
	u, err := url.ParseRequestURI(input.URL)
	if err != nil {
		ctx.WithError(err).Fatal("not a valid URL")
		http.Error(w, "Not a valid URL", http.StatusBadRequest)
		return
	}

	// Make sure the document_url is from our bucket
	if u.Host != "s3-ap-southeast-1.amazonaws.com" &&
		strings.HasPrefix(u.Path, "/dev-media-unee-t/") {
		http.Error(w, "Source must be from our S3", 400)
		return
	}

	fmt.Fprintf(w, "here")
	return

	var out []byte
	path, err := exec.LookPath("./prince/bin/prince")
	if err == nil {
		out, err = exec.Command(path, "doc.html", "-o", "/tmp/out.pdf").CombinedOutput()
		log.Infof("out: %s", out)
		if err != nil {
			log.WithError(err).Warnf("hello failed: %s", out)
		}
	}

	outputpdf, err := ioutil.ReadFile("/tmp/out.pdf")
	if err != nil {
		log.WithError(err).Fatal("failed to read output")
		http.Error(w, err.Error(), 500)
		return
	}

	cfg, err := external.LoadDefaultAWSConfig(external.WithSharedConfigProfile("uneet-dev"))
	if err != nil {
		log.WithError(err).Fatal("failed to get config")
		http.Error(w, err.Error(), 500)
		return
	}
	svc := s3.New(cfg)

	pdffilename := time.Now().Format("2006-01-02") + fmt.Sprintf("/%d.pdf", time.Now().Unix())
	putparams := &s3.PutObjectInput{
		Bucket:      aws.String("dev-media-unee-t"),
		Body:        bytes.NewReader(outputpdf),
		Key:         aws.String(pdffilename),
		ACL:         s3.ObjectCannedACLPublicRead,
		ContentType: aws.String("application/pdf; charset=UTF-8"),
	}

	req := svc.PutObjectRequest(putparams)
	_, err = req.Send()
	if err != nil {
		log.WithError(err).Fatal("failed to upload to s3")
		http.Error(w, err.Error(), 500)
		return
	}

	response.JSON(w, struct {
		PDF string
	}{
		fmt.Sprintf("https://s3-ap-southeast-1.amazonaws.com/dev-media-unee-t/%s", pdffilename),
	})
}
