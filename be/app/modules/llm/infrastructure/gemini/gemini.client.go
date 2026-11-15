// Package gemini calls Google's generateContent REST API with a JSON response schema.
package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	llmin "ielts-practice/app/modules/llm/ports/in"
	"ielts-practice/pkg/httpjson"
)

const api = "https://generativelanguage.googleapis.com/v1beta/models/"

// Client is one Gemini model.
type Client struct {
	key    string
	model  string
	effort string
	client *http.Client
}

var _ llmin.IModel = (*Client)(nil)

func NewClient(apiKey, model, effort string) *Client {
	return &Client{key: apiKey, model: model, effort: thinkingLevel(effort), client: &http.Client{Timeout: 3 * time.Minute}}
}

func (c *Client) Name() string     { return c.model }
func (c *Client) Configured() bool { return c.key != "" && c.model != "" }

// thinkingLevel maps VOCAB_LLM_EFFORT onto Gemini's thinking levels.
func thinkingLevel(effort string) string {
	switch effort {
	case "minimal", "low", "medium", "high":
		return effort
	case "xhigh", "max":
		return "high"
	default:
		return "low"
	}
}

type part struct {
	Text    string `json:"text"`
	Thought bool   `json:"thought,omitempty"`
}

type response struct {
	Candidates []struct {
		Content struct {
			Parts []part `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback *struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	UsageMetadata struct {
		PromptTokenCount        int `json:"promptTokenCount"`
		CandidatesTokenCount    int `json:"candidatesTokenCount"`
		ThoughtsTokenCount      int `json:"thoughtsTokenCount"`
		CachedContentTokenCount int `json:"cachedContentTokenCount"`
	} `json:"usageMetadata"`
}

func (c *Client) CompleteJSON(ctx context.Context, req llmin.Request) (string, llmin.Usage, error) {
	if !c.Configured() {
		return "", llmin.Usage{}, llmin.ErrNotConfigured
	}
	body, _ := json.Marshal(map[string]any{
		"systemInstruction": map[string]any{"parts": []map[string]string{{"text": req.System}}},
		"contents":          []map[string]any{{"role": "user", "parts": []map[string]string{{"text": req.User}}}},
		"generationConfig": map[string]any{
			"responseMimeType":   "application/json",
			"responseJsonSchema": req.Schema,
			"thinkingConfig":     map[string]string{"thinkingLevel": c.effort},
			// Thinking tokens count against the limit, so Gemini gets twice the requested budget.
			"maxOutputTokens": req.MaxTokens * 2,
		},
	})
	var res response
	endpoint := api + url.PathEscape(c.model) + ":generateContent"
	if err := httpjson.Send(ctx, http.MethodPost, endpoint, map[string]string{"x-goog-api-key": c.key}, body, &res, c.client); err != nil {
		return "", llmin.Usage{}, fmt.Errorf("gemini: %w", err)
	}
	// Thinking tokens are billed as output.
	usage := llmin.Usage{
		InputTokens:     res.UsageMetadata.PromptTokenCount,
		OutputTokens:    res.UsageMetadata.CandidatesTokenCount + res.UsageMetadata.ThoughtsTokenCount,
		CacheReadTokens: res.UsageMetadata.CachedContentTokenCount,
	}
	if res.PromptFeedback != nil && res.PromptFeedback.BlockReason != "" {
		return "", usage, llmin.ErrRefused
	}
	if len(res.Candidates) == 0 {
		return "", usage, fmt.Errorf("gemini: no candidates")
	}
	cand := res.Candidates[0]
	switch cand.FinishReason {
	case "MAX_TOKENS":
		return "", usage, llmin.ErrMaxTokens
	case "SAFETY", "PROHIBITED_CONTENT", "BLOCKLIST", "SPII", "RECITATION":
		return "", usage, llmin.ErrRefused
	}
	var text strings.Builder
	for _, p := range cand.Content.Parts {
		if !p.Thought {
			text.WriteString(p.Text)
		}
	}
	return text.String(), usage, nil
}
