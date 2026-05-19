package storage

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Test_Fake_RoundTrip_PersistsBytes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	storage := NewFakeStorage()

	url, err := storage.Upload(ctx, "chat/thread-1/message-1.txt", strings.NewReader("hello kilat"))
	if err != nil {
		t.Fatalf("upload fake object: %v", err)
	}

	reader, err := storage.Download(ctx, url)
	if err != nil {
		t.Fatalf("download fake object: %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read fake object: %v", err)
	}
	if string(got) != "hello kilat" {
		t.Fatalf("downloaded bytes = %q, want %q", got, "hello kilat")
	}
}

func Test_MinIO_RoundTrip_PersistsBytes(t *testing.T) {
	ctx := context.Background()
	storage := newTestMinIOStorage(t)

	body := []byte("minio stores this")
	url, err := storage.Upload(ctx, "proof/booking-1/photo.jpg", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("upload minio object: %v", err)
	}

	reader, err := storage.Download(ctx, url)
	if err != nil {
		t.Fatalf("download minio object: %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read minio object: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Fatalf("downloaded bytes = %q, want %q", got, body)
	}
}

func Test_SignedURL_ExpiresAfterTTL(t *testing.T) {
	ctx := context.Background()
	storage := newTestMinIOStorage(t)

	url, err := storage.Upload(ctx, "chat/thread-1/message-2.txt", strings.NewReader("short lived"))
	if err != nil {
		t.Fatalf("upload minio object: %v", err)
	}

	signedURL, err := storage.SignedURL(ctx, url, time.Second)
	if err != nil {
		t.Fatalf("sign minio object: %v", err)
	}

	before, err := http.Get(signedURL)
	if err != nil {
		t.Fatalf("get signed URL before expiry: %v", err)
	}
	_ = before.Body.Close()
	if before.StatusCode != http.StatusOK {
		t.Fatalf("signed URL before expiry status = %d, want 200", before.StatusCode)
	}

	time.Sleep(2 * time.Second)

	after, err := http.Get(signedURL)
	if err != nil {
		t.Fatalf("get signed URL after expiry: %v", err)
	}
	_ = after.Body.Close()
	if after.StatusCode == http.StatusOK {
		t.Fatalf("signed URL after expiry status = 200, want non-200")
	}
}

func newTestMinIOStorage(t *testing.T) *MinIOStorage {
	t.Helper()

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:RELEASE.2025-04-22T22-12-26Z",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     "minioadmin",
			"MINIO_ROOT_PASSWORD": "minioadmin",
		},
		Cmd: []string{"server", "/data", "--console-address", ":9001"},
		WaitingFor: wait.ForHTTP("/minio/health/ready").
			WithPort("9000/tcp").
			WithStartupTimeout(60 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("skipping MinIO integration test; container provider unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("read minio host: %v", err)
	}
	port, err := container.MappedPort(ctx, "9000")
	if err != nil {
		t.Fatalf("read minio port: %v", err)
	}

	storage, err := NewMinIOStorage(MinIOConfig{
		Endpoint:  host + ":" + port.Port(),
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		Bucket:    "kilat-test-" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		Secure:    false,
	})
	if err != nil {
		t.Fatalf("create minio storage: %v", err)
	}
	return storage
}
