package domain

import (
	"context"

	"ielts-practice/app/modules/guides/domain/entity"
)

type IGuidesRepository interface {
	// ListPublished returns published guides in display order; an empty topic means every topic.
	ListPublished(ctx context.Context, topic string) ([]entity.GuideVideo, error)
	FindBySlug(ctx context.Context, slug string) (*entity.GuideVideo, error)
	// Upsert inserts by slug or updates an existing guide in place (same id).
	Upsert(ctx context.Context, g *entity.GuideVideo) error
	// HasMedia reports whether a published guide uses this object key (video or poster).
	HasMedia(ctx context.Context, key string) (bool, error)
}
