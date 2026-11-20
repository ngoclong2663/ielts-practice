package rules

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"ielts-practice/app/modules/exam/domain/entity"
)

func TestExpandVariants(t *testing.T) {
	cases := map[string][]string{
		"10/ten":                 {"10", "ten"},
		"sport(s)":               {"sports", "sport"},
		"(the) harbour":          {"the harbour", "harbour"},
		"(mental) concentration": {"mental concentration", "concentration"},
		"10.25 (am)":             {"10.25 am", "10.25"},
		"intensive/sport(s)":     {"intensive", "sports", "sport"},
		"Café":                   {"café"},
	}
	for in, want := range cases {
		if got := ExpandVariants(in); !slices.Equal(got, want) {
			t.Errorf("ExpandVariants(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestWithinLimit(t *testing.T) {
	cases := []struct {
		answer string
		limit  int
		num    bool
		ok     bool
	}{
		{"sensors", 1, false, true},
		{"motion sensors", 1, false, false},
		{"check-in", 1, false, true},
		{"30", 1, true, true},
		{"30", 1, false, true},
		{"30 kilometres", 1, true, true},
		{"30 long kilometres", 1, true, false},
		{"10 30", 1, true, false},
		{"3,000", 0, true, true},
		{"three thousand", 0, true, false},
		{"in the morning", 2, true, false},
		{"anything at all goes", 0, false, true},
		{"£35", 1, false, true},
	}
	for _, c := range cases {
		if got := WithinLimit(c.answer, c.limit, c.num); got != c.ok {
			t.Errorf("WithinLimit(%q, %d, %v) = %v, want %v", c.answer, c.limit, c.num, got, c.ok)
		}
	}
}

func TestMatchesAny(t *testing.T) {
	if !MatchesAny("  Towel. ", []string{"towel"}) || !MatchesAny("THIRTY", []string{"30/thirty"}) || !MatchesAny("sport", []string{"sport(s)"}) {
		t.Fatal("expected matches")
	}
	if MatchesAny("towels", []string{"towel"}) || MatchesAny("", []string{"towel"}) || MatchesAny("towl", []string{"towel"}) {
		t.Fatal("unexpected match")
	}
	if !MatchesAny("don’t", []string{"don't"}) {
		t.Fatal("curly apostrophe must match a straight one")
	}
}

func TestBand(t *testing.T) {
	cases := []struct {
		skill   string
		correct int
		want    float64
	}{
		{entity.SkillReading, 40, 9}, {entity.SkillReading, 33, 7.5}, {entity.SkillReading, 32, 7}, {entity.SkillReading, 23, 6},
		{entity.SkillReading, 22, 5.5}, {entity.SkillReading, 15, 5}, {entity.SkillReading, 10, 4},
		{entity.SkillListening, 32, 7.5}, {entity.SkillListening, 31, 7}, {entity.SkillListening, 26, 6.5}, {entity.SkillListening, 18, 5.5},
		{entity.SkillListening, 17, 5}, {entity.SkillListening, 13, 4.5},
	}
	for _, c := range cases {
		if got := Band(c.skill, c.correct); got == nil || *got != c.want {
			t.Errorf("Band(%s, %d) = %v, want %v", c.skill, c.correct, got, c.want)
		}
	}
	if Band(entity.SkillReading, 9) != nil {
		t.Error("below 10 has no band")
	}
}

// samplePart is a small valid reading part used by the grading and validation tests.
func samplePart() (*entity.Part, entity.Key) {
	p := &entity.Part{Slug: "s0-t1-r1", Skill: entity.SkillReading, PartNo: 1, Title: "Light after dark", QFrom: 1, QTo: 6,
		Content: entity.Content{
			Paragraphs: []entity.Paragraph{
				{ID: "A", Label: "A", Text: "Councils replaced the lamps mainly because LEDs use half as much electricity."},
				{ID: "B", Label: "B", Text: "Lights are dimmed to thirty per cent and motion sensors raise them again. Stations and hospitals need constant light."},
			},
			Groups: []entity.Group{
				{Type: entity.TypeTFNG, From: 1, To: 2, Instructions: "Do the statements agree?", Items: []entity.Item{{N: 1, Text: "LEDs cut costs."}, {N: 2, Text: "A company paid for the study."}}},
				{Type: entity.TypeNotes, From: 3, To: 4, Instructions: "ONE WORD AND/OR A NUMBER", WordLimit: 1, AllowNumber: true,
					Lines: []string{"Dimmed to {3} per cent", "{4} raise the lights"}},
				{Type: entity.TypeChooseN, From: 5, To: 6, Pick: 2, Instructions: "Choose TWO letters", Stem: "Which TWO places need constant light?",
					Options: []entity.Option{{Key: "A", Text: "parks"}, {Key: "B", Text: "stations"}, {Key: "C", Text: "hospitals"}}},
			},
		}}
	ev := func(b, q string) *entity.Evidence { return &entity.Evidence{Block: b, Quote: q} }
	key := entity.Key{
		"1": {Answer: []string{"TRUE"}, Evidence: ev("A", "mainly because LEDs use half as much electricity"), Why: "Lower cost is the main reason."},
		"2": {Answer: []string{"NOT GIVEN"}, Why: "The passage does not say who paid."},
		"3": {Answer: []string{"30/thirty"}, Evidence: ev("B", "dimmed to thirty per cent"), Why: "A number is allowed."},
		"4": {Answer: []string{"sensors"}, Evidence: ev("B", "motion sensors raise them again"), Why: "One word only."},
		"5": {Answer: []string{"B", "C"}, Evidence: ev("B", "Stations and hospitals need constant light."), Why: "Both are named."},
		"6": {Answer: []string{"B", "C"}, Evidence: ev("B", "Stations and hospitals need constant light."), Why: "See 5."},
	}
	return p, key
}

func answers(m map[string]any) entity.Answers {
	out := entity.Answers{}
	for k, v := range m {
		b, _ := json.Marshal(v)
		out[k] = b
	}
	return out
}

func TestGradePart(t *testing.T) {
	p, key := samplePart()
	res := GradePart(p, key, answers(map[string]any{"1": "TRUE", "2": "FALSE", "3": "thirty", "4": "motion sensors", "5": []string{"c", "A", "c"}}), nil)
	if len(res) != 6 {
		t.Fatalf("results: %d", len(res))
	}
	got := map[int]entity.Result{}
	for _, r := range res {
		got[r.N] = r
	}
	if !got[1].Correct || got[2].Correct || !got[3].Correct || got[4].Correct || !got[4].OverLimit {
		t.Fatalf("marking: %+v", res)
	}
	// Choose TWO: C is right, A is wrong, the repeated C scores nothing.
	if !got[5].Correct || got[5].Given != "C" || got[6].Correct || got[6].Given != "A" {
		t.Fatalf("choose_n: %+v %+v", got[5], got[6])
	}
	if CountCorrect(res) != 3 {
		t.Fatalf("correct: %d", CountCorrect(res))
	}
	// A redo marks only the chosen questions; a non-string answer counts as blank.
	redo := GradePart(p, key, answers(map[string]any{"4": 7, "6": nil}), map[int]bool{4: true, 6: true})
	if len(redo) != 2 || redo[0].N != 4 || redo[0].Given != "" || redo[1].N != 6 {
		t.Fatalf("redo: %+v", redo)
	}
}

func TestValidatePart(t *testing.T) {
	p, key := samplePart()
	if err := ValidatePart(p, key); err != nil {
		t.Fatalf("valid part refused: %v", err)
	}
	broken := []struct {
		name   string
		mutate func(*entity.Part, entity.Key)
		want   string
	}{
		{"quote not in text", func(_ *entity.Part, k entity.Key) {
			k["1"] = entity.KeyItem{Answer: []string{"TRUE"}, Evidence: &entity.Evidence{Block: "A", Quote: "half the electricity"}, Why: "x"}
		}, "not found"},
		{"answer over the limit", func(_ *entity.Part, k entity.Key) {
			k["4"] = entity.KeyItem{Answer: []string{"motion sensors"}, Evidence: &entity.Evidence{Block: "B", Quote: "motion sensors"}, Why: "x"}
		}, "word limit"},
		{"missing key", func(_ *entity.Part, k entity.Key) { delete(k, "2") }, "no answer"},
		{"bad tfng value", func(_ *entity.Part, k entity.Key) {
			k["2"] = entity.KeyItem{Answer: []string{"NG"}, Why: "x"}
		}, "not one of"},
		{"gap in numbering", func(p *entity.Part, _ entity.Key) { p.Content.Groups[1].Lines = []string{"{3}", "no gap"} }, "shows questions"},
		{"choose_n sets differ", func(_ *entity.Part, k entity.Key) {
			k["6"] = entity.KeyItem{Answer: []string{"A", "C"}, Evidence: &entity.Evidence{Block: "B", Quote: "Stations"}, Why: "x"}
		}, "same 2 letters"},
		{"no explanation", func(_ *entity.Part, k entity.Key) {
			k["3"] = entity.KeyItem{Answer: []string{"30"}, Evidence: &entity.Evidence{Block: "B", Quote: "thirty"}}
		}, "no explanation"},
		{"key outside part", func(_ *entity.Part, k entity.Key) { k["9"] = entity.KeyItem{Answer: []string{"x"}, Why: "x"} }, "not a question"},
	}
	for _, b := range broken {
		p, key := samplePart()
		b.mutate(p, key)
		err := ValidatePart(p, key)
		if err == nil || !strings.Contains(err.Error(), b.want) {
			t.Errorf("%s: got %v, want an error containing %q", b.name, err, b.want)
		}
	}
}

func TestValidateTest(t *testing.T) {
	mk := func(no, from, to int) *entity.Part {
		return &entity.Part{Slug: "p", PartNo: no, QFrom: from, QTo: to}
	}
	if err := ValidateTest("s1-t1", entity.SkillReading, []*entity.Part{mk(2, 14, 26), mk(1, 1, 13), mk(3, 27, 40)}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTest("s1-t1", entity.SkillReading, []*entity.Part{mk(1, 1, 13), mk(2, 15, 26), mk(3, 27, 40)}); err == nil {
		t.Fatal("a numbering gap must be refused")
	}
	if err := ValidateTest("s1-t1", entity.SkillListening, []*entity.Part{mk(1, 1, 10)}); err == nil {
		t.Fatal("an incomplete test must be refused")
	}
}
