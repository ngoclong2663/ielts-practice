package infrastructure

import (
	"context"
	"strings"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/vocab/domain"
	"ielts-practice/app/modules/vocab/domain/entity"
)

// FallbackGenerator tries the primary model first; words it could not answer (API errors,
// overload, refusals, broken output) are sent to the fallback model in a second call.
type FallbackGenerator struct {
	primary      domain.IContentGenerator
	fallback     domain.IContentGenerator
	primaryName  string
	fallbackName string
	logger       zerolog.Logger
}

var _ domain.IContentGenerator = (*FallbackGenerator)(nil)

func NewFallbackGenerator(primary domain.IContentGenerator, primaryName string, fallback domain.IContentGenerator, fallbackName string, logger zerolog.Logger) *FallbackGenerator {
	return &FallbackGenerator{primary: primary, fallback: fallback, primaryName: primaryName, fallbackName: fallbackName, logger: logger}
}

func (f *FallbackGenerator) Configured() bool {
	return f.primary.Configured() || f.fallback.Configured()
}

func (f *FallbackGenerator) Generate(ctx context.Context, words []domain.WordRequest) ([]domain.WordResult, entity.Usage, error) {
	if !f.primary.Configured() {
		return f.fallback.Generate(ctx, words)
	}
	results, usage, err := f.primary.Generate(ctx, words)
	if !f.fallback.Configured() {
		return results, usage, err
	}
	answered := map[string]bool{}
	for _, r := range results {
		answered[strings.ToLower(strings.TrimSpace(r.Input))] = true
	}
	var missing []domain.WordRequest
	for _, w := range words {
		if !answered[strings.ToLower(w.Input)] {
			missing = append(missing, w)
		}
	}
	if len(missing) == 0 {
		return results, usage, nil
	}
	f.logger.Warn().Err(err).Str("primary", f.primaryName).Str("fallback", f.fallbackName).
		Int("words", len(missing)).Msg("Primary model failed; using the fallback model")
	more, extra, ferr := f.fallback.Generate(ctx, missing)
	usage.Add(extra)
	return append(results, more...), usage, ferr
}
