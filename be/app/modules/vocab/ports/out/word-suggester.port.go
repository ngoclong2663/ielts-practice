package out

import (
	"context"

	"ielts-practice/app/modules/vocab/domain/entity"
)

// IWordSuggester proposes new words for a topic. Implementations: Claude with Gemini fallback, fake.
type IWordSuggester interface {
	Configured() bool
	Suggest(ctx context.Context, req entity.SuggestRequest) ([]entity.SuggestedWord, entity.Usage, error)
}
