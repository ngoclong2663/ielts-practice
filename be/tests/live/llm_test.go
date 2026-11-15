package live

import (
	"context"
	"encoding/json"
	vocabinfra "ielts-practice/app/modules/vocab/infrastructure"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"ielts-practice/app/modules/vocab/domain"
)

// Calls the real Claude and Gemini APIs with a few words (costs a few cents).
// Run with LLM_INTEGRATION=1; keys come from be/.env.
func TestContentGeneratorsLive(t *testing.T) {
	if os.Getenv("LLM_INTEGRATION") != "1" {
		t.Skip("set LLM_INTEGRATION=1 to call the real LLM APIs")
	}
	_ = godotenv.Load("../../.env") // be/.env
	effort := os.Getenv("VOCAB_LLM_EFFORT")
	if effort == "" {
		effort = "low"
	}
	model := os.Getenv("VOCAB_LLM_MODEL")
	if model == "" {
		model = "claude-haiku-5-5"
	}
	generators := map[string]domain.IContentGenerator{
		model:                   vocabinfra.NewClaudeGenerator(os.Getenv("ANTHROPIC_API_KEY"), model, effort),
		"gemini-3.1-flash-lite": vocabinfra.NewGeminiGenerator(os.Getenv("GEMINI_API_KEY"), "gemini-3.1-flash-lite", effort),
	}
	words := []domain.WordRequest{{Input: "decide"}, {Input: "take into account"}, {Input: "recieve"}, {Input: "bank", SenseHint: "river"}}

	for name, g := range generators {
		t.Run(name, func(t *testing.T) {
			if !g.Configured() {
				t.Skip("no API key")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			start := time.Now()
			results, usage, err := g.Generate(ctx, words)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			if len(results) != len(words) {
				t.Fatalf("got %d results for %d words", len(results), len(words))
			}
			byInput := map[string]domain.WordResult{}
			for i := range results {
				domain.SanitizeResult(&results[i])
				byInput[results[i].Input] = results[i]
			}
			if r := byInput["recieve"]; r.IsValid || r.Suggestion != "receive" {
				t.Errorf("typo not detected: valid=%v suggestion=%q", r.IsValid, r.Suggestion)
			}
			for _, w := range []string{"decide", "take into account", "bank"} {
				r := byInput[w]
				if !r.IsValid || r.Content.MeaningVI == "" || len(r.Content.WordForms) == 0 || len(r.Content.Collocations) < 3 {
					t.Errorf("%s: incomplete card: %+v", w, r)
				}
			}
			sample, _ := json.MarshalIndent(byInput["decide"].Content, "", "  ")
			t.Logf("%s: %v, input %d / output %d / cache-read %d tokens, flags %v\n%s",
				name, time.Since(start).Round(time.Millisecond), usage.InputTokens, usage.OutputTokens, usage.CacheReadTokens,
				byInput["decide"].QualityFlags, sample)
			t.Logf("bank (river): %s · %s", byInput["bank"].Content.MeaningVI, byInput["bank"].Content.DefinitionEN)
		})
	}
}

// Topic word suggestions (Epic 4) with each real model.
func TestWordSuggesterLive(t *testing.T) {
	if os.Getenv("LLM_INTEGRATION") != "1" {
		t.Skip("set LLM_INTEGRATION=1 to call the real LLM APIs")
	}
	_ = godotenv.Load("../../.env") // be/.env
	model := os.Getenv("VOCAB_LLM_MODEL")
	if model == "" {
		model = "claude-haiku-5-5"
	}
	// One suggester per model, without a fallback, so each model is tested on its own.
	suggesters := map[string]*vocabinfra.LLMSuggester{
		model: vocabinfra.NewLLMSuggester(vocabinfra.NewClaudeGenerator(os.Getenv("ANTHROPIC_API_KEY"), model, "low"),
			model, nil, "", zerolog.Nop()),
		"gemini-3.1-flash-lite": vocabinfra.NewLLMSuggester(vocabinfra.NewGeminiGenerator(os.Getenv("GEMINI_API_KEY"), "gemini-3.1-flash-lite", "low"),
			"gemini-3.1-flash-lite", nil, "", zerolog.Nop()),
	}
	for name, s := range suggesters {
		t.Run(name, func(t *testing.T) {
			if !s.Configured() {
				t.Skip("no API key")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			words, usage, err := s.Suggest(ctx, domain.SuggestRequest{
				Topic: "health and hospitals: illness, treatment, hospitals, healthcare systems, public health",
				Count: 12, Exclude: []string{"diagnosis", "treatment", "patient"},
			})
			if err != nil {
				t.Fatalf("suggest: %v", err)
			}
			var list []string
			for _, w := range words {
				if w.Lemma == "diagnosis" || w.Lemma == "treatment" || w.Lemma == "patient" {
					t.Logf("model ignored the exclude list for %q (the server filters it)", w.Lemma)
				}
				list = append(list, w.Lemma+" ("+w.CEFR+", "+w.MeaningVI+")")
			}
			if len(words) < 10 {
				t.Errorf("only %d words", len(words))
			}
			t.Logf("%s: %d words, input %d / output %d tokens\n%v", name, len(words), usage.InputTokens, usage.OutputTokens, list)
		})
	}
}
