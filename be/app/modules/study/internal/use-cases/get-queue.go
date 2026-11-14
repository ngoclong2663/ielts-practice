package usecases

import (
	"context"

	"ielts-practice/app/modules/study/dto"
)

// GetQueue returns the next cards to study (a deck, or every deck when deck_id is empty).
func (uc *StudyUseCases) GetQueue(ctx context.Context, userID string, q dto.QueueQueryDto) (dto.StudyQueueDto, error) {
	sc, err := uc.loadContext(ctx, userID)
	if err != nil {
		return dto.StudyQueueDto{}, err
	}
	return uc.buildQueue(ctx, sc, userID, q.DeckID, q.Size, nil)
}
