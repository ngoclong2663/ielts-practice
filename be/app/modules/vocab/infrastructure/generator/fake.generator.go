package generator

import (
	"context"
	"strings"
	"sync/atomic"

	"ielts-practice/app/modules/vocab/domain/entity"
	vocabout "ielts-practice/app/modules/vocab/ports/out"
)

// FakeGenerator returns deterministic content without network calls (tests, e2e,
// VOCAB_FAKE_PROVIDERS=true). Entries starting with "zz" are treated as typos.
type FakeGenerator struct {
	Calls atomic.Int64
	Words atomic.Int64
}

var _ vocabout.IContentGenerator = (*FakeGenerator)(nil)

func NewFakeGenerator() *FakeGenerator { return &FakeGenerator{} }

func (f *FakeGenerator) Configured() bool { return true }

func (f *FakeGenerator) Generate(_ context.Context, words []entity.WordRequest) ([]entity.WordResult, entity.Usage, error) {
	f.Calls.Add(1)
	f.Words.Add(int64(len(words)))
	out := make([]entity.WordResult, 0, len(words))
	for _, w := range words {
		if strings.HasPrefix(w.Input, "zz") {
			out = append(out, entity.WordResult{Input: w.Input, IsValid: false, Suggestion: strings.TrimPrefix(w.Input, "zz")})
			continue
		}
		hw := w.Input
		out = append(out, entity.WordResult{
			Input:        w.Input,
			IsValid:      true,
			ImageQuery:   hw + " photo",
			Visualizable: true,
			Topic:        "environment",
			Content: entity.CardContent{
				Headword:     hw,
				Lemma:        hw,
				SenseHint:    w.SenseHint,
				IPA:          &entity.IPA{UK: "/" + hw + "/", US: "/" + hw + "/"},
				PartOfSpeech: "noun",
				CEFR:         "B2",
				MeaningVI:    "nghĩa của " + hw,
				DefinitionEN: "the meaning of " + hw,
				WordForms: []entity.WordForm{{
					Form: hw, POS: "noun", IPA: "/" + hw + "/", MeaningVI: "nghĩa của " + hw,
					Example: &entity.Example{EN: "Students often discuss " + hw + " in class.", VI: "Học sinh thường thảo luận về " + hw + " trên lớp."},
				}},
				Collocations: []entity.Collocation{
					{Phrase: "important " + hw, MeaningVI: hw + " quan trọng"},
					{Phrase: "discuss " + hw, MeaningVI: "thảo luận " + hw},
					{Phrase: hw + " policy", MeaningVI: "chính sách " + hw},
				},
				ExtraExamples: []entity.Example{{EN: "Many experts believe " + hw + " matters.", VI: "Nhiều chuyên gia tin rằng " + hw + " quan trọng."}},
				Synonyms:      []string{hw + "-like"},
			},
		})
	}
	return out, entity.Usage{InputTokens: 100 * len(words), OutputTokens: 400 * len(words)}, nil
}

// FakeImageProvider returns inline SVG images (no network). It never expires.
