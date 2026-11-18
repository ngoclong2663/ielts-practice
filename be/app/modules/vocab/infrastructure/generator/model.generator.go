package generator

import (
	"context"
	"errors"
	"fmt"

	llmin "ielts-practice/app/modules/llm/ports/in"
	"ielts-practice/app/modules/vocab/domain/entity"
	vocabout "ielts-practice/app/modules/vocab/ports/out"
)

// cardMaxTokens bounds one card batch (the Gemini client doubles it for thinking tokens).
const cardMaxTokens = 16000

// ModelGenerator writes card content with one language model. A batch cut off at the token limit
// is split in half and retried, so every word still gets a full answer.
type ModelGenerator struct {
	model llmin.IModel
}

var _ vocabout.IContentGenerator = (*ModelGenerator)(nil)

func NewModelGenerator(model llmin.IModel) *ModelGenerator { return &ModelGenerator{model: model} }

func (g *ModelGenerator) Configured() bool { return g.model.Configured() }

func (g *ModelGenerator) Generate(ctx context.Context, words []entity.WordRequest) ([]entity.WordResult, entity.Usage, error) {
	if !g.model.Configured() {
		return nil, entity.Usage{}, vocabout.ErrGeneratorNotConfigured
	}
	if len(words) == 0 {
		return nil, entity.Usage{}, nil
	}
	return splitOnMaxTokens(words, func(batch []entity.WordRequest) ([]entity.WordResult, entity.Usage, error) {
		return g.call(ctx, batch)
	})
}

func (g *ModelGenerator) call(ctx context.Context, words []entity.WordRequest) ([]entity.WordResult, entity.Usage, error) {
	text, u, err := g.model.CompleteJSON(ctx, llmin.Request{
		System: systemPrompt, User: buildUserText(words), Schema: outputSchema, MaxTokens: cardMaxTokens,
		Feature: llmin.FeatureVocabCards,
	})
	usage := entity.Usage{InputTokens: u.InputTokens, OutputTokens: u.OutputTokens, CacheReadTokens: u.CacheReadTokens}
	switch {
	case errors.Is(err, llmin.ErrRefused):
		return nil, usage, vocabout.ErrGeneratorRefused
	case errors.Is(err, llmin.ErrNotConfigured):
		return nil, usage, vocabout.ErrGeneratorNotConfigured
	case err != nil:
		return nil, usage, err
	}
	results, err := parseResults(text)
	if err != nil {
		return nil, usage, fmt.Errorf("%s: %w", g.model.Name(), err)
	}
	return results, usage, nil
}
