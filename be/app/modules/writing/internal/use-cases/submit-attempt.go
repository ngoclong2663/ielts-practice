package usecases

import (
	"context"
	"net/http"
	"strings"

	"ielts-practice/app/modules/writing/domain/rules"

	"ielts-practice/app/modules/writing/domain/entity"
	"ielts-practice/app/modules/writing/dto"
	"ielts-practice/pkg/exceptions"
)

// SubmitAttempt reviews the full answer (one AI call) and closes the attempt.
func (uc *WritingUseCases) SubmitAttempt(ctx context.Context, userID, attemptID string, req dto.SubmitRequestDto) (dto.AttemptDto, error) {
	attempt, task, lesson, err := uc.loadAttempt(ctx, userID, attemptID)
	if err != nil {
		return dto.AttemptDto{}, err
	}
	if attempt.Status != entity.StatusInProgress {
		return dto.AttemptDto{}, appErr(http.StatusConflict, exceptions.CodeWritingSubmitted, "This attempt is already submitted")
	}
	text := strings.TrimSpace(strings.ReplaceAll(req.Text, "\r\n", "\n"))
	words := rules.CountWords(text)
	if words < rules.MinSubmitWords {
		e := appErr(http.StatusBadRequest, exceptions.CodeWritingTooShort, "The answer is too short to review")
		e.Params = map[string]any{"min": rules.MinSubmitWords}
		return dto.AttemptDto{}, e
	}
	if err := uc.checkLimit(ctx, userID, entity.KindSubmit); err != nil {
		return dto.AttemptDto{}, err
	}

	rctx, cancel := context.WithTimeout(ctx, reviewTimeout)
	defer cancel()
	feedback, usage, err := uc.coach.ReviewEssay(rctx, entity.ReviewRequest{Task: task, Position: attempt.Position, Text: text, WordCount: words})
	if err != nil {
		return dto.AttemptDto{}, coachErr(err)
	}
	uc.record(ctx, userID, &attempt.ID, entity.KindSubmit, usage)

	low, high := rules.BandRange(feedback.Criteria)
	saved, err := uc.repo.SaveSubmission(ctx, attempt.ID, &entity.Submission{
		Text: text, WordCount: words, SecondsSpent: req.SecondsSpent, Feedback: feedback, BandLow: low, BandHigh: high,
	})
	if err != nil {
		return dto.AttemptDto{}, internalErr("Failed to save the answer", err)
	}
	if !saved {
		return dto.AttemptDto{}, appErr(http.StatusConflict, exceptions.CodeWritingSubmitted, "This attempt is already submitted")
	}
	attempt.Status = entity.StatusSubmitted
	now := uc.now()
	attempt.SubmittedAt = &now
	return uc.attemptView(ctx, attempt, task, lesson)
}
