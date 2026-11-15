package out

import (
	"context"

	"ielts-practice/app/modules/vocab/domain/entity"
)

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
