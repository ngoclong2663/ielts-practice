package usecases

import (
	"context"
	"net/http"

	"ielts-practice/app/modules/decks/domain"
	"ielts-practice/app/modules/decks/dto"
	"ielts-practice/pkg/exceptions"
)

func (uc *DeckUseCases) BulkCards(ctx context.Context, userID string, req dto.BulkCardsRequestDto) (dto.BulkCardsResponseDto, error) {
	action := domain.BulkAction(req.Action)
	if action == domain.BulkMove {
		if req.DeckID == "" {
			e := exceptions.NewException(http.StatusBadRequest, exceptions.CodeValidationFailed, "deck_id is required for move", nil)
			e.Fields = map[string]string{"deck_id": "VALIDATION_REQUIRED"}
			return dto.BulkCardsResponseDto{}, e
		}
		if _, err := uc.mustOwnDeck(ctx, userID, req.DeckID); err != nil {
			return dto.BulkCardsResponseDto{}, err
		}
	}
	affected, err := uc.cards.Bulk(ctx, userID, req.CardIDs, action, req.DeckID, uc.now())
	if err != nil {
		return dto.BulkCardsResponseDto{}, internalErr("Failed to update cards", err)
	}
	return dto.BulkCardsResponseDto{Affected: affected}, nil
}
