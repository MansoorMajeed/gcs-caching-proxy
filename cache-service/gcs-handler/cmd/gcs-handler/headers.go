package main

import (
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
)

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
