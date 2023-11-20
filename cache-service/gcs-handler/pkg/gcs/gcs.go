package gcs

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

// NewClient creates a new GCS client
func NewClient(ctx context.Context) (*storage.Client, error) {
	return storage.NewClient(ctx)
}

// ReadFile reads a file from GCS
func ReadFile(ctx context.Context, client *storage.Client, bucket, object string) (io.Reader, error) {
	return client.Bucket(bucket).Object(object).NewReader(ctx)
}
