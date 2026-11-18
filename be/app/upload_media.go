package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/fx"

	"ielts-practice/app/modules/guides/domain/rules"
	"ielts-practice/app/modules/storage"
	storagein "ielts-practice/app/modules/storage/ports/in"
	"ielts-practice/configs"
)

var mediaTypes = map[string]string{".mp4": "video/mp4", ".jpg": "image/jpeg"}

// RunUploadMedia uploads guide videos (.mp4) and posters (.jpg) to S3 under content-addressed keys
// (`api upload-media <file>...`, Epic 7 §5.2) and prints the name to put in seeds/guides/videos.json.
// The same file always gets the same key, so running it twice is harmless.
func RunUploadMedia(files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("usage: api upload-media <video.mp4|poster.jpg>...")
	}
	var store storagein.IObjectStorage
	fxApp := fx.New(
		fx.Provide(configs.LoadConfig, configs.InitLoggerConfig),
		storage.NewStorageModule(),
		fx.Populate(&store),
		fx.NopLogger,
	)
	if err := fxApp.Err(); err != nil {
		return err
	}
	if !store.Configured() {
		return fmt.Errorf("S3 is not configured (S3_BUCKET_NAME and AWS credentials)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	for _, path := range files {
		ext := strings.ToLower(filepath.Ext(path))
		contentType, ok := mediaTypes[ext]
		if !ok {
			return fmt.Errorf("%s: only .mp4 videos and .jpg posters", path)
		}
		name, size, err := hashFile(path)
		if err != nil {
			return err
		}
		name += ext
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		key := rules.KeyForFile(name)
		err = store.Put(ctx, key, f, size, contentType)
		_ = f.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		fmt.Printf("%s → %s (%d bytes)\n  seed name: %s\n", path, key, size, name)
	}
	return nil
}

func hashFile(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), n, err
}
