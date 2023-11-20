package main

import (
	"io"
	"log"
	"net/http"

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

	// Stream the file back to the client
	io.Copy(w, reader)
}
