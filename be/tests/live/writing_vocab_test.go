package live

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"ielts-practice/app/modules/llm/infrastructure/chain"
	"ielts-practice/app/modules/writing/domain/entity"
	"ielts-practice/app/modules/writing/infrastructure/coach"
	"ielts-practice/configs"
)

// Calls the real models for one lesson's useful vocabulary (Epic 7); costs well under a cent.
// Run with LLM_INTEGRATION=1. WRITING_LIVE_TASK picks the task (default cam21-test1-task1),
// WRITING_LIVE_MODEL_ANSWER a file holding the lesson's model answer, WRITING_LIVE_OUT a folder for the JSON.
func TestVocabLive(t *testing.T) {
	if os.Getenv("LLM_INTEGRATION") != "1" {
		t.Skip("set LLM_INTEGRATION=1 to call the real LLM APIs")
	}
	_ = godotenv.Load("../../.env")
	cfg, err := configs.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.FakeProviders = false
	c := coach.NewCoach(cfg, chain.NewJSONModel(cfg, zerolog.Nop()), zerolog.Nop())
	if !c.Configured() {
		t.Skip("no API key")
	}
	slug := os.Getenv("WRITING_LIVE_TASK")
	if slug == "" {
		slug = "cam21-test1-task1"
	}
	req := entity.VocabRequest{Task: cam21Task(t, slug), Position: os.Getenv("WRITING_LIVE_POSITION")}
	if f := os.Getenv("WRITING_LIVE_MODEL_ANSWER"); f != "" {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		req.ModelAnswer = string(b)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	start := time.Now()
	vocab, usage, err := c.WriteVocab(ctx, req)
	if err != nil {
		t.Fatalf("vocab: %v", err)
	}
	t.Logf("vocab by %s in %s, %d in / %d out tokens", usage.Model, time.Since(start).Round(time.Millisecond), usage.InputTokens, usage.OutputTokens)
	if len(vocab.Groups) != 3 {
		t.Errorf("want 3 groups, got %d", len(vocab.Groups))
	}
	for _, g := range vocab.Groups {
		t.Logf("%s: %d items", g.Key, len(g.Items))
		for _, it := range g.Items {
			if len(it.Synonyms) == 0 {
				t.Errorf("%s: no synonyms", it.Term)
			}
		}
	}
	if dir := os.Getenv("WRITING_LIVE_OUT"); dir != "" {
		b, _ := json.MarshalIndent(vocab, "", "  ")
		_ = os.WriteFile(filepath.Join(dir, "vocab-"+slug+".json"), b, 0o644)
	}
}
