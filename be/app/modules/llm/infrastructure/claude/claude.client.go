// Package claude calls Anthropic's Messages API with structured outputs.
package claude

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"

	llmin "ielts-practice/app/modules/llm/ports/in"
)

// Client is one Claude model: structured outputs for a guaranteed JSON shape, the stable system prompt
// cached, and server-side refusal fallbacks where the model supports them.
type Client struct {
	client anthropic.Client
	model  string
	effort string
	ready  bool
}

var _ llmin.IModel = (*Client)(nil)

func NewClient(apiKey, model, effort string) *Client {
	c := &Client{model: model, effort: effort, ready: apiKey != "" && model != ""}
	if c.ready {
		c.client = anthropic.NewClient(option.WithAPIKey(apiKey), option.WithRequestTimeout(3*time.Minute))
	}
	return c
}

// IsSubscriptionToken reports a Claude subscription OAuth token (sk-ant-oat…, from `claude setup-token`).
// It is not an API key: the Messages API rejects it, so the app needs a key from console.anthropic.com.
func IsSubscriptionToken(key string) bool {
	return strings.HasPrefix(strings.TrimSpace(key), "sk-ant-oat")
}

func (c *Client) Name() string     { return c.model }
func (c *Client) Configured() bool { return c.ready }

func (c *Client) CompleteJSON(ctx context.Context, req llmin.Request) (string, llmin.Usage, error) {
	if !c.ready {
		return "", llmin.Usage{}, llmin.ErrNotConfigured
	}
	params := anthropic.BetaMessageNewParams{
		Model:     c.model,
		MaxTokens: req.MaxTokens,
		System: []anthropic.BetaTextBlockParam{{
			Text:         req.System,
			CacheControl: anthropic.NewBetaCacheControlEphemeralParam(),
		}},
		Messages: []anthropic.BetaMessageParam{
			anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(req.User)),
		},
		OutputConfig: anthropic.BetaOutputConfigParam{
			Effort: anthropic.BetaOutputConfigEffort(c.effort),
			Format: anthropic.BetaJSONOutputFormatParam{Schema: req.Schema},
		},
	}
	if supportsServerFallback(c.model) {
		// Policy declines are re-served by the server's default fallback model inside the same call.
		params.Betas = []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01}
		params.Fallbacks = anthropic.BetaFallbacksParamOfDefault()
	}
	resp, err := c.client.Beta.Messages.New(ctx, params)
	if err != nil {
		return "", llmin.Usage{}, fmt.Errorf("claude: %w", err)
	}
	usage := llmin.Usage{
		InputTokens:     int(resp.Usage.InputTokens),
		OutputTokens:    int(resp.Usage.OutputTokens),
		CacheReadTokens: int(resp.Usage.CacheReadInputTokens),
	}
	switch resp.StopReason {
	case anthropic.BetaStopReasonRefusal:
		return "", usage, llmin.ErrRefused
	case anthropic.BetaStopReasonMaxTokens:
		return "", usage, llmin.ErrMaxTokens
	}
	var text strings.Builder
	for _, block := range resp.Content {
		if b, ok := block.AsAny().(anthropic.BetaTextBlock); ok {
			text.WriteString(b.Text)
		}
	}
	return text.String(), usage, nil
}

// supportsServerFallback: the `fallbacks: "default"` parameter exists only for these models
// (Haiku has no server-side refusal fallback; our Gemini fallback covers it instead).
func supportsServerFallback(model string) bool {
	switch model {
	case "claude-fable-5-1", "claude-opus-5-5", "claude-opus-5", "claude-sonnet-5-5":
		return true
	}
	return false
}
