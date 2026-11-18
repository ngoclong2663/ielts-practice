package usecases

import (
	"context"
	"strings"

	"ielts-practice/app/modules/guides/domain/entity"
	"ielts-practice/app/modules/guides/dto"
	"ielts-practice/pkg/exceptions"
)

// MediaPath is where the API serves guide media (a redirect to a short-lived S3 link, resolve-media.go).
const MediaPath = "/api/v1/guides/media/"

func (uc *GuidesUseCases) ListGuides(ctx context.Context, topic string) ([]dto.GuideDto, error) {
	guides, err := uc.repo.ListPublished(ctx, topic)
	if err != nil {
		return nil, exceptions.ErrInternalServerError("Failed to load guides", err)
	}
	out := make([]dto.GuideDto, 0, len(guides))
	for _, g := range guides {
		d := dto.GuideDto{
			Slug: g.Slug, Topic: g.Topic, Lang: g.Lang, Title: g.Title, Description: g.Description,
			DurationSeconds: g.DurationSeconds, VideoURL: MediaPath + strings.TrimPrefix(g.VideoKey, entity.VideoKeyPrefix),
		}
		if g.PosterKey != "" {
			d.PosterURL = MediaPath + strings.TrimPrefix(g.PosterKey, entity.PosterKeyPrefix)
		}
		out = append(out, d)
	}
	return out, nil
}
