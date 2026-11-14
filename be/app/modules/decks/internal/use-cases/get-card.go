package usecases

import (
	"context"

	"ielts-practice/app/modules/decks/dto"
)

func (uc *DeckUseCases) GetCard(ctx context.Context, userID, cardID string) (dto.CardDto, error) {
	card, err := uc.cards.FindByID(ctx, userID, cardID)
	if err != nil {
		return dto.CardDto{}, internalErr("Failed to load card", err)
	}
	if card == nil {
		return dto.CardDto{}, cardNotFound()
	}
	return uc.mapper.CardEntityToDto(card), nil
}
