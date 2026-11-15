package live

// Real S3 round trip. Runs only with S3_INTEGRATION=1 and the AWS/S3 variables set, e.g.:
//   set -a; . ./.env; set +a; S3_INTEGRATION=1 go test ./tests/live/ -run S3

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"ielts-practice/app/modules/storage/infrastructure/s3"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"ielts-practice/configs"
)

func TestS3RoundTrip(t *testing.T) {
	if os.Getenv("S3_INTEGRATION") != "1" {
		t.Skip("set S3_INTEGRATION=1 to run against the real bucket")
	}
	cfg := &configs.Config{Storage: configs.StorageConfig{
		Bucket: os.Getenv("S3_BUCKET_NAME"),
		Region: os.Getenv("AWS_REGION"),
	}}
	store, err := s3.NewS3Storage(cfg, zerolog.Nop())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := store.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}

	key := "healthcheck/" + uuid.NewString() + ".txt"
	body := "hello from ielts-practice"
	if err := store.Put(ctx, key, strings.NewReader(body), int64(len(body)), "text/plain"); err != nil {
		t.Fatalf("put: %v", err)
	}
	t.Cleanup(func() { _ = store.Delete(context.Background(), key) })

	url, err := store.PresignGet(ctx, key, time.Minute)
	if err != nil {
		t.Fatalf("presign: %v", err)
	}
	res, err := http.Get(url)
	if err != nil {
		t.Fatalf("get presigned: %v", err)
	}
	got, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != 200 || string(got) != body {
		t.Fatalf("presigned GET returned %d %q", res.StatusCode, got)
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("delete: %v", err)
	}
	res, err = http.Get(url)
	if err == nil {
		_ = res.Body.Close()
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 after delete, got %d", res.StatusCode)
		}
	}
}
