package rules

import (
	"math"
	"testing"

	"ielts-practice/app/modules/admin/domain/entity"
)

func TestCost(t *testing.T) {
	cases := []struct {
		model string
		t     entity.Tokens
		want  float64
	}{
		{"claude-haiku-5-5", entity.Tokens{Input: 1_000_000, Output: 1_000_000}, 0.60},
		{"claude-haiku-5-5", entity.Tokens{Input: 2000, Output: 500, CacheRead: 10_000}, 0.00055},
		{"gemini-3.1-flash-lite", entity.Tokens{Input: 1_000_000, Output: 2_000_000, CacheRead: 1_000_000}, 3.275},
		{" Claude-Haiku-5-5 ", entity.Tokens{Output: 1_000_000}, 0.50},
		{"some-new-model", entity.Tokens{Input: 1_000_000}, 0},
	}
	for _, c := range cases {
		if got := Cost(c.model, c.t); math.Abs(got-c.want) > 1e-9 {
			t.Errorf("Cost(%q, %+v) = %v, want %v", c.model, c.t, got, c.want)
		}
	}
	if _, ok := PriceOf("some-new-model"); ok {
		t.Error("unknown model should have no price")
	}
}
