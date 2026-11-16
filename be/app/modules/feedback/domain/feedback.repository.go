package domain

import (
	"context"
	"time"

	"ielts-practice/app/modules/feedback/domain/entity"
)

type IFeedbackRepository interface {
	// Create inserts the feedback and fills ID, CreatedAt and UpdatedAt.
	Create(ctx context.Context, fb *entity.Feedback) error
	CountSince(ctx context.Context, userID string, since time.Time) (int, error)
	MarkNotified(ctx context.Context, id string, at time.Time) error

	// List returns one page, newest first, with the author joined.
	List(ctx context.Context, f entity.FeedbackFilter) ([]entity.Feedback, int64, error)
	// CountByStatus counts every status; category "" means all categories.
	CountByStatus(ctx context.Context, category string) (entity.StatusCounts, error)
	// FindByID returns nil when there is no such feedback.
	FindByID(ctx context.Context, id string) (*entity.Feedback, error)
	UpdateReview(ctx context.Context, id, status, adminNote string, resolvedAt *time.Time) error
}
