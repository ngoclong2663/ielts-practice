package generator

import (
	"github.com/rs/zerolog"

	llmin "ielts-practice/app/modules/llm/ports/in"
	vocabout "ielts-practice/app/modules/vocab/ports/out"
	"ielts-practice/configs"
)

// NewContentGenerator: the first model of the chain (Claude, VOCAB_LLM_MODEL), then the second (Gemini,
// VOCAB_FALLBACK_MODEL) for only the words the first could not answer. With one model, it is used alone.
func NewContentGenerator(cfg *configs.Config, models llmin.IJSONModel, logger zerolog.Logger) vocabout.IContentGenerator {
	if cfg.FakeProviders {
		return NewFakeGenerator()
	}
	chain := models.Models()
	primary := NewModelGenerator(chain[0])
	if len(chain) < 2 {
		return primary
	}
	return NewFallbackGenerator(primary, chain[0].Name(), NewModelGenerator(chain[1]), chain[1].Name(),
		logger.With().Str("module", "vocab").Logger())
}

// NewWordSuggester: topic word lists (Epic 4) with the same model chain as card generation.
func NewWordSuggester(cfg *configs.Config, models llmin.IJSONModel) vocabout.IWordSuggester {
	if cfg.FakeProviders {
		return NewFakeSuggester()
	}
	return NewLLMSuggester(models)
}
