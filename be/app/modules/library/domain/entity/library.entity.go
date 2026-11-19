// Package entity holds the vocabulary library (Epic 8 §3).
//
// Entities (have an identity): Collection, Word. Value objects: Localized, TopicCount, SeedWord.
package entity

import cards "ielts-practice/app/modules/decks/domain/entity"

type CardContent = cards.CardContent

const (
	StatusPublished = "published"
	StatusHidden    = "hidden"

	// Lexeme origins: a library card is reviewed by hand and never rewritten by the AI.
	OriginLibrary = "library"

	SourceIELTS4000 = "ielts4000" // headword from the IELTS 4000 Academic Word List
	SourceCore      = "core"      // essential topic word added by us
)

// Kinds of collection; only IELTS has content for now.
var Kinds = []string{"ielts", "toeic"}

// Sources a library word may come from.
var Sources = []string{SourceIELTS4000, SourceCore}

// Levels a word may be filtered by.
var Levels = []string{"A1", "A2", "B1", "B2", "C1", "C2"}

type Localized struct {
	VI string `json:"vi"`
	EN string `json:"en"`
}

// TopicCount is one topic of a collection and how many words it has.
type TopicCount struct {
	TopicID string
	Words   int
}

type Collection struct {
	ID          string
	Slug        string
	Kind        string
	Title       Localized
	Description Localized
	Sort        int
	Status      string
	ContentHash string
	Topics      []TopicCount
}

// Word is a library word with its card, as a learner sees it.
type Word struct {
	ID       string
	TopicID  string
	LexemeID string
	CEFR     string
	Source   string
	Card     CardContent
	InDeck   bool // already a card in one of the learner's decks
	Known    bool // marked as known (Epic 4)
}

// WordFilter narrows a collection's words. An empty field does not filter.
type WordFilter struct {
	TopicID string
	Level   string
	Query   string // folded with rules.Fold
	Limit   int
}

// SeedWord is one word of a topic seed file, ready to store.
type SeedWord struct {
	Key          string // lexeme cache key (vocab rules.LexemeKey)
	Card         CardContent
	CEFR         string
	Source       string
	ImageQuery   string
	SearchText   string
	Sort         int
	PixabayImage bool // the card's image is a Pixabay photo behind /api/v1/images (its link must be registered)
}
