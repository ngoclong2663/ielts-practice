package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ielts-practice/app/modules/vocab/domain"
	"ielts-practice/app/modules/vocab/domain/entity"
)

const geminiAPI = "https://generativelanguage.googleapis.com/v1beta/models/"

// GeminiGenerator writes card content with Google Gemini (generateContent REST API), using the same
// prompt and JSON schema as the Claude generator so results are interchangeable.
type GeminiGenerator struct {
	key    string
	model  string
	effort string
	client *http.Client
}

var _ domain.IContentGenerator = (*GeminiGenerator)(nil)

func NewGeminiGenerator(apiKey, model, effort string) *GeminiGenerator {
	return &GeminiGenerator{key: apiKey, model: model, effort: geminiThinkingLevel(effort), client: &http.Client{Timeout: 3 * time.Minute}}
}

func (g *GeminiGenerator) Configured() bool { return g.key != "" && g.model != "" }

// Name identifies the model in logs.
func (g *GeminiGenerator) Name() string { return g.model }

// geminiThinkingLevel maps VOCAB_LLM_EFFORT onto Gemini's thinking levels.
func geminiThinkingLevel(effort string) string {
	switch effort {
	case "minimal", "low", "medium", "high":
		return effort
	case "xhigh", "max":
		return "high"
	default:
		return "low"
	}
}

func (g *GeminiGenerator) Generate(ctx context.Context, words []domain.WordRequest) ([]domain.WordResult, entity.Usage, error) {
	if !g.Configured() {
		return nil, entity.Usage{}, domain.ErrGeneratorNotConfigured
	}
	if len(words) == 0 {
		return nil, entity.Usage{}, nil
	}
	return splitOnMaxTokens(words, func(batch []domain.WordRequest) ([]domain.WordResult, entity.Usage, error) {
		return g.call(ctx, batch)
	})
}

type geminiPart struct {
	Text    string `json:"text"`
	Thought bool   `json:"thought,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
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

func (g *GeminiGenerator) call(ctx context.Context, words []domain.WordRequest) ([]domain.WordResult, entity.Usage, error) {
	text, usage, err := g.completeJSON(ctx, systemPrompt, buildUserText(words), outputSchema, 32000)
	if err != nil {
		return nil, usage, err
	}
	results, err := parseResults(text)
	if err != nil {
		return nil, usage, fmt.Errorf("gemini: %w", err)
	}
	return results, usage, nil
}

// completeJSON sends one generateContent request whose answer must match `schema`.
func (g *GeminiGenerator) completeJSON(ctx context.Context, system, user string, schema map[string]any, maxTokens int64) (string, entity.Usage, error) {
	body, _ := json.Marshal(map[string]any{
		"systemInstruction": map[string]any{"parts": []map[string]string{{"text": system}}},
		"contents":          []map[string]any{{"role": "user", "parts": []map[string]string{{"text": user}}}},
		"generationConfig": map[string]any{
			"responseMimeType":   "application/json",
			"responseJsonSchema": schema,
			"thinkingConfig":     map[string]string{"thinkingLevel": g.effort},
			"maxOutputTokens":    maxTokens,
		},
	})
	var res geminiResponse
	endpoint := geminiAPI + url.PathEscape(g.model) + ":generateContent"
	if err := sendJSON(ctx, http.MethodPost, endpoint, map[string]string{"x-goog-api-key": g.key}, body, &res, g.client); err != nil {
		return "", entity.Usage{}, fmt.Errorf("gemini: %w", err)
	}
	// Thinking tokens are billed as output.
	usage := entity.Usage{
		InputTokens:     res.UsageMetadata.PromptTokenCount,
		OutputTokens:    res.UsageMetadata.CandidatesTokenCount + res.UsageMetadata.ThoughtsTokenCount,
		CacheReadTokens: res.UsageMetadata.CachedContentTokenCount,
	}
	if res.PromptFeedback != nil && res.PromptFeedback.BlockReason != "" {
		return "", usage, domain.ErrGeneratorRefused
	}
	if len(res.Candidates) == 0 {
		return "", usage, fmt.Errorf("gemini: no candidates")
	}
	c := res.Candidates[0]
	switch c.FinishReason {
	case "MAX_TOKENS":
		return "", usage, errMaxTokens
	case "SAFETY", "PROHIBITED_CONTENT", "BLOCKLIST", "SPII", "RECITATION":
		return "", usage, domain.ErrGeneratorRefused
	}
	var text strings.Builder
	for _, p := range c.Content.Parts {
		if !p.Thought {
			text.WriteString(p.Text)
		}
	}
	return text.String(), usage, nil
}
