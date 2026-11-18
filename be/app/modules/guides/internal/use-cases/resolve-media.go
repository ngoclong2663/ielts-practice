package usecases

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"ielts-practice/app/modules/guides/domain/rules"
	"ielts-practice/pkg/exceptions"
)

// mediaTTL is how long a presigned media link works; MediaCacheMaxAge keeps a cached redirect inside it.
const (
	mediaTTL         = 2 * time.Hour
	MediaCacheMaxAge = 90 * time.Minute
)

var mediaFile = regexp.MustCompile(`^[0-9a-f]{64}\.(mp4|jpg)$`)

// ResolveMedia returns a short-lived S3 link for the video or poster of a published guide. The bucket is
// private; only keys a published guide uses are signed.
func (uc *GuidesUseCases) ResolveMedia(ctx context.Context, file string) (string, error) {
	notFound := exceptions.NewException(http.StatusNotFound, exceptions.CodeGuideNotFound, "Guide media not found", nil)
	if !mediaFile.MatchString(file) || !uc.storage.Configured() {
		return "", notFound
	}
	key := rules.KeyForFile(file)
	ok, err := uc.repo.HasMedia(ctx, key)
	if err != nil {
		return "", exceptions.ErrInternalServerError("Failed to load guide media", err)
	}
	if !ok {
		return "", notFound
	}
	url, err := uc.storage.PresignGet(ctx, key, mediaTTL)
	if err != nil {
		return "", exceptions.ErrInternalServerError("Failed to sign guide media link", err)
	}
	return url, nil
}
