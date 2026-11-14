package infrastructure

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync/atomic"

	"ielts-practice/app/modules/vocab/domain"
	"ielts-practice/app/modules/vocab/domain/entity"
)

// FakeGenerator returns deterministic content without network calls (tests, e2e,
// VOCAB_FAKE_PROVIDERS=true). Entries starting with "zz" are treated as typos.
type FakeGenerator struct {
	Calls atomic.Int64
	Words atomic.Int64
}

var _ domain.IContentGenerator = (*FakeGenerator)(nil)

func NewFakeGenerator() *FakeGenerator { return &FakeGenerator{} }

func (f *FakeGenerator) Configured() bool { return true }

func (f *FakeGenerator) Generate(_ context.Context, words []domain.WordRequest) ([]domain.WordResult, entity.Usage, error) {
	f.Calls.Add(1)
	f.Words.Add(int64(len(words)))
	out := make([]domain.WordResult, 0, len(words))
	for _, w := range words {
		if strings.HasPrefix(w.Input, "zz") {
			out = append(out, domain.WordResult{Input: w.Input, IsValid: false, Suggestion: strings.TrimPrefix(w.Input, "zz")})
			continue
		}
		hw := w.Input
		out = append(out, domain.WordResult{
			Input:        w.Input,
			IsValid:      true,
			ImageQuery:   hw + " photo",
			Visualizable: true,
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
type FakeImageProvider struct{}

var _ domain.IImageProvider = (*FakeImageProvider)(nil)

func NewFakeImageProvider() *FakeImageProvider { return &FakeImageProvider{} }

func (FakeImageProvider) Name() string     { return "fake" }
func (FakeImageProvider) Configured() bool { return true }
func (FakeImageProvider) Expiring() bool   { return false }

var fakeColors = []string{"#10b981", "#0ea5e9", "#f59e0b", "#8b5cf6", "#ef4444", "#14b8a6"}

func (FakeImageProvider) Search(_ context.Context, query string, limit int) ([]entity.CardImage, error) {
	out := make([]entity.CardImage, 0, limit)
	for i := 0; i < limit && i < len(fakeColors); i++ {
		svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="640" height="400"><rect width="100%%" height="100%%" fill="%s"/><text x="50%%" y="50%%" font-size="40" text-anchor="middle" fill="white" font-family="sans-serif">%s</text></svg>`,
			fakeColors[i], escapeXML(query))
		uri := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
		out = append(out, entity.CardImage{
			Provider: "fake", ProviderID: fmt.Sprintf("%s-%d", strings.ReplaceAll(query, " ", "-"), i),
			URL: uri, ThumbURL: uri, Width: 640, Height: 400, Alt: query, AuthorName: "Test images",
		})
	}
	return out, nil
}

func (FakeImageProvider) Lookup(context.Context, string) (*entity.CardImage, error) { return nil, nil }
func (FakeImageProvider) TrackDownload(context.Context, entity.CardImage) error     { return nil }

func escapeXML(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
