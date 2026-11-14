package ports

import (
	"context"
	"io"
	"time"
)

// IObjectStorage stores media (card images, TTS audio). Keys are bucket-relative paths
// such as "media/audio/<hash>.mp3".
type IObjectStorage interface {
	// Configured reports whether a bucket is set; when false every other call returns ErrNotConfigured.
	Configured() bool
	Ping(ctx context.Context) error
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}
