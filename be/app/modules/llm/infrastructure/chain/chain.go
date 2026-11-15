// Package chain tries the configured models in order (primary, then fallback).
package chain

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/llm/infrastructure/claude"
	"ielts-practice/app/modules/llm/infrastructure/gemini"
	llmin "ielts-practice/app/modules/llm/ports/in"
	"ielts-practice/configs"
)

// Chain tries each configured model in order until one gives an accepted answer.
type Chain struct {
	models []llmin.IModel
	logger zerolog.Logger
}

var _ llmin.IJSONModel = (*Chain)(nil)

// NewJSONModel: Claude (VOCAB_LLM_MODEL) first, then Gemini (VOCAB_FALLBACK_MODEL) unless it is set empty.
func NewJSONModel(cfg *configs.Config, logger zerolog.Logger) llmin.IJSONModel {
	models := []llmin.IModel{claude.NewClient(cfg.LLM.AnthropicAPIKey, cfg.LLM.Model, cfg.LLM.Effort)}
	if cfg.LLM.FallbackModel != "" {
		models = append(models, gemini.NewClient(cfg.LLM.GeminiAPIKey, cfg.LLM.FallbackModel, cfg.LLM.Effort))
	}
	return NewChain(models, logger)
}

func NewChain(models []llmin.IModel, logger zerolog.Logger) *Chain {
	return &Chain{models: models, logger: logger.With().Str("module", "llm").Logger()}
}

func (c *Chain) Models() []llmin.IModel { return c.models }

func (c *Chain) Configured() bool {
	for _, m := range c.models {
		if m.Configured() {
			return true
		}
	}
	return false
}

func (c *Chain) Complete(ctx context.Context, req llmin.Request, accept func(string) error) (llmin.Result, error) {
	var res llmin.Result
	var lastErr error = llmin.ErrNotConfigured
	var previous string
	for _, m := range c.models {
		if !m.Configured() {
			continue
		}
		if !errors.Is(lastErr, llmin.ErrNotConfigured) {
			c.logger.Warn().Err(lastErr).Str("failed", previous).Str("next", m.Name()).Msg("Model failed; trying the next model")
		}
		text, usage, err := m.CompleteJSON(ctx, req)
		res.Usage.Add(usage)
		if err == nil {
			if err = accept(text); err == nil {
				res.Model = m.Name()
				return res, nil
			}
		}
		lastErr, previous = err, m.Name()
	}
	return res, lastErr
}
