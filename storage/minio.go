package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const defaultMinIOBucket = "kilat-runner"

// MinIOConfig contains the settings required to connect to a MinIO/S3-compatible endpoint.
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Secure    bool
}

// MinIOStorage implements Storage against MinIO's S3-compatible API.
type MinIOStorage struct {
	client *minio.Client
	bucket string
}

// NewMinIOStorageFromEnv builds MinIO storage from MINIO_* environment variables.
func NewMinIOStorageFromEnv() (*MinIOStorage, error) {
	secure, err := strconv.ParseBool(envOrDefault("MINIO_USE_SSL", "false"))
	if err != nil {
		return nil, fmt.Errorf("invalid MINIO_USE_SSL: %w", err)
	}

	return NewMinIOStorage(MinIOConfig{
		Endpoint:  envOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey: envOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey: envOrDefault("MINIO_SECRET_KEY", "minioadmin"),
		Bucket:    envOrDefault("MINIO_BUCKET", defaultMinIOBucket),
		Secure:    secure,
	})
}

// NewMinIOStorage creates a new MinIO-backed storage client.
func NewMinIOStorage(cfg MinIOConfig) (*MinIOStorage, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("minio endpoint is required")
	}
	if cfg.AccessKey == "" {
		return nil, fmt.Errorf("minio access key is required")
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("minio secret key is required")
	}
	if cfg.Bucket == "" {
		cfg.Bucket = defaultMinIOBucket
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.Secure,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	return &MinIOStorage{client: client, bucket: cfg.Bucket}, nil
}

// Upload stores the reader at key and returns a canonical storage URL.
func (s *MinIOStorage) Upload(ctx context.Context, key string, reader io.Reader) (string, error) {
	key, err := s.normalizeKey(key)
	if err != nil {
		return "", err
	}
	if err := s.ensureBucket(ctx); err != nil {
		return "", err
	}
	if _, err := s.client.PutObject(ctx, s.bucket, key, reader, -1, minio.PutObjectOptions{}); err != nil {
		return "", fmt.Errorf("upload object %q: %w", key, err)
	}
	return s.objectURL(key), nil
}

// Download fetches an object by key or canonical storage URL.
func (s *MinIOStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	key, err := s.normalizeKey(key)
	if err != nil {
		return nil, err
	}

	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("download object %q: %w", key, err)
	}
	if _, err := object.Stat(); err != nil {
		_ = object.Close()
		return nil, fmt.Errorf("stat object %q: %w", key, err)
	}
	return object, nil
}

// SignedURL returns a temporary GET URL for a key or canonical storage URL.
func (s *MinIOStorage) SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	key, err := s.normalizeKey(key)
	if err != nil {
		return "", err
	}
	if ttl <= 0 {
		return "", fmt.Errorf("signed URL ttl must be positive")
	}

	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, ttl, url.Values{})
	if err != nil {
		return "", fmt.Errorf("sign object %q: %w", key, err)
	}
	return u.String(), nil
}

// Delete removes an object by key or canonical storage URL.
func (s *MinIOStorage) Delete(ctx context.Context, key string) error {
	key, err := s.normalizeKey(key)
	if err != nil {
		return err
	}
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}

func (s *MinIOStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *MinIOStorage) objectURL(key string) string {
	return fmt.Sprintf("minio://%s/%s", s.bucket, key)
}

func (s *MinIOStorage) normalizeKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", fmt.Errorf("storage key is required")
	}
	for _, prefix := range []string{
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

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
