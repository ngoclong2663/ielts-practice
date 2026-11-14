package usecases

import (
	"context"
	"strings"

	"ielts-practice/app/modules/decks/dto"
)

func (uc *DeckUseCases) UpdateDeck(ctx context.Context, userID, deckID string, req dto.UpdateDeckRequestDto) (dto.DeckDto, error) {
	deck, err := uc.mustOwnDeck(ctx, userID, deckID)
	if err != nil {
		return dto.DeckDto{}, err
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		taken, err := uc.decks.NameTaken(ctx, userID, name, deckID)
		if err != nil {
			return dto.DeckDto{}, internalErr("Failed to check deck name", err)
		}
		if taken {
			return dto.DeckDto{}, nameTaken()
		}
		deck.Name = name
	}
	if req.Description != nil {
		deck.Description = strings.TrimSpace(*req.Description)
	}
	if req.UseDefaultLimits {
		deck.NewPerDay, deck.ReviewsPerDay = nil, nil
	} else {
		if req.NewPerDay != nil {
			deck.NewPerDay = req.NewPerDay
		}
		if req.ReviewsPerDay != nil {
			deck.ReviewsPerDay = req.ReviewsPerDay
		}
	}
	if err := uc.decks.Update(ctx, deck); err != nil {
		return dto.DeckDto{}, internalErr("Failed to update deck", err)
	}
	return uc.mapper.DeckEntityToDto(deck), nil
}
