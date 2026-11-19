// Package in is what the library module offers to the rest of the app.
package in

import (
	"context"
	"io/fs"

	cards "ielts-practice/app/modules/decks/domain/entity"
)

type SeedResult struct {
	Topics    int // topics stored (new or changed)
	Unchanged int
	Words     int
}

// ILibrarySeeder loads seeds/library (Epic 8 §3.4) at `api seed`.
type ILibrarySeeder interface {
	SeedLibrary(ctx context.Context, fsys fs.FS, root string) (SeedResult, error)
}

// SeedCard is one reviewed card from a seed file outside the library (seeded writing vocabulary, Epic 8 §4).
type SeedCard struct {
	ImageQuery   string
	Visualizable bool
	Card         cards.CardContent
}

// ICardStore stores reviewed cards in the shared card cache (origin 'library', never rewritten by the AI), so
// other modules can offer them for adding to a deck through POST /library/add.
type ICardStore interface {
	// StoreCards validates and stores the cards; it returns their lexeme ids in order.
	StoreCards(ctx context.Context, cards []SeedCard) ([]string, error)
	// FindCard returns the id of the library card for a headword ("" when there is none).
	FindCard(ctx context.Context, headword string) (string, error)
}

// ITopicWords lists the library's words of a topic (Epic 4 topic id), lower case, so topic suggestions can skip
// them: the learner adds those from the library, and the AI suggests the words the library does not have.
type ITopicWords interface {
	TopicWords(ctx context.Context, topicID string) ([]string, error)
}
