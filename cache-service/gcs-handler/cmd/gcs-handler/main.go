package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
	"github.com/mansoormajeed/gcs-caching-proxy/cache-service/gcs-handler/pkg/gcs"
)

func main() {

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

	bucket := vars["bucket"]
	object := vars["object"]

	// Create a GCS client
	client, err := gcs.NewClient(ctx)
	if err != nil {
		log.Println("error creating GCS client:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()

	reader, err := gcs.ReadFile(ctx, client, bucket, object)
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

	attrs, err := gcs.GetObjectAttrs(ctx, client, bucket, object)
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
