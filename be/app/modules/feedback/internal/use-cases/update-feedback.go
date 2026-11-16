package usecases

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"ielts-practice/app/modules/feedback/domain/rules"
	"ielts-practice/app/modules/feedback/dto"
)

// UpdateFeedback changes the status (new → seen → resolved, or reopen) and the private note.
func (uc *FeedbackUseCases) UpdateFeedback(ctx context.Context, id string, req dto.UpdateFeedbackRequestDto) (dto.FeedbackItemDto, error) {
	if _, err := uuid.Parse(id); err != nil {
		return dto.FeedbackItemDto{}, notFound()
	}
	fb, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return dto.FeedbackItemDto{}, internalErr("Failed to load feedback", err)
	}
	if fb == nil {
		return dto.FeedbackItemDto{}, notFound()
	}
	if req.Status == nil && req.AdminNote == nil {
		return toItemDto(fb), nil
	}

	status, note := fb.Status, fb.AdminNote
	if req.Status != nil {
		status = *req.Status
	}
	if req.AdminNote != nil {
		note = strings.TrimSpace(*req.AdminNote)
	}
	resolvedAt := rules.ResolvedAt(fb.ResolvedAt, fb.Status, status, uc.now())
	if err := uc.repo.UpdateReview(ctx, id, status, note, resolvedAt); err != nil {
		return dto.FeedbackItemDto{}, internalErr("Failed to update feedback", err)
	}
	fb, err = uc.repo.FindByID(ctx, id)
	if err != nil || fb == nil {
		return dto.FeedbackItemDto{}, internalErr("Failed to load feedback", err)
	}
	return toItemDto(fb), nil
}
