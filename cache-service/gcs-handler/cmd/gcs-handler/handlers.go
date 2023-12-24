package main

import (
	"errors"
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
	// We could obviously hardcode the bucket as well, but I think this is more
	// flexible. We can use the same credentials with multiple buckets if we need.
	if requestedBucket != bucketName {
		http.Error(w, fmt.Sprintf("[%s] Unknown bucket requested", requestedBucket), http.StatusNotFound)
		return
	}

	reader, err := gcs.ReadFile(ctx, gcsClient, bucketName, object)
	if err != nil {

		log.Println("error reading file from GCS:", err)

		if errors.Is(err, storage.ErrObjectNotExist) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		} else if errors.Is(err, storage.ErrBucketNotExist) {
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

	_, err = io.Copy(w, reader)
	if err != nil {
		log.Printf("error streaming file to client: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := io.WriteString(w, "Happy and Healthy\n")
	if err != nil {
		log.Printf("error writing health check response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
