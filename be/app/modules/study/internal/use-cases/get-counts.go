package usecases

import (
	"context"

	"ielts-practice/app/modules/study/dto"
)

// GetCounts returns today's new · learning · review counts per deck and in total.
func (uc *StudyUseCases) GetCounts(ctx context.Context, userID string) (dto.CountsResponseDto, error) {
	sc, err := uc.loadContext(ctx, userID)
	if err != nil {
		return dto.CountsResponseDto{}, err
	}
	lim, err := uc.computeLimits(ctx, sc, userID, "")
	if err != nil {
		return dto.CountsResponseDto{}, err
	}
	decks := lim.perDeck
	if decks == nil {
		decks = []dto.DeckCountsDto{}
	}
	return dto.CountsResponseDto{Decks: decks, Total: lim.counts, DayEndsAt: sc.tomorrowStart}, nil
}
