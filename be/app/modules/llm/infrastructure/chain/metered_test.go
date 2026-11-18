package chain

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"

	llmin "ielts-practice/app/modules/llm/ports/in"
	llmout "ielts-practice/app/modules/llm/ports/out"
	"ielts-practice/configs"
)

type stubModel struct {
	name  string
	usage llmin.Usage
	err   error
}

func (s stubModel) Name() string     { return s.name }
func (s stubModel) Configured() bool { return true }
func (s stubModel) CompleteJSON(context.Context, llmin.Request) (string, llmin.Usage, error) {
	return `{}`, s.usage, s.err
}

type memLedger struct{ records []llmout.UsageRecord }

func (m *memLedger) Record(_ context.Context, r llmout.UsageRecord) error {
	m.records = append(m.records, r)
	return nil
}

func TestMeteredChainRecordsEveryCall(t *testing.T) {
	ledger := &memLedger{}
	primary := meter(stubModel{name: "claude-haiku-5-5", usage: llmin.Usage{InputTokens: 100, OutputTokens: 40},
		err: llmin.ErrMaxTokens}, llmout.ProviderAnthropic, ledger, zerolog.Nop())
	fallback := meter(stubModel{name: "gemini-3.1-flash-lite", usage: llmin.Usage{InputTokens: 120, OutputTokens: 30, CacheReadTokens: 20}},
		llmout.ProviderGoogle, ledger, zerolog.Nop())
	rateLimited := meter(stubModel{name: "claude-haiku-5-5", err: errors.New("429 rate_limit_error")}, llmout.ProviderAnthropic, ledger, zerolog.Nop())
	noKey := meter(stubModel{name: "claude-haiku-5-5", err: llmin.ErrNotConfigured}, llmout.ProviderAnthropic, ledger, zerolog.Nop())

	res, err := NewChain([]llmin.IModel{primary, fallback}, zerolog.Nop()).
		Complete(context.Background(), llmin.Request{Feature: llmin.FeatureWritingLesson}, func(string) error { return nil })
	if err != nil || res.Model != "gemini-3.1-flash-lite" {
		t.Fatalf("chain: %v %v", res, err)
	}
	if _, _, err := rateLimited.CompleteJSON(context.Background(), llmin.Request{Feature: llmin.FeatureVocabCards}); err == nil {
		t.Fatal("expected the stub error")
	}
	if _, _, err := noKey.CompleteJSON(context.Background(), llmin.Request{}); err == nil {
		t.Fatal("expected the stub error")
	}

	if len(ledger.records) != 3 { // not the call without a key
		t.Fatalf("records: %+v", ledger.records)
	}
	failed, answered := ledger.records[0], ledger.records[1]
	if failed.OK || failed.Provider != llmout.ProviderAnthropic || failed.Feature != llmin.FeatureWritingLesson || failed.OutputTokens != 40 {
		t.Fatalf("failed call: %+v", failed)
	}
	// Gemini counts cached tokens inside the prompt; the ledger keeps them apart.
	if !answered.OK || answered.Model != "gemini-3.1-flash-lite" || answered.InputTokens != 100 || answered.CacheReadTokens != 20 {
		t.Fatalf("answered call: %+v", answered)
	}
	if r := ledger.records[2]; r.OK || r.Feature != llmin.FeatureVocabCards || r.Usage != (llmin.Usage{}) {
		t.Fatalf("rate-limited call: %+v", r)
	}
}

func TestSubscriptionTokenTurnsClaudeOff(t *testing.T) {
	cfg := &configs.Config{}
	cfg.LLM.AnthropicAPIKey, cfg.LLM.Model = "sk-ant-oat01-example", "claude-haiku-5-5"
	if NewMeteredJSONModel(cfg, &memLedger{}, zerolog.Nop()).Configured() {
		t.Fatal("a subscription token must not configure Claude")
	}
	cfg.LLM.AnthropicAPIKey = "sk-ant-api03-example"
	if !NewMeteredJSONModel(cfg, &memLedger{}, zerolog.Nop()).Configured() {
		t.Fatal("an API key configures Claude")
	}
}
