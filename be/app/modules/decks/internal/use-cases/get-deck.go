package usecases

import (
	"context"

	"ielts-practice/app/modules/decks/dto"
)

func (uc *DeckUseCases) GetDeck(ctx context.Context, userID, deckID string) (dto.DeckDto, error) {
	deck, err := uc.mustOwnDeck(ctx, userID, deckID)
	if err != nil {
		return dto.DeckDto{}, err
	}
	return uc.mapper.DeckEntityToDto(deck), nil
}
