package usecases

import (
	"context"
	"net/http"

	"ielts-practice/app/modules/decks/domain/entity"
	"ielts-practice/app/modules/decks/dto"
	"ielts-practice/pkg/exceptions"
)

func duplicateHeadword() *exceptions.AppError {
	e := exceptions.NewException(http.StatusConflict, exceptions.CodeCardDuplicate, "This word is already in the deck", nil)
	e.Fields = map[string]string{"headword": exceptions.CodeCardDuplicate}
	return e
}

func (uc *DeckUseCases) CreateCard(ctx context.Context, userID, deckID string, req dto.CreateCardRequestDto) (dto.CardDto, error) {
	trimFields(&req.ManualCardFields)
	if err := validateManualFields(req.ManualCardFields); err != nil {
		return dto.CardDto{}, err
	}
	missing := map[string]string{}
	if req.Headword == nil || *req.Headword == "" {
		missing["headword"] = "VALIDATION_REQUIRED"
	}
	if req.MeaningVI == nil || *req.MeaningVI == "" {
		missing["meaning_vi"] = "VALIDATION_REQUIRED"
	}
	if len(missing) > 0 {
		e := exceptions.NewException(http.StatusBadRequest, exceptions.CodeValidationFailed, "Validation failed", nil)
		e.Fields = missing
		return dto.CardDto{}, e
	}

	var created *entity.CardEntity
	err := uc.transaction.Execute(ctx, func(txCtx context.Context) error {
		if _, err := uc.mustOwnDeck(txCtx, userID, deckID); err != nil {
			return err
		}
		if !req.AllowDuplicate {
			exists, err := uc.cards.HeadwordExists(txCtx, deckID, *req.Headword, "")
			if err != nil {
				return internalErr("Failed to check duplicates", err)
			}
			if exists {
				return duplicateHeadword()
			}
		}
		position, err := uc.cards.NextPosition(txCtx, deckID)
		if err != nil {
			return internalErr("Failed to compute position", err)
		}
		content := entity.CardContent{}
		uc.mapper.ApplyManualFields(&content, req.ManualCardFields)
		created, err = uc.cards.Create(txCtx, &entity.CardEntity{
			UserID:   userID,
			DeckID:   deckID,
			Headword: content.Headword,
			Content:  content,
			Due:      uc.now(),
			Position: position,
		})
		if err != nil {
			return internalErr("Failed to create card", err)
		}
		return nil
	})
	if err != nil {
		return dto.CardDto{}, err
	}
	return uc.mapper.CardEntityToDto(created), nil
}
