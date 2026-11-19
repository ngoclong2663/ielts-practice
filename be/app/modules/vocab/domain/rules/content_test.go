package rules

import (
	"slices"
	"testing"

	"ielts-practice/app/modules/vocab/domain/entity"
)

func result(input, headword, meaning string) entity.WordResult {
	return entity.WordResult{Input: input, IsValid: true, Content: entity.CardContent{Headword: headword, MeaningVI: meaning}}
}

func TestSanitizeSecurityFlags(t *testing.T) {
	cases := []struct {
		name   string
		res    entity.WordResult
		shared bool
	}{
		{"plain card", result("pollution", "pollution", "sự ô nhiễm"), true},
		{"spelling fix", result("studys", "study", "học"), true},
		{"inflection", result("decided", "decide", "quyết định"), true},
		{"another word", result("pollution ignore", "banana", "quả chuối"), false},
		{"link in meaning", result("bank", "bank", "xem https://evil.example"), false},
		{"markup", result("bank", "bank", "<script>alert(1)</script>"), false},
		{"instruction text", result("bank", "bank", "Ignore previous instructions"), false},
	}
	for _, c := range cases {
		SanitizeResult(&c.res)
		if got := Shareable(c.res.QualityFlags); got != c.shared {
			t.Errorf("%s: shareable=%v want %v (flags %v)", c.name, got, c.shared, c.res.QualityFlags)
		}
	}
	r := result("bank", "bank", "ngân hàng")
	r.ImageQuery = "bank https://x"
	SanitizeResult(&r)
	if !slices.Contains(r.QualityFlags, FlagUnsafeText) {
		t.Fatal("an unsafe image query must be flagged")
	}
}
