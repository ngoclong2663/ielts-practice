package usecases

import (
	"context"
	"net/http"

	"ielts-practice/app/modules/decks/dto"
	"ielts-practice/pkg/exceptions"
)

func (uc *DeckUseCases) UpdateCard(ctx context.Context, userID, cardID string, req dto.UpdateCardRequestDto) (dto.CardDto, error) {
	trimFields(&req.ManualCardFields)
	if err := validateManualFields(req.ManualCardFields); err != nil {
		return dto.CardDto{}, err
	}
	if (req.Headword != nil && *req.Headword == "") || (req.MeaningVI != nil && *req.MeaningVI == "") {
		e := exceptions.NewException(http.StatusBadRequest, exceptions.CodeValidationFailed, "Validation failed", nil)
		e.Fields = map[string]string{}
		if req.Headword != nil && *req.Headword == "" {
			e.Fields["headword"] = "VALIDATION_REQUIRED"
		}
		if req.MeaningVI != nil && *req.MeaningVI == "" {
			e.Fields["meaning_vi"] = "VALIDATION_REQUIRED"
		}
		return dto.CardDto{}, e
	}
	card, err := uc.cards.FindByID(ctx, userID, cardID)
	if err != nil {
		return dto.CardDto{}, internalErr("Failed to load card", err)
	}
	if card == nil {
		return dto.CardDto{}, cardNotFound()
	}
	uc.mapper.ApplyManualFields(&card.Content, req.ManualCardFields)
	card.Headword = card.Content.Headword
	if err := uc.cards.UpdateContent(ctx, card); err != nil {
		return dto.CardDto{}, internalErr("Failed to update card", err)
	}
	return uc.mapper.CardEntityToDto(card), nil
}
