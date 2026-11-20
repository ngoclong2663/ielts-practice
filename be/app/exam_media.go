package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/fx"

	"ielts-practice/app/modules/exam/domain/entity"
	"ielts-practice/app/modules/storage"
	storagein "ielts-practice/app/modules/storage/ports/in"
	"ielts-practice/configs"
)

var examMediaTypes = map[string]string{".mp3": "audio/mpeg", ".m4a": "audio/mp4"}

// RunExamMedia uploads listening audio built by scripts/exam-audio to S3 (`api exam-media <file.mp3>...`,
// Epic 10 §8.3). Keys are content-addressed (exam/media/<sha256>.mp3), the same name the build wrote into
// the seed file, so running it twice is harmless.
func RunExamMedia(files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("usage: api exam-media <file.mp3>...")
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	for _, path := range files {
		ext := strings.ToLower(filepath.Ext(path))
		contentType, ok := examMediaTypes[ext]
		if !ok {
			return fmt.Errorf("%s: only .mp3 or .m4a audio", path)
		}
		sum, size, err := hashFile(path)
		if err != nil {
			return err
		}
		key := entity.MediaKeyPrefix + sum + ext
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		err = store.Put(ctx, key, f, size, contentType)
		_ = f.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		fmt.Printf("%s → %s (%d bytes)\n", path, key, size)
	}
	return nil
}
