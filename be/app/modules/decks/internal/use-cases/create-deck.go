package usecases

import (
	"context"
	"net/http"
	"strings"

	"ielts-practice/app/modules/decks/domain/entity"
	"ielts-practice/app/modules/decks/dto"
	"ielts-practice/pkg/exceptions"
)

func (uc *DeckUseCases) CreateDeck(ctx context.Context, userID string, req dto.CreateDeckRequestDto) (dto.DeckDto, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		e := exceptions.NewException(http.StatusBadRequest, exceptions.CodeValidationFailed, "Name is required", nil)
		e.Fields = map[string]string{"name": "VALIDATION_REQUIRED"}
		return dto.DeckDto{}, e
	}
	taken, err := uc.decks.NameTaken(ctx, userID, name, "")
	if err != nil {
		return dto.DeckDto{}, internalErr("Failed to check deck name", err)
	}
	if taken {
		return dto.DeckDto{}, nameTaken()
	}
	deck, err := uc.decks.Create(ctx, &entity.DeckEntity{
		UserID:      userID,
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		Source:      "manual",
	})
	if err != nil {
		return dto.DeckDto{}, internalErr("Failed to create deck", err)
	}
	return uc.mapper.DeckEntityToDto(deck), nil
}
