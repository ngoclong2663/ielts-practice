package infrastructure

import (
	"github.com/rs/zerolog"

	"ielts-practice/app/modules/vocab/domain"
	"ielts-practice/configs"
)

// NewContentGenerator: Claude (VOCAB_LLM_MODEL) first, Gemini (VOCAB_FALLBACK_MODEL) for the words Claude
// could not answer. With only one key configured, that provider is used alone.
func NewContentGenerator(cfg *configs.Config, logger zerolog.Logger) domain.IContentGenerator {
	if cfg.Vocab.FakeProviders {
		return NewFakeGenerator()
	}
	primary := NewClaudeGenerator(cfg.Vocab.AnthropicAPIKey, cfg.Vocab.Model, cfg.Vocab.Effort)
	if cfg.Vocab.FallbackModel == "" {
		return primary
	}
	fallback := NewGeminiGenerator(cfg.Vocab.GeminiAPIKey, cfg.Vocab.FallbackModel, cfg.Vocab.Effort)
	return NewFallbackGenerator(primary, cfg.Vocab.Model, fallback, cfg.Vocab.FallbackModel, logger.With().Str("module", "vocab").Logger())
}

// NewImageProviders returns the providers in the configured search order (VOCAB_IMAGE_PROVIDERS).
func NewImageProviders(cfg *configs.Config) domain.ImageProviders {
	if cfg.Vocab.FakeProviders {
		return domain.ImageProviders{NewFakeImageProvider()}
	}
	out := domain.ImageProviders{}
	for _, name := range cfg.Vocab.ImageProviders {
		switch name {
		case "pixabay":
			out = append(out, NewPixabayProvider(cfg.Vocab.PixabayAPIKey))
		case "unsplash":
			out = append(out, NewUnsplashProvider(cfg.Vocab.UnsplashKey))
		}
	}
	return out
}
