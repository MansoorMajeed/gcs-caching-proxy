package main

import (
	"context"
	"log"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
)

var bucketName string
var gcsClient *storage.Client

func main() {

	bucketName = os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("GCS_BUCKET_NAME environment variable is not set")
	}

	// Authenticate with GCS, this is only for startup to ensure that
	// we have access to the bucket
	ctx := context.Background()
	var err error
	gcsClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create GCS client: %v", err)
	}

	// Confirm bucket exists
	if _, err := gcsClient.Bucket(bucketName).Attrs(ctx); err != nil {
		log.Fatalf("Failed to get bucket %v: %v", bucketName, err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/{bucket}/{object:.*}", requestHandler)
	r.HandleFunc("/health", healthCheck)

	srv := startServer(":8080", r)
	gracefulShutdown(srv, gcsClient, 30*time.Second)
}
