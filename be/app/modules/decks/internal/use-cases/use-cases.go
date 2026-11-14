package usecases

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/decks/domain"
	"ielts-practice/app/modules/decks/domain/entity"
	"ielts-practice/app/modules/decks/dto"
	"ielts-practice/app/modules/decks/internal/mapper"
	"ielts-practice/pkg/exceptions"
)

type DeckUseCases struct {
	decks       domain.IDeckRepository
	cards       domain.ICardRepository
	transaction *database.TransactionStruct
	mapper      *mapper.DeckMapper
	logger      zerolog.Logger
	now         func() time.Time
}

func NewDeckUseCases(
	decks domain.IDeckRepository,
	cards domain.ICardRepository,
	tx *database.TransactionStruct,
	mapper *mapper.DeckMapper,
	logger zerolog.Logger,
) *DeckUseCases {
	return &DeckUseCases{decks: decks, cards: cards, transaction: tx, mapper: mapper, logger: logger, now: time.Now}
}

func deckNotFound() *exceptions.AppError {
	return exceptions.NewException(http.StatusNotFound, exceptions.CodeDeckNotFound, "Deck not found", nil)
}

func cardNotFound() *exceptions.AppError {
	return exceptions.NewException(http.StatusNotFound, exceptions.CodeCardNotFound, "Card not found", nil)
}

func nameTaken() *exceptions.AppError {
	e := exceptions.NewException(http.StatusConflict, exceptions.CodeDeckNameTaken, "A deck with this name already exists", nil)
	e.Fields = map[string]string{"name": exceptions.CodeDeckNameTaken}
	return e
}

func internalErr(msg string, err error) *exceptions.AppError {
	return exceptions.ErrInternalServerError(msg, err)
}

// mustOwnDeck loads an active deck owned by the user.
func (uc *DeckUseCases) mustOwnDeck(ctx context.Context, userID, deckID string) (*entity.DeckEntity, error) {
	deck, err := uc.decks.FindByID(ctx, userID, deckID)
	if err != nil {
		return nil, internalErr("Failed to load deck", err)
	}
	if deck == nil {
		return nil, deckNotFound()
	}
	return deck, nil
}

func isNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// validateManualFields checks rules the binding tags can't express for optional strings.
func validateManualFields(f dto.ManualCardFields) *exceptions.AppError {
	if f.ImageURL != nil && *f.ImageURL != "" {
		u, err := url.ParseRequestURI(*f.ImageURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			e := exceptions.NewException(http.StatusBadRequest, exceptions.CodeValidationFailed, "Validation failed", nil)
			e.Fields = map[string]string{"image_url": "VALIDATION_URL"}
			return e
		}
	}
	return nil
}

func trimFields(f *dto.ManualCardFields) {
	for _, p := range []*string{f.Headword, f.MeaningVI, f.IPA, f.PartOfSpeech, f.ExampleEN, f.ExampleVI, f.ImageURL} {
		if p != nil {
			*p = strings.TrimSpace(*p)
		}
	}
}
