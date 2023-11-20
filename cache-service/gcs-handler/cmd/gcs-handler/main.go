package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"github.com/mansoormajeed/gcs-caching-proxy/cache-service/gcs-handler/pkg/gcs"
)

var bucketName string

func main() {

	bucketName = os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("GCS_BUCKET_NAME environment variable is not set")
	}

	// Authenticate with GCS, this is only for startup to ensure that
	// we have access to the bucket
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}
	defer client.Close()
	// Confirm bucket exists
	if _, err := client.Bucket(bucketName).Attrs(ctx); err != nil {
		log.Fatalf("Failed to get bucket %v: %v", bucketName, err)
	}

	r := mux.NewRouter()

	// Define your routes
	r.HandleFunc("/{bucket}/{object:.*}", requestHandler)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}

	// Set up HTTP server and routes
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/", requestHandler)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)

	requestedBucket := vars["bucket"]
	object := vars["object"]

	// Do not allow bucket names that do not match the configured bucket
	if requestedBucket != bucketName {
		http.Error(w, fmt.Sprintf("[%s] Unknown bucket requested", requestedBucket), http.StatusNotFound)
		return
	}

	// Create a GCS client
	client, err := gcs.NewClient(ctx)
	if err != nil {
		log.Println("error creating GCS client:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	reader, err := gcs.ReadFile(ctx, client, bucketName, object)
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

	attrs, err := gcs.GetObjectAttrs(ctx, client, bucketName, object)
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

func setHeaders(w http.ResponseWriter, attrs *storage.ObjectAttrs) {
	// Set Content-Type
	if attrs.ContentType != "" {
		w.Header().Set("Content-Type", attrs.ContentType)
	}

	// Set Cache-Control
	if attrs.CacheControl != "" {
		w.Header().Set("Cache-Control", attrs.CacheControl)
	}

	if attrs.ContentEncoding != "" {
		w.Header().Set("Content-Encoding", attrs.ContentEncoding)
	}

	if attrs.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(attrs.Size, 10))
	}

	if attrs.Etag != "" {
		w.Header().Set("ETag", attrs.Etag)
	}

	if attrs.ContentDisposition != "" {
		w.Header().Set("Content-Disposition", attrs.ContentDisposition)
	}

	if attrs.Updated != (time.Time{}) {
		w.Header().Set("Last-Modified", attrs.Updated.Format(http.TimeFormat))
	}
}
