package dto

import (
	"time"

	"ielts-practice/app/modules/library/domain/entity"
)

type LocalizedDto struct {
	VI string `json:"vi"`
	EN string `json:"en"`
}

type TopicCountDto struct {
	ID    string `json:"id"`
	Words int    `json:"words"`
}

type CollectionDto struct {
	Slug        string          `json:"slug"`
	Kind        string          `json:"kind"`
	Title       LocalizedDto    `json:"title"`
	Description LocalizedDto    `json:"description"`
	Words       int             `json:"words"`
	Topics      []TopicCountDto `json:"topics"`
}

type CollectionsResponseDto struct {
	Collections []CollectionDto `json:"collections"`
}

type ListWordsQueryDto struct {
	Topic string `form:"topic" binding:"omitempty,max=32"`
	Level string `form:"level" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Q     string `form:"q" binding:"max=60"`
}

type WordDto struct {
	ID       string             `json:"id"`
	Topic    string             `json:"topic"`
	LexemeID string             `json:"lexeme_id"`
	CEFR     string             `json:"cefr"`
	Source   string             `json:"source"`
	Card     entity.CardContent `json:"card"`
	InDeck   bool               `json:"in_deck"`
	Known    bool               `json:"known"`
}

type WordsResponseDto struct {
	Collection CollectionDto `json:"collection"`
	Words      []WordDto     `json:"words"`
	// Truncated: more words match than were returned (a search across topics); narrow the search.
	Truncated bool `json:"truncated"`
}

// AddWordsRequestDto adds library cards (or seeded writing vocabulary) to a deck, without the AI.
type AddWordsRequestDto struct {
	LexemeIDs   []string `json:"lexeme_ids" binding:"required,min=1,max=100,dive,uuid"`
	DeckID      string   `json:"deck_id" binding:"omitempty,uuid"`
	NewDeckName string   `json:"new_deck_name" binding:"omitempty,max=80"`
}

type AddWordsResponseDto struct {
	DeckID   string   `json:"deck_id"`
	DeckName string   `json:"deck_name"`
	Created  int      `json:"created"`
	Skipped  []string `json:"skipped"`
}

// ---- candidates (Epic 9, admin) ----

type CandidateDto struct {
	LexemeID    string             `json:"lexeme_id"`
	Card        entity.CardContent `json:"card"`
	Users       int                `json:"users"`
	ReportCount int                `json:"report_count"`
	LastAddedAt time.Time          `json:"last_added_at"`
	// SuggestedTopic: the topic whose words the card shares most ("other" when none fits); the owner may change it.
	SuggestedTopic string `json:"suggested_topic"`
}

type CandidateCollectionDto struct {
	Slug   string       `json:"slug"`
	Title  LocalizedDto `json:"title"`
	Topics []string     `json:"topics"`
}

type CandidatesResponseDto struct {
	MinUsers    int                      `json:"min_users"`
	Topics      []string                 `json:"topics"` // topics a word may be approved into (Epic 4 ids and "other")
	Collections []CandidateCollectionDto `json:"collections"`
	Items       []CandidateDto           `json:"items"`
}

type ApproveItemDto struct {
	LexemeID string `json:"lexeme_id" binding:"required,uuid"`
	TopicID  string `json:"topic_id" binding:"required,max=32"`
}

type ApproveCandidatesRequestDto struct {
	Collection string           `json:"collection" binding:"omitempty,max=64"`
	Items      []ApproveItemDto `json:"items" binding:"required,min=1,max=100,dive"`
}

type RejectCandidatesRequestDto struct {
	LexemeIDs []string `json:"lexeme_ids" binding:"required,min=1,max=100,dive,uuid"`
}

type DecisionResponseDto struct {
	Count int `json:"count"`
}
