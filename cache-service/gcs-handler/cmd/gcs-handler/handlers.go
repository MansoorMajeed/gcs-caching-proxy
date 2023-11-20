package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"github.com/mansoormajeed/gcs-caching-proxy/cache-service/gcs-handler/pkg/gcs"
)

func requestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	requestedBucket := vars["bucket"]
	object := vars["object"]

	log.Printf("[gcs] received request for [%s/%s]", requestedBucket, object)

	// Do not allow bucket names that do not match the configured bucket
	if requestedBucket != bucketName {
		http.Error(w, fmt.Sprintf("[%s] Unknown bucket requested", requestedBucket), http.StatusNotFound)
		return
	}

	reader, err := gcs.ReadFile(ctx, gcsClient, bucketName, object)
	if err != nil {

		log.Println("error reading file from GCS:", err)

		if err == storage.ErrObjectNotExist {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		} else if err == storage.ErrBucketNotExist {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	attrs, err := gcs.GetObjectAttrs(ctx, gcsClient, bucketName, object)
	if err != nil {
		log.Printf("error getting object attributes: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set headers based on object attributes
	setHeaders(w, attrs)
	// Stream the file back to the client
	io.Copy(w, reader)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Happy and Healthy\n")
}
