package usecases

import (
	"context"
	"net/http"

	"ielts-practice/app/modules/study/dto"
	"ielts-practice/pkg/exceptions"
)

// Undo reverts the most recent answer of the current study day and shows that card again.
func (uc *StudyUseCases) Undo(ctx context.Context, userID string, req dto.UndoRequestDto) (dto.UndoResponseDto, error) {
	sc, err := uc.loadContext(ctx, userID)
	if err != nil {
		return dto.UndoResponseDto{}, err
	}

	var cardID string
	err = uc.transaction.Execute(ctx, func(txCtx context.Context) error {
		log, err := uc.repo.LastUndoableLog(txCtx, userID, sc.todayStart)
		if err != nil {
			return internalErr("Failed to load last answer", err)
		}
		if log == nil {
			return exceptions.NewException(http.StatusNotFound, exceptions.CodeStudyNothingToUndo, "Nothing to undo", nil)
		}
		if err := uc.repo.SaveSchedule(txCtx, log.CardID, log.Before, log.IsLeechBefore); err != nil {
			return internalErr("Failed to restore card", err)
		}
		if err := uc.repo.MarkUndone(txCtx, log.ID, sc.now); err != nil {
			return internalErr("Failed to update review log", err)
		}
		cardID = log.CardID
		return nil
	})
	if err != nil {
		return dto.UndoResponseDto{}, err
	}

	restored, err := uc.repo.GetCard(ctx, userID, cardID)
	if err != nil {
		return dto.UndoResponseDto{}, internalErr("Failed to load card", err)
	}
	queue, err := uc.buildQueue(ctx, sc, userID, req.DeckID, req.Size, restored)
	if err != nil {
		return dto.UndoResponseDto{}, err
	}
	return dto.UndoResponseDto{StudyQueueDto: queue, UndoneCardID: cardID}, nil
}
