package domain

import (
	"context"
	"time"

	"ielts-practice/app/modules/decks/domain/entity"
)

// Find* methods return (nil, nil) when nothing matches. All queries are scoped to userID.
type IDeckRepository interface {
	List(ctx context.Context, userID string) ([]entity.DeckEntity, error)
	FindByID(ctx context.Context, userID, deckID string) (*entity.DeckEntity, error)
	NameTaken(ctx context.Context, userID, name, exceptDeckID string) (bool, error)
	Create(ctx context.Context, deck *entity.DeckEntity) (*entity.DeckEntity, error)
	Update(ctx context.Context, deck *entity.DeckEntity) error
	Archive(ctx context.Context, userID, deckID string, at time.Time) error
}

type BulkAction string

const (
	BulkSuspend   BulkAction = "suspend"
	BulkUnsuspend BulkAction = "unsuspend"
	BulkReset     BulkAction = "reset"
	BulkDelete    BulkAction = "delete"
	BulkMove      BulkAction = "move"
)

type ICardRepository interface {
	List(ctx context.Context, userID string, filter entity.CardFilter) ([]entity.CardEntity, int64, error)
	FindByID(ctx context.Context, userID, cardID string) (*entity.CardEntity, error)
	HeadwordExists(ctx context.Context, deckID, headword, exceptCardID string) (bool, error)
	NextPosition(ctx context.Context, deckID string) (int, error)
	Create(ctx context.Context, card *entity.CardEntity) (*entity.CardEntity, error)
	UpdateContent(ctx context.Context, card *entity.CardEntity) error
	// Bulk applies an action to the user's cards and returns how many rows changed.
	Bulk(ctx context.Context, userID string, cardIDs []string, action BulkAction, targetDeckID string, now time.Time) (int64, error)
}
