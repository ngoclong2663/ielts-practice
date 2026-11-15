package ports

import (
	"context"
	"errors"
)

// ErrModelNotConfigured: neither the primary nor the fallback model has an API key.
var ErrModelNotConfigured = errors.New("llm: no model is configured")

// JSONRequest is one structured-output call. Schema is a JSON Schema the answer must match.
type JSONRequest struct {
	System    string
	User      string
	Schema    map[string]any
	MaxTokens int64
}

// JSONResult reports which model answered and the tokens used by every attempt.
type JSONResult struct {
	Model        string
	InputTokens  int
	OutputTokens int
}

// IJSONModel gives other modules (writing) the card-generation models: the primary model
// (VOCAB_LLM_MODEL), then the fallback (VOCAB_FALLBACK_MODEL) when the primary fails or its
// answer is rejected by `accept`.
type IJSONModel interface {
	Configured() bool
	Complete(ctx context.Context, req JSONRequest, accept func(text string) error) (JSONResult, error)
}
