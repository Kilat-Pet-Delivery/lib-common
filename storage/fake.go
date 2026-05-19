package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// FakeStorage is an in-memory Storage implementation for unit tests.
type FakeStorage struct {
	mu      sync.RWMutex
	bucket  string
	objects map[string][]byte
}

// NewFakeStorage creates an in-memory storage fake with a deterministic bucket name.
func NewFakeStorage() *FakeStorage {
	return NewFakeStorageWithBucket("fake")
}

// NewFakeStorageWithBucket creates an in-memory storage fake with the given bucket name.
func NewFakeStorageWithBucket(bucket string) *FakeStorage {
	if bucket == "" {
		bucket = "fake"
	}
	return &FakeStorage{
		bucket:  bucket,
		objects: make(map[string][]byte),
	}
}

// Upload stores the reader bytes at key and returns a canonical fake storage URL.
func (s *FakeStorage) Upload(_ context.Context, key string, reader io.Reader) (string, error) {
	key, err := s.normalizeKey(key)
	if err != nil {
		return "", err
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read upload bytes: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = append([]byte(nil), data...)
	return s.objectURL(key), nil
}

// Download returns a copy of the stored bytes for key.
func (s *FakeStorage) Download(_ context.Context, key string) (io.ReadCloser, error) {
	key, err := s.normalizeKey(key)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.objects[key]
	if !ok {
		return nil, ErrObjectNotFound
	}
	return io.NopCloser(bytes.NewReader(append([]byte(nil), data...))), nil
}

// SignedURL returns a deterministic in-memory URL containing the expiry timestamp.
func (s *FakeStorage) SignedURL(_ context.Context, key string, ttl time.Duration) (string, error) {
	key, err := s.normalizeKey(key)
	if err != nil {
		return "", err
	}
	if ttl <= 0 {
		return "", fmt.Errorf("signed URL ttl must be positive")
	}
	expiresAt := time.Now().UTC().Add(ttl).Unix()
	return fmt.Sprintf("%s?expires_at=%d", s.objectURL(key), expiresAt), nil
}

// Delete removes key if present.
func (s *FakeStorage) Delete(_ context.Context, key string) error {
	key, err := s.normalizeKey(key)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, key)
	return nil
}

func (s *FakeStorage) objectURL(key string) string {
	return fmt.Sprintf("memory://%s/%s", s.bucket, key)
}

func (s *FakeStorage) normalizeKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("storage key is required")
	}
	for _, prefix := range []string{
		"memory://" + s.bucket + "/",
		"minio://" + s.bucket + "/",
		"s3://" + s.bucket + "/",
	} {
		key = strings.TrimPrefix(key, prefix)
	}
	key = strings.TrimPrefix(key, "/")
	if key == "" || strings.Contains(key, "..") {
		return "", fmt.Errorf("invalid storage key %q", key)
	}
	return key, nil
}
