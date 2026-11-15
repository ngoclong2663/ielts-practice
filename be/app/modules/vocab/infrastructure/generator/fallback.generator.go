package generator

import (
	"context"
	"strings"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/vocab/domain/entity"
	vocabout "ielts-practice/app/modules/vocab/ports/out"
)

// FallbackGenerator tries the primary model first; words it could not answer (API errors,
// overload, refusals, broken output) are sent to the fallback model in a second call.
type FallbackGenerator struct {
	primary      vocabout.IContentGenerator
	fallback     vocabout.IContentGenerator
	primaryName  string
	fallbackName string
	logger       zerolog.Logger
}

var _ vocabout.IContentGenerator = (*FallbackGenerator)(nil)

func NewFallbackGenerator(primary vocabout.IContentGenerator, primaryName string, fallback vocabout.IContentGenerator, fallbackName string, logger zerolog.Logger) *FallbackGenerator {
	return &FallbackGenerator{primary: primary, fallback: fallback, primaryName: primaryName, fallbackName: fallbackName, logger: logger}
}

func (f *FallbackGenerator) Configured() bool {
	return f.primary.Configured() || f.fallback.Configured()
}

func (f *FallbackGenerator) Generate(ctx context.Context, words []entity.WordRequest) ([]entity.WordResult, entity.Usage, error) {
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
	var missing []entity.WordRequest
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
