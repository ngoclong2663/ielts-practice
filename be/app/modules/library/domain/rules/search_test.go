package rules

import (
	"strings"
	"testing"

	"ielts-practice/app/modules/library/domain/entity"
)

func TestFold(t *testing.T) {
	cases := map[string]string{
		"Ô nhiễm":               "o nhiem",
		"sự ô nhiễm môi trường": "su o nhiem moi truong",
		"Đô thị hoá":            "do thi hoa",
		"  Renewable-Energy ":   "renewable energy",
	}
	for in, want := range cases {
		if got := Fold(in); got != want {
			t.Errorf("Fold(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSearchTextMatches(t *testing.T) {
	c := &entity.CardContent{Headword: "pollution", MeaningVI: "sự ô nhiễm", Synonyms: []string{"contamination"}}
	s := SearchText(c)
	for _, q := range []string{"pollu", "o nhiem", "ô nhiễm", "contamin"} {
		if !strings.Contains(s, Fold(q)) {
			t.Errorf("%q not found in %q", q, s)
		}
	}
	if LikePattern("50%_a") != `%50\%\_a%` {
		t.Fatal(LikePattern("50%_a"))
	}
}
