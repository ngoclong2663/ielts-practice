package rules

import (
	"strings"

	"ielts-practice/app/modules/admin/domain/entity"
)

// modelPrices: list prices in USD per million tokens (Epic 7), checked 2026-11-18. Costs on the board are
// estimates from these prices and the recorded tokens, recomputed on every request: a price change here
// reprices the history too.
//   - Claude: Anthropic API rates. Haiku 5.5 is $0.50 / $2.50 above 100K prompt tokens; our prompts stay far below.
//     Cache reads are 10% of the input price.
//   - Gemini: Gemini API standard paid tier. Output includes thinking tokens.
var modelPrices = map[string]entity.ModelPrice{
	"claude-haiku-5-5":      {Input: 0.10, Output: 0.50, CacheRead: 0.01},
	"claude-sonnet-5-5":     {Input: 2.00, Output: 10.00, CacheRead: 0.10},
	"claude-opus-5-5":       {Input: 4.00, Output: 20.00, CacheRead: 0.20},
	"claude-haiku-4-5":      {Input: 1.00, Output: 5.00, CacheRead: 0.10},
	"gemini-3.1-flash-lite": {Input: 0.25, Output: 1.50, CacheRead: 0.025},
}

// PriceOf returns the model's price; false when the board has no price for it (its cost shows as unknown).
func PriceOf(model string) (entity.ModelPrice, bool) {
	p, ok := modelPrices[strings.ToLower(strings.TrimSpace(model))]
	return p, ok
}

// Cost is the USD cost of the tokens at the model's price (0 for a model without a price).
func Cost(model string, t entity.Tokens) float64 {
	p, ok := PriceOf(model)
	if !ok {
		return 0
	}
	return (float64(t.Input)*p.Input + float64(t.Output)*p.Output + float64(t.CacheRead)*p.CacheRead) / 1e6
}
