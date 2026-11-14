package usecases

import (
	"context"
	"errors"
	"net/http"

	"gorm.io/gorm"

	"ielts-practice/app/modules/study/domain"
	"ielts-practice/app/modules/study/domain/entity"
	"ielts-practice/app/modules/study/dto"
	"ielts-practice/pkg/exceptions"
)

// Answer grades a card (Again/Hard/Good/Easy), saves the new FSRS state and a review log,
// then returns the next cards. Repeating the same client_review_id is a no-op (idempotent).
func (uc *StudyUseCases) Answer(ctx context.Context, userID string, req dto.AnswerRequestDto) (dto.AnswerResponseDto, error) {
	sc, err := uc.loadContext(ctx, userID)
	if err != nil {
		return dto.AnswerResponseDto{}, err
	}

	var answered *dto.AnsweredDto
	err = uc.transaction.Execute(ctx, func(txCtx context.Context) error {
		exists, err := uc.repo.ReviewExists(txCtx, userID, req.ClientReviewID)
		if err != nil {
			return internalErr("Failed to check review", err)
		}
		if exists {
			return nil // already applied (retry or double tap)
		}

		card, err := uc.repo.LockCard(txCtx, userID, req.CardID)
		if err != nil {
			return internalErr("Failed to load card", err)
		}
		if card == nil {
			return exceptions.NewException(http.StatusNotFound, exceptions.CodeCardNotFound, "Card not found", nil)
		}
		if card.Suspended || (req.ExpectedReps != nil && *req.ExpectedReps != card.Reps) {
			return stateChanged()
		}

		rating := entity.Rating(req.Rating)
		before := card.SchedulingState
		after, elapsed, err := sc.scheduler.Answer(before, rating, sc.now)
		if err != nil {
			return internalErr("Failed to schedule card", err)
		}
		becameLeech := !card.IsLeech && domain.BecomesLeech(before, after)
		if err := uc.repo.SaveSchedule(txCtx, card.ID, after, card.IsLeech || becameLeech); err != nil {
			return internalErr("Failed to save card", err)
		}
		if err := uc.repo.InsertLog(txCtx, &entity.ReviewLogEntity{
			UserID:         userID,
			CardID:         card.ID,
			DeckID:         card.DeckID,
			ClientReviewID: req.ClientReviewID,
			Rating:         rating,
			Before:         before,
			IsLeechBefore:  card.IsLeech,
			After:          after,
			ElapsedDays:    elapsed,
			DurationMs:     min(req.DurationMs, maxAnswerDuration),
			ReviewedAt:     sc.now,
		}); err != nil {
			return err
		}
		answered = &dto.AnsweredDto{
			CardID: card.ID, Rating: req.Rating, State: stateName(after.State), Due: after.Due, BecameLeech: becameLeech,
		}
		return nil
	})
	// Two identical requests racing: the unique client_review_id makes the second one a no-op.
	if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
		return dto.AnswerResponseDto{}, err
	}

	queue, err := uc.buildQueue(ctx, sc, userID, req.DeckID, req.Size, nil)
	if err != nil {
		return dto.AnswerResponseDto{}, err
	}
	return dto.AnswerResponseDto{StudyQueueDto: queue, Answered: answered}, nil
}
