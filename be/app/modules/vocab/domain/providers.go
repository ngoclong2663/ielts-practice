package domain

import (
	"context"
	"errors"

	"ielts-practice/app/modules/vocab/domain/entity"
)

// WordRequest is one entry sent to the content generator.
type WordRequest struct {
	Input     string
	SenseHint string
}

// WordResult is the generator's answer for one entry.
type WordResult struct {
	Input        string
	IsValid      bool
	Suggestion   string
	Content      entity.CardContent
	ImageQuery   string
	Visualizable bool
	QualityFlags []string
}

// ErrGeneratorNotConfigured: no API key, so the item fails with a clear error code.
var ErrGeneratorNotConfigured = errors.New("content generator is not configured")

// ErrGeneratorRefused: the model declined the request (not retried).
var ErrGeneratorRefused = errors.New("content generator refused the request")

// IContentGenerator writes card content (Epic 3 §5.3). Implementations: Claude, fake.
type IContentGenerator interface {
	Configured() bool
	// Generate returns one result per request that the model answered; missing inputs are retried by the caller.
	Generate(ctx context.Context, words []WordRequest) ([]WordResult, entity.Usage, error)
}

// IImageProvider searches stock photos (Epic 3 §5.4). Implementations: Pixabay, Unsplash, fake.
type IImageProvider interface {
	Name() string
	Configured() bool
	// Expiring: direct links expire, so cards go through /api/v1/images/:provider/:id.
	Expiring() bool
	Search(ctx context.Context, query string, limit int) ([]entity.CardImage, error)
	// Lookup refreshes one photo by id. Returns (nil, nil) when the photo no longer exists.
	Lookup(ctx context.Context, id string) (*entity.CardImage, error)
	// TrackDownload reports that a photo was chosen for a card (required by Unsplash).
	TrackDownload(ctx context.Context, image entity.CardImage) error
}

// ImageProviders is the configured search order.
type ImageProviders []IImageProvider
