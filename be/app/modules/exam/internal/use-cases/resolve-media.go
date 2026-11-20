package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"ielts-practice/app/modules/exam/domain/entity"
	"ielts-practice/pkg/exceptions"
	"ielts-practice/seeds"
)

// Audio and figures live in S3 under content-addressed keys (Epic 10 §8.1).
const (
	mediaRoute       = "/api/v1/exam/media/"
	mediaTTL         = 2 * time.Hour
	MediaCacheMaxAge = 90 * time.Minute
)

var mediaFile = regexp.MustCompile(`^[0-9a-f]{64}\.(mp3|m4a|svg|png)$`)

// MediaURL maps a stored key (exam/media/<sha256>.mp3) to the API path the browser loads.
func MediaURL(key string) string {
	return mediaRoute + strings.TrimPrefix(key, entity.MediaKeyPrefix)
}

// ResolveMedia returns a short-lived S3 link for an exam media file. The bucket is private and the names are
// content hashes, so only files someone already knows the hash of can be fetched.
func (uc *ExamUseCases) ResolveMedia(ctx context.Context, file string) (string, error) {
	notFound := exceptions.NewException(http.StatusNotFound, exceptions.CodeExamMediaNotFound, "Media not found", nil)
	if !mediaFile.MatchString(file) || !uc.storage.Configured() {
		return "", notFound
	}
	url, err := uc.storage.PresignGet(ctx, entity.MediaKeyPrefix+file, mediaTTL)
	if err != nil {
		return "", exceptions.ErrInternalServerError("Failed to sign media link", err)
	}
	return url, nil
}

// FigureRoute is where the browser loads a map, plan or diagram (public, immutable: the name is its hash).
const FigureRoute = "/api/v1/exam/figures/"

var (
	figureOnce  sync.Once
	figureIndex map[string][]byte
)

// ResolveFigure returns an exam figure compiled into the binary (seeds/exam/<set>/figures/*.svg), by its
// content name (<sha256>.svg). Figures are small, so no S3 round trip is needed.
func (uc *ExamUseCases) ResolveFigure(name string) ([]byte, bool) {
	figureOnce.Do(func() {
		figureIndex = map[string][]byte{}
		matches, _ := fs.Glob(seeds.Exam, "exam/*/figures/*.svg")
		for _, m := range matches {
			b, err := fs.ReadFile(seeds.Exam, m)
			if err != nil {
				continue
			}
			sum := sha256.Sum256(b)
			figureIndex[hex.EncodeToString(sum[:])+".svg"] = b
		}
	})
	b, ok := figureIndex[name]
	return b, ok
}
