package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
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
