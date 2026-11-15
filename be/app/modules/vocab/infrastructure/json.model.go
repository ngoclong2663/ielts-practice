package infrastructure

import (
	"context"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/vocab/ports"
	"ielts-practice/configs"
)

// JSONModel exposes the primary/fallback model pair through ports.IJSONModel.
type JSONModel struct {
	primary, fallback         jsonCompleter
	primaryName, fallbackName string
	logger                    zerolog.Logger
}

var _ ports.IJSONModel = (*JSONModel)(nil)

func NewJSONModel(cfg *configs.Config, logger zerolog.Logger) ports.IJSONModel {
	m := &JSONModel{
		primary:     NewClaudeGenerator(cfg.Vocab.AnthropicAPIKey, cfg.Vocab.Model, cfg.Vocab.Effort),
		primaryName: cfg.Vocab.Model,
		logger:      logger.With().Str("module", "llm").Logger(),
	}
	if cfg.Vocab.FallbackModel != "" {
		m.fallback = NewGeminiGenerator(cfg.Vocab.GeminiAPIKey, cfg.Vocab.FallbackModel, cfg.Vocab.Effort)
		m.fallbackName = cfg.Vocab.FallbackModel
	}
	return m
}

func (m *JSONModel) Configured() bool {
	return (m.primary != nil && m.primary.Configured()) || (m.fallback != nil && m.fallback.Configured())
}

func (m *JSONModel) Complete(ctx context.Context, req ports.JSONRequest, accept func(string) error) (ports.JSONResult, error) {
	var res ports.JSONResult
	var lastErr error = ports.ErrModelNotConfigured
	names := []string{m.primaryName, m.fallbackName}
	for i, c := range []jsonCompleter{m.primary, m.fallback} {
		if c == nil || !c.Configured() {
			continue
		}
		if i == 1 && lastErr != ports.ErrModelNotConfigured {
			m.logger.Warn().Err(lastErr).Str("primary", m.primaryName).Str("fallback", m.fallbackName).
				Msg("Primary model failed; using the fallback model")
		}
		text, u, err := c.completeJSON(ctx, req.System, req.User, req.Schema, req.MaxTokens)
		res.InputTokens += u.InputTokens
		res.OutputTokens += u.OutputTokens
		if err == nil {
			if err = accept(text); err == nil {
				res.Model = names[i]
				return res, nil
			}
		}
		lastErr = err
	}
	return res, lastErr
}
