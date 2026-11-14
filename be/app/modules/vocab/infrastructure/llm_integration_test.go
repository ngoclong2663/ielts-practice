package infrastructure

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	"ielts-practice/app/modules/vocab/domain"
)

// Calls the real Claude and Gemini APIs with a few words (costs a few cents).
// Run with LLM_INTEGRATION=1; keys come from be/.env or the project-root .env.
func TestContentGeneratorsLive(t *testing.T) {
	if os.Getenv("LLM_INTEGRATION") != "1" {
		t.Skip("set LLM_INTEGRATION=1 to call the real LLM APIs")
	}
	_ = godotenv.Load("../../../../.env")
	_ = godotenv.Load("../../../../../.env")
	effort := os.Getenv("VOCAB_LLM_EFFORT")
	if effort == "" {
		effort = "low"
	}
	model := os.Getenv("VOCAB_LLM_MODEL")
	if model == "" {
		model = "claude-haiku-5-5"
	}
	generators := map[string]domain.IContentGenerator{
		model:                   NewClaudeGenerator(os.Getenv("ANTHROPIC_API_KEY"), model, effort),
		"gemini-3.1-flash-lite": NewGeminiGenerator(os.Getenv("GEMINI_API_KEY"), "gemini-3.1-flash-lite", effort),
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
