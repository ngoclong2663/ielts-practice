// Package in is what the llm module offers to other modules: JSON answers from the configured
// language models (VOCAB_LLM_MODEL first, VOCAB_FALLBACK_MODEL as fallback).
package in

import (
	"context"
	"errors"
)

var (
	// ErrNotConfigured: the model has no API key (or no model is configured at all).
	ErrNotConfigured = errors.New("llm: model is not configured")
	// ErrMaxTokens: the answer was cut off at the token limit; callers may split the work and retry.
	ErrMaxTokens = errors.New("llm: output hit the token limit")
	// ErrRefused: the model declined the request; retrying the same model will not help.
	ErrRefused = errors.New("llm: the model refused the request")
)

// Request is one structured-output call. Schema is a JSON Schema the answer must match.
type Request struct {
	System    string
	User      string
	Schema    map[string]any
	MaxTokens int64
	// Feature names what the call is for in the usage ledger (e.g. "writing.lesson"); see Feature* below.
	Feature string
}

// Features: the labels of the usage ledger (admin board, Epic 7).
const (
	FeatureVocabCards    = "vocab.cards"
	FeatureVocabTopic    = "vocab.topic_words"
	FeatureWritingLesson = "writing.lesson"
	FeatureWritingVocab  = "writing.vocab"
	FeatureWritingCheck  = "writing.check"
	FeatureWritingReview = "writing.submit"
)

type Usage struct {
	InputTokens     int
	OutputTokens    int
	CacheReadTokens int
}

func (u *Usage) Add(o Usage) {
	u.InputTokens += o.InputTokens
	u.OutputTokens += o.OutputTokens
	u.CacheReadTokens += o.CacheReadTokens
}

// IModel is one model (Claude or Gemini).
type IModel interface {
	Name() string
	Configured() bool
	// CompleteJSON returns the raw JSON text of the answer.
	CompleteJSON(ctx context.Context, req Request) (string, Usage, error)
}

// Result reports which model answered and the tokens used by every attempt.
type Result struct {
	Model string
	Usage
}

// IJSONModel is the configured chain: the primary model, then the fallback when the primary fails or
// its answer is rejected by `accept`.
type IJSONModel interface {
	Configured() bool
	// Models returns the chain in order, for callers that split work between models themselves.
	Models() []IModel
	Complete(ctx context.Context, req Request, accept func(text string) error) (Result, error)
}
