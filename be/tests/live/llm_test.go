package live

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/app/modules/vocab/domain/rules"

	"ielts-practice/app/modules/llm/infrastructure/chain"
	"ielts-practice/app/modules/llm/infrastructure/claude"
	"ielts-practice/app/modules/llm/infrastructure/gemini"
	llmin "ielts-practice/app/modules/llm/ports/in"
	"ielts-practice/app/modules/vocab/infrastructure/generator"
	vocabout "ielts-practice/app/modules/vocab/ports/out"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
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
	generators := map[string]vocabout.IContentGenerator{
		model:                   generator.NewModelGenerator(claude.NewClient(os.Getenv("ANTHROPIC_API_KEY"), model, effort)),
		"gemini-3.1-flash-lite": generator.NewModelGenerator(gemini.NewClient(os.Getenv("GEMINI_API_KEY"), "gemini-3.1-flash-lite", effort)),
	}
	words := []entity.WordRequest{{Input: "decide"}, {Input: "take into account"}, {Input: "recieve"}, {Input: "bank", SenseHint: "river"}, {Input: "sightseeing"}}

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
			byInput := map[string]entity.WordResult{}
			for i := range results {
				rules.SanitizeResult(&results[i])
				byInput[results[i].Input] = results[i]
			}
			if r := byInput["recieve"]; r.IsValid || r.Suggestion != "receive" {
				t.Errorf("typo not detected: valid=%v suggestion=%q", r.IsValid, r.Suggestion)
			}
			// The topic tag (Epic 9): a travel word goes to travel, general language to "other".
			if got := byInput["sightseeing"].Topic; got != "travel" {
				t.Errorf("sightseeing topic %q, want travel", got)
			}
			t.Logf("topics: decide=%s, take into account=%s, bank=%s, sightseeing=%s", byInput["decide"].Topic,
				byInput["take into account"].Topic, byInput["bank"].Topic, byInput["sightseeing"].Topic)
			for _, w := range []string{"decide", "take into account", "bank", "sightseeing"} {
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
	single := func(m llmin.IModel) *generator.LLMSuggester {
		return generator.NewLLMSuggester(chain.NewChain([]llmin.IModel{m}, zerolog.Nop()))
	}
	suggesters := map[string]*generator.LLMSuggester{
		model:                   single(claude.NewClient(os.Getenv("ANTHROPIC_API_KEY"), model, "low")),
		"gemini-3.1-flash-lite": single(gemini.NewClient(os.Getenv("GEMINI_API_KEY"), "gemini-3.1-flash-lite", "low")),
	}
	for name, s := range suggesters {
		t.Run(name, func(t *testing.T) {
			if !s.Configured() {
				t.Skip("no API key")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			words, usage, err := s.Suggest(ctx, entity.SuggestRequest{
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
