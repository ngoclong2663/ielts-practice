package usecases

import (
	"context"

	"ielts-practice/app/modules/decks/dto"
)

func (uc *DeckUseCases) ListDecks(ctx context.Context, userID string) (dto.DeckListDto, error) {
	decks, err := uc.decks.List(ctx, userID)
	if err != nil {
		return dto.DeckListDto{}, internalErr("Failed to list decks", err)
	}
	out := dto.DeckListDto{Items: make([]dto.DeckDto, 0, len(decks))}
	for i := range decks {
		out.Items = append(out.Items, uc.mapper.DeckEntityToDto(&decks[i]))
	}
	return out, nil
}
