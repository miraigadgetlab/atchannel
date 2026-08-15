package storage

import (
	"context"
	"errors"
	"io"
)

// Storage abstracts object storage so handlers/services never know whether
// files live on local disk or behind an S3-compatible endpoint.
type Storage interface {
	// Put stores data under key and returns the storage key.
	Put(ctx context.Context, key string, data []byte, contentType string) error
	// Get returns a reader for the object at key.
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	// Delete removes the object at key. Missing objects are not an error.
	Delete(ctx context.Context, key string) error
	// URL returns a public URL for the object at key, if the backend can
	// produce one. Local implementations may return a path served by nginx.
	URL(ctx context.Context, key string) (string, error)
	// Stat returns the size of the object, or ErrNotFound.
	Stat(ctx context.Context, key string) (int64, error)
}

var ErrNotFound = errors.New("storage: object not found")