package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrObjectNotFound is returned when an object key is not present.
var ErrObjectNotFound = errors.New("storage object not found")

// Storage provides the shared object-storage contract used by services that
// upload user-owned files such as chat attachments and proof photos.
type Storage interface {
	Upload(ctx context.Context, key string, reader io.Reader) (string, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}
