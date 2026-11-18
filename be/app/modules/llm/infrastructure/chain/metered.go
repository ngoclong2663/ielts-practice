package chain

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"

	llmin "ielts-practice/app/modules/llm/ports/in"
	llmout "ielts-practice/app/modules/llm/ports/out"
)

const recordTimeout = 5 * time.Second

// metered records every call of one model in the usage ledger, answered or not: a failed call can still be
// billed, a fallback answer costs the primary call too, and a model that keeps failing should be visible.
// Recording never fails the call.
type metered struct {
	llmin.IModel
	provider string
	ledger   llmout.IUsageLedger
	logger   zerolog.Logger
}

func meter(m llmin.IModel, provider string, ledger llmout.IUsageLedger, logger zerolog.Logger) llmin.IModel {
	return &metered{IModel: m, provider: provider, ledger: ledger, logger: logger}
}

func (m *metered) CompleteJSON(ctx context.Context, req llmin.Request) (string, llmin.Usage, error) {
	text, usage, err := m.IModel.CompleteJSON(ctx, req)
	if errors.Is(err, llmin.ErrNotConfigured) {
		return text, usage, err // no key: nothing was sent
	}
	// A failed call is recorded even with no tokens (e.g. a 429), so the board shows a model that keeps failing.
	rec := llmout.UsageRecord{Provider: m.provider, Model: m.Name(), Feature: req.Feature, OK: err == nil, Usage: usage}
	if m.provider == llmout.ProviderGoogle {
		// Gemini's prompt count includes the cached tokens; the ledger keeps them apart, as Claude reports them.
		rec.InputTokens = max(0, rec.InputTokens-rec.CacheReadTokens)
	}
	rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), recordTimeout)
	defer cancel()
	if lerr := m.ledger.Record(rctx, rec); lerr != nil {
		m.logger.Error().Err(lerr).Str("model", rec.Model).Msg("Failed to record AI usage")
	}
	return text, usage, err
}
