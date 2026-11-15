package in

import (
	"context"

	"ielts-practice/app/modules/decks/domain/entity"
)

// GeneratedCard is one finished card coming from another module (Epic 3 generation).
type GeneratedCard struct {
	LexemeID *string
	Content  entity.CardContent
}

// SaveGeneratedInput targets an existing deck (DeckID) or creates one (NewDeckName).
type SaveGeneratedInput struct {
	UserID      string
	DeckID      string
	NewDeckName string
	Cards       []GeneratedCard
}

type SaveGeneratedResult struct {
	DeckID   string
	DeckName string
	Created  int
	// Skipped lists headwords already in the deck.
	Skipped []string
}

type IDeckService interface {
	// SaveGeneratedCards creates new cards in one transaction, skipping words already in the deck.
	SaveGeneratedCards(ctx context.Context, input SaveGeneratedInput) (SaveGeneratedResult, error)
	// DeckWords returns the lowercase headwords and lemmas of the cards in a deck, newest first.
	DeckWords(ctx context.Context, userID, deckID string) ([]string, error)
	// OwnsDeck reports whether the user has an active deck with this id.
	OwnsDeck(ctx context.Context, userID, deckID string) (bool, error)
}
