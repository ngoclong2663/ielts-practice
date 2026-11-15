package usecases

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"ielts-practice/pkg/exceptions"
)

// figureTTL is how long a presigned figure link works; browsers may cache the redirect a little less.
const figureTTL = time.Hour

var figureFile = regexp.MustCompile(`^[0-9a-f]{64}\.(svg|png|jpg|jpeg|webp)$`)

// ResolveFigure returns a short-lived S3 link for a seeded figure (Epic 5 §6.4). The bucket is private,
// so the database stores only the object key and the API hands out presigned links on demand.
func (uc *WritingUseCases) ResolveFigure(ctx context.Context, file string) (string, error) {
	if !figureFile.MatchString(file) || !uc.storage.Configured() {
		return "", exceptions.NewException(http.StatusNotFound, exceptions.CodeImageNotFound, "Figure not found", nil)
	}
	url, err := uc.storage.PresignGet(ctx, FigureKeyPrefix+file, figureTTL)
	if err != nil {
		return "", exceptions.ErrInternalServerError("Failed to sign figure link", err)
	}
	return url, nil
}

// FigureCacheMaxAge is shorter than the presigned link's lifetime, so a cached redirect never points at an expired link.
const FigureCacheMaxAge = 50 * time.Minute
