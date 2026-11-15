package usecases

import (
	"context"
	"math"
	"net/http"
	"strings"

	"ielts-practice/app/modules/writing/domain"
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
	if attempt.Status != domain.StatusInProgress {
		return dto.AttemptDto{}, appErr(http.StatusConflict, exceptions.CodeWritingSubmitted, "This attempt is already submitted")
	}
	text := strings.TrimSpace(strings.ReplaceAll(req.Text, "\r\n", "\n"))
	words := domain.CountWords(text)
	if words < domain.MinSubmitWords {
		e := appErr(http.StatusBadRequest, exceptions.CodeWritingTooShort, "The answer is too short to review")
		e.Params = map[string]any{"min": domain.MinSubmitWords}
		return dto.AttemptDto{}, e
	}
	if err := uc.checkLimit(ctx, userID, domain.KindSubmit); err != nil {
		return dto.AttemptDto{}, err
	}

	rctx, cancel := context.WithTimeout(ctx, reviewTimeout)
	defer cancel()
	feedback, usage, err := uc.coach.ReviewEssay(rctx, domain.ReviewRequest{Task: task, Position: attempt.Position, Text: text, WordCount: words})
	if err != nil {
		return dto.AttemptDto{}, coachErr(err)
	}
	uc.record(ctx, userID, &attempt.ID, domain.KindSubmit, usage)

	low, high := BandRange(feedback.Criteria)
	saved, err := uc.repo.SaveSubmission(ctx, attempt.ID, &entity.Submission{
		Text: text, WordCount: words, SecondsSpent: req.SecondsSpent, Feedback: feedback, BandLow: low, BandHigh: high,
	})
	if err != nil {
		return dto.AttemptDto{}, internalErr("Failed to save the answer", err)
	}
	if !saved {
		return dto.AttemptDto{}, appErr(http.StatusConflict, exceptions.CodeWritingSubmitted, "This attempt is already submitted")
	}
	attempt.Status = domain.StatusSubmitted
	now := uc.now()
	attempt.SubmittedAt = &now
	return uc.attemptView(ctx, attempt, task, lesson)
}

// BandRange turns the four criterion bands into the range shown to the learner. The overall band is the
// mean rounded the IELTS way (to the nearest half band, .25 and .75 rounding up); AI marking is only
// roughly as reliable as an examiner, so the range is that band and the half band below it.
func BandRange(criteria []entity.CriterionFeedback) (float64, float64) {
	if len(criteria) == 0 {
		return 0, 0
	}
	sum := 0.0
	for _, c := range criteria {
		sum += c.Band
	}
	overall := math.Floor(sum/float64(len(criteria))*2+0.5) / 2
	return max(0, overall-0.5), overall
}
