package usecases

import (
	"math"
	"testing"

	"ielts-practice/app/modules/admin/domain/entity"
)

func TestSummarizeAICost(t *testing.T) {
	rows := []entity.AIUsageRow{
		// 30 days: 1M in + 1M out on Haiku = $0.60; all time adds 1M more output = $1.10
		{Provider: "anthropic", Model: "claude-haiku-5-5", Feature: "writing.lesson", Calls: 3, Calls30d: 2,
			InputTokens: 1_000_000, OutputTokens: 2_000_000, InputTokens30d: 1_000_000, OutputTokens30d: 1_000_000},
		// a failed fallback-model call: 1M output on Flash-Lite = $1.50
		{Provider: "google", Model: "gemini-3.1-flash-lite", Feature: "writing.lesson", Calls: 1, Failed: 1, Calls30d: 1, Failed30d: 1,
			OutputTokens: 1_000_000, OutputTokens30d: 1_000_000},
		{Provider: "anthropic", Model: "claude-haiku-5-5", Feature: "vocab.cards", Calls: 1, Calls30d: 1,
			InputTokens: 1_000_000, InputTokens30d: 1_000_000},
		{Provider: "anthropic", Model: "claude-mystery", Feature: "vocab.cards", Calls: 1, Calls30d: 1, InputTokens: 5, InputTokens30d: 5},
	}
	got := summarizeAICost(rows, nil)
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
	if !near(got.Cost30dUSD, 0.60+1.50+0.10) || !near(got.CostTotalUSD, 1.10+1.50+0.10) {
		t.Fatalf("totals: %v / %v", got.Cost30dUSD, got.CostTotalUSD)
	}
	if len(got.Providers) != 2 || got.Providers[0].Provider != "google" || !near(got.Providers[1].CostTotalUSD, 1.20) {
		t.Fatalf("providers: %+v", got.Providers)
	}
	if len(got.Models) != 3 || got.Models[1].Model != "claude-haiku-5-5" || got.Models[1].CallsTotal != 4 || got.Models[1].Price == nil {
		t.Fatalf("models: %+v", got.Models)
	}
	if last := got.Models[2]; last.Model != "claude-mystery" || last.Price != nil || last.CostTotalUSD != 0 {
		t.Fatalf("unpriced model: %+v", last)
	}
	if got.Models[0].Failed30d != 1 {
		t.Fatalf("failed calls: %+v", got.Models[0])
	}
	if len(got.Features) != 2 || got.Features[0].Feature != "writing.lesson" || got.Features[0].Calls30d != 3 {
		t.Fatalf("features: %+v", got.Features)
	}
}
