package usecases

import (
	"net/http"

	"github.com/rs/zerolog"

	decksin "ielts-practice/app/modules/decks/ports/in"
	"ielts-practice/app/modules/library/domain"
	"ielts-practice/app/modules/library/domain/entity"
	"ielts-practice/app/modules/library/dto"
	"ielts-practice/pkg/exceptions"
)

// LibraryUseCases implements Epic 8 §3: browsing the library and adding its cards to a deck.
type LibraryUseCases struct {
	repo   domain.ILibraryRepository
	decks  decksin.IDeckService
	logger zerolog.Logger
}

func NewLibraryUseCases(repo domain.ILibraryRepository, decks decksin.IDeckService, logger zerolog.Logger) *LibraryUseCases {
	return &LibraryUseCases{repo: repo, decks: decks, logger: logger.With().Str("module", "library").Logger()}
}

func appErr(status int, code, msg string) *exceptions.AppError {
	return exceptions.NewException(status, code, msg, nil)
}

func internalErr(msg string, err error) *exceptions.AppError {
	return exceptions.ErrInternalServerError(msg, err)
}

func notFound() *exceptions.AppError {
	return appErr(http.StatusNotFound, exceptions.CodeLibraryNotFound, "Library collection not found")
}

func toCollectionDto(c *entity.Collection) dto.CollectionDto {
	out := dto.CollectionDto{
		Slug: c.Slug, Kind: c.Kind,
		Title:       dto.LocalizedDto{VI: c.Title.VI, EN: c.Title.EN},
		Description: dto.LocalizedDto{VI: c.Description.VI, EN: c.Description.EN},
		Topics:      make([]dto.TopicCountDto, 0, len(c.Topics)),
	}
	for _, t := range c.Topics {
		out.Words += t.Words
		out.Topics = append(out.Topics, dto.TopicCountDto{ID: t.TopicID, Words: t.Words})
	}
	return out
}
