package usecases

import (
	"context"
	"net/http"
	"strings"

	"ielts-practice/app/modules/feedback/domain/entity"
	"ielts-practice/app/modules/feedback/domain/rules"
	"ielts-practice/app/modules/feedback/dto"
	"ielts-practice/pkg/exceptions"
)

// SendFeedback saves a learner's feedback (Epic 6 §3) and notifies the owner in the background.
func (uc *FeedbackUseCases) SendFeedback(ctx context.Context, userID, userAgent string, req dto.CreateFeedbackRequestDto) (dto.CreatedFeedbackDto, error) {
	msg, code := rules.NormalizeMessage(req.Message)
	if code != "" {
		e := exceptions.NewException(http.StatusBadRequest, exceptions.CodeValidationFailed, "Validation failed", nil).
			WithParams(map[string]any{"min": rules.MinMessageRunes, "max": rules.MaxMessageRunes})
		e.Fields = map[string]string{"message": code}
		return dto.CreatedFeedbackDto{}, e
	}

	limit := uc.cfg.Feedback.DailyLimit
	used, err := uc.repo.CountSince(ctx, userID, uc.now().Add(-limitWindow))
	if err != nil {
		return dto.CreatedFeedbackDto{}, internalErr("Failed to load feedback limit", err)
	}
	if used >= limit {
		return dto.CreatedFeedbackDto{}, exceptions.NewException(http.StatusTooManyRequests, exceptions.CodeFeedbackLimit, "Daily feedback limit reached", nil).
			WithParams(map[string]any{"limit": limit})
	}

	fb := entity.Feedback{
		UserID:    userID,
		Category:  req.Category,
		Message:   msg,
		Page:      rules.Truncate(strings.TrimSpace(req.Page), rules.MaxPageRunes),
		UILocale:  req.UILocale,
		UserAgent: rules.Truncate(strings.TrimSpace(userAgent), rules.MaxAgentRunes),
	}
	if err := uc.repo.Create(ctx, &fb); err != nil {
		return dto.CreatedFeedbackDto{}, internalErr("Failed to save feedback", err)
	}
	uc.notifyInBackground(fb)
	return dto.CreatedFeedbackDto{ID: fb.ID, CreatedAt: fb.CreatedAt}, nil
}
