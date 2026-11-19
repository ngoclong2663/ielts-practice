// Package entity holds the vocabulary library (Epic 8 §3).
//
// Entities (have an identity): Collection, Word. Value objects: Localized, TopicCount, SeedWord.
package entity

import (
	"time"

	cards "ielts-practice/app/modules/decks/domain/entity"
)

type CardContent = cards.CardContent

const (
	StatusPublished = "published"
	StatusHidden    = "hidden"

	// Lexeme origins: a library card is reviewed by hand and never rewritten by the AI.
	OriginLibrary = "library"

	SourceIELTS4000 = "ielts4000" // headword from the IELTS 4000 Academic Word List
	SourceCore      = "core"      // essential topic word added by us
	SourceLearner   = "learner"   // an AI card learners kept, approved by the owner (Epic 9)

	// OtherTopic holds approved words that fit none of the preset topics.
	OtherTopic = "other"

	DecisionApproved = "approved"
	DecisionRejected = "rejected"
)

// Kinds of collection; only IELTS has content for now.
var Kinds = []string{"ielts", "toeic"}

// Sources a library word may come from.
var Sources = []string{SourceIELTS4000, SourceCore, SourceLearner}

// SeedSources are the sources a seed file manages; words from other sources survive a re-seed.
var SeedSources = []string{SourceIELTS4000, SourceCore}

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

// Candidate is an AI-made card several learners keep in a deck, offered for the library (Epic 9).
type Candidate struct {
	LexemeID    string
	Card        CardContent
	Users       int // learners who keep it in a deck
	ReportCount int
	LastAddedAt time.Time
	Topic       string // tagged by the card generator; "" for cards made before Epic 9
}

// Approval puts one candidate into a topic of a collection.
type Approval struct {
	LexemeID string
	TopicID  string
	Word     SeedWord // card, level and search text, from the stored card
}
