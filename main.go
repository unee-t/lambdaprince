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
	if err != nil {
		log.WithError(err).Error("no prince binary found")
		http.Error(w, "Error finding the prince binary", http.StatusInternalServerError)
		return
	}
	out, err = exec.Command(path, "--version").CombinedOutput()
	if err != nil {
		log.WithError(err).Errorf("prince failed: %s", out)
		http.Error(w, "Prince failed", http.StatusInternalServerError)
		return
	}
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
		ctx.Fatal("bad source")
		http.Error(w, "Source must be from our S3", 400)
		return
	}

	resp, err := http.Get(input.URL)

	log.Infof("Fetched content type: %s", resp.Header.Get("Content-Type"))

	if err != nil {
		ctx.WithError(err).Fatalf("failed to fetch")
		http.Error(w, err.Error(), 500)
		return
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ctx.WithError(err).Fatalf("failed to read body")
		http.Error(w, err.Error(), 500)
		return
	}

	err = ioutil.WriteFile("/tmp/doc.html", contents, 0644)
	if err != nil {
		ctx.WithError(err).Fatalf("failed to write fetched document_url")
		http.Error(w, err.Error(), 500)
		return
	}

	var out []byte
	path, err := exec.LookPath("./prince/bin/prince")
	if err != nil {
		ctx.WithError(err).Fatal("not a valid URL")
		http.Error(w, "Not a valid URL", http.StatusBadRequest)
		return
	}

	out, err = exec.Command(path, "/tmp/doc.html", "-o", "/tmp/out.pdf").CombinedOutput()
	log.Infof("out: %s", out)
	if err != nil {
		log.WithError(err).Warnf("hello failed: %s", out)
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

	pdfurl := fmt.Sprintf("https://s3-ap-southeast-1.amazonaws.com/dev-media-unee-t/%s", pdffilename)

	log.Infof("Produced %s", pdfurl)

	response.JSON(w, struct {
		PDF string
	}{
		pdfurl,
	})
}
