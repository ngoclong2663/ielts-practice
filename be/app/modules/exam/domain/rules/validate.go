package rules

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"ielts-practice/app/modules/exam/domain/entity"
)

var placeholder = regexp.MustCompile(`\{(\d+)\}`)

// PartsPerSkill is how many parts a full test has.
func PartsPerSkill(skill string) int {
	if skill == entity.SkillListening {
		return 4
	}
	return 3
}

// Placeholders returns the question numbers written as {n} in a text, in order.
func Placeholders(s string) []int {
	var out []int
	for _, m := range placeholder.FindAllStringSubmatch(s, -1) {
		n, _ := strconv.Atoi(m[1])
		out = append(out, n)
	}
	return out
}

// groupQuestions lists the question numbers a group shows: its items, and the {n} gaps in its stem, lines
// and table cells.
func groupQuestions(g entity.Group) []int {
	var out []int
	for _, it := range g.Items {
		out = append(out, it.N)
	}
	out = append(out, Placeholders(g.Stem)...)
	for _, l := range g.Lines {
		out = append(out, Placeholders(l)...)
	}
	for _, row := range g.Rows {
		for _, cell := range row {
			out = append(out, Placeholders(cell)...)
		}
	}
	return out
}

// QuestionTypes lists a part's group types once each, in order.
func QuestionTypes(c entity.Content) []string {
	var out []string
	for _, g := range c.Groups {
		if !slices.Contains(out, g.Type) {
			out = append(out, g.Type)
		}
	}
	return out
}

// ValidatePart checks one part and its key (Epic 10 §8.2): numbering, structure of every group, an answer
// and an explanation for every question, answers that fit their group, and evidence quoted word for word.
func ValidatePart(p *entity.Part, key entity.Key) error {
	fail := func(format string, a ...any) error { return fmt.Errorf("%s: "+format, append([]any{p.Slug}, a...)...) }
	if !slices.Contains(entity.Skills, p.Skill) {
		return fail("unknown skill %q", p.Skill)
	}
	if p.PartNo < 1 || p.PartNo > PartsPerSkill(p.Skill) {
		return fail("part %d out of range", p.PartNo)
	}
	if strings.TrimSpace(p.Title) == "" {
		return fail("title is required")
	}
	c := p.Content
	blocks := map[string]string{}
	if p.Skill == entity.SkillReading {
		if len(c.Paragraphs) == 0 || len(c.Lines) != 0 {
			return fail("a reading part needs paragraphs and no transcript lines")
		}
		for _, para := range c.Paragraphs {
			if para.ID == "" || strings.TrimSpace(para.Text) == "" || blocks[para.ID] != "" {
				return fail("paragraph %q is empty or duplicated", para.ID)
			}
			blocks[para.ID] = para.Text
		}
	} else {
		if len(c.Lines) == 0 || len(c.Paragraphs) != 0 {
			return fail("a listening part needs transcript lines and no paragraphs")
		}
		last := -1.0
		for _, l := range c.Lines {
			if l.ID == "" || strings.TrimSpace(l.Text) == "" || blocks[l.ID] != "" {
				return fail("line %q is empty or duplicated", l.ID)
			}
			if l.At < last {
				return fail("line %s starts before the line above it", l.ID)
			}
			last = l.At
			blocks[l.ID] = l.Text
		}
		if p.Audio == nil || p.Audio.Key == "" || p.Audio.Duration <= 0 {
			return fail("a listening part needs audio")
		}
		if last >= p.Audio.Duration {
			return fail("the transcript runs past the end of the audio")
		}
		for _, cue := range p.Audio.Cues {
			if cue.At < 0 || cue.At > p.Audio.Duration || cue.Seconds <= 0 || cue.Label == "" {
				return fail("cue at %.1fs is outside the audio or incomplete", cue.At)
			}
			if _, ok := blocks[cue.After]; cue.After != "" && !ok {
				return fail("cue %q follows line %q, which does not exist", cue.Label, cue.After)
			}
		}
	}

	if len(c.Groups) == 0 {
		return fail("no question groups")
	}
	next := p.QFrom
	for gi, g := range c.Groups {
		if err := validateGroup(g, next); err != nil {
			return fail("group %d (%s): %v", gi+1, g.Type, err)
		}
		next = g.To + 1
	}
	if next-1 != p.QTo {
		return fail("groups end at question %d, the part ends at %d", next-1, p.QTo)
	}

	for k := range key {
		n, err := strconv.Atoi(k)
		if err != nil || n < p.QFrom || n > p.QTo {
			return fail("key %q is not a question of this part", k)
		}
	}
	for _, g := range c.Groups {
		for n := g.From; n <= g.To; n++ {
			if err := validateKeyItem(g, key, n, blocks); err != nil {
				return fail("question %d: %v", n, err)
			}
		}
	}
	return nil
}

func validateGroup(g entity.Group, from int) error {
	mode := entity.AnswerMode(g.Type)
	if mode == "" {
		return fmt.Errorf("unknown type")
	}
	if g.From != from || g.To < g.From {
		return fmt.Errorf("numbered %d–%d, expected to start at %d", g.From, g.To, from)
	}
	if strings.TrimSpace(g.Instructions) == "" {
		return fmt.Errorf("instructions are required")
	}
	shown := groupQuestions(g)
	if mode == entity.AnswerChooseN {
		if g.Pick != g.To-g.From+1 || g.Stem == "" || len(g.Options) <= g.Pick {
			return fmt.Errorf("choose_n needs a stem, pick = number of questions and more options than pick")
		}
	} else {
		want := make([]int, 0, g.To-g.From+1)
		for n := g.From; n <= g.To; n++ {
			want = append(want, n)
		}
		got := slices.Clone(shown)
		slices.Sort(got)
		if !slices.Equal(got, want) {
			return fmt.Errorf("shows questions %v, expected %v", shown, want)
		}
	}
	switch mode {
	case entity.AnswerGap:
		if g.WordLimit == 0 && !g.AllowNumber {
			return fmt.Errorf("a gap group needs a word limit")
		}
	case entity.AnswerLetter:
		if len(g.Options) < 2 {
			return fmt.Errorf("needs options")
		}
	case entity.AnswerChoice:
		if g.Type == entity.TypeMCQ {
			for _, it := range g.Items {
				if len(it.Options) < 3 {
					return fmt.Errorf("question %d needs at least 3 options", it.N)
				}
			}
		}
	}
	keys := map[string]bool{}
	for _, o := range g.Options {
		if o.Key == "" || keys[o.Key] {
			return fmt.Errorf("option %q is empty or duplicated", o.Key)
		}
		keys[o.Key] = true
	}
	return nil
}

func validateKeyItem(g entity.Group, key entity.Key, n int, blocks map[string]string) error {
	k, ok := key[strconv.Itoa(n)]
	if !ok || len(k.Answer) == 0 {
		return fmt.Errorf("no answer in the key")
	}
	if strings.TrimSpace(k.Why) == "" {
		return fmt.Errorf("no explanation")
	}
	notGiven := false
	switch entity.AnswerMode(g.Type) {
	case entity.AnswerChoice:
		allowed := optionKeys(g, n)
		for _, a := range k.Answer {
			if !slices.Contains(allowed, a) {
				return fmt.Errorf("answer %q is not one of %v", a, allowed)
			}
			notGiven = notGiven || a == "NOT GIVEN"
		}
	case entity.AnswerLetter:
		for _, a := range k.Answer {
			if !slices.Contains(optionKeys(g, n), a) {
				return fmt.Errorf("answer %q is not an option", a)
			}
		}
	case entity.AnswerChooseN:
		first := key[strconv.Itoa(g.From)].Answer
		if len(k.Answer) != g.Pick || !sameSet(k.Answer, first) {
			return fmt.Errorf("every question of a choose_n group needs the same %d letters", g.Pick)
		}
		for _, a := range k.Answer {
			if !slices.Contains(optionKeys(g, n), a) {
				return fmt.Errorf("answer %q is not an option", a)
			}
		}
	case entity.AnswerGap:
		for _, entry := range k.Answer {
			vs := ExpandVariants(entry)
			if len(vs) == 0 {
				return fmt.Errorf("empty answer")
			}
			for _, v := range vs {
				if !WithinLimit(v, g.WordLimit, g.AllowNumber) {
					return fmt.Errorf("answer %q breaks the group's word limit", v)
				}
			}
		}
	}
	if k.Evidence == nil {
		if notGiven {
			return nil
		}
		return fmt.Errorf("evidence is required (only NOT GIVEN may go without)")
	}
	text, ok := blocks[k.Evidence.Block]
	if !ok {
		return fmt.Errorf("evidence block %q does not exist", k.Evidence.Block)
	}
	if k.Evidence.Quote == "" || !strings.Contains(text, k.Evidence.Quote) {
		return fmt.Errorf("evidence quote not found word for word in %s: %q", k.Evidence.Block, k.Evidence.Quote)
	}
	return nil
}

func optionKeys(g entity.Group, n int) []string {
	switch g.Type {
	case entity.TypeTFNG:
		return []string{"TRUE", "FALSE", "NOT GIVEN"}
	case entity.TypeYNNG:
		return []string{"YES", "NO", "NOT GIVEN"}
	}
	var out []string
	opts := g.Options
	for _, it := range g.Items {
		if it.N == n && len(it.Options) > 0 {
			opts = it.Options
		}
	}
	for _, o := range opts {
		out = append(out, o.Key)
	}
	return out
}

func sameSet(a, b []string) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	slices.Sort(x)
	slices.Sort(y)
	return slices.Equal(x, y)
}

// ValidateTest checks that a test's parts of one skill are complete and numbered 1–40 without gaps.
func ValidateTest(testSlug, skill string, parts []*entity.Part) error {
	if len(parts) == 0 {
		return nil
	}
	if len(parts) != PartsPerSkill(skill) {
		return fmt.Errorf("%s %s: has %d parts, a full test has %d", testSlug, skill, len(parts), PartsPerSkill(skill))
	}
	sorted := slices.Clone(parts)
	slices.SortFunc(sorted, func(a, b *entity.Part) int { return a.PartNo - b.PartNo })
	next := 1
	for i, p := range sorted {
		if p.PartNo != i+1 {
			return fmt.Errorf("%s %s: part numbers must be 1..%d", testSlug, skill, len(parts))
		}
		if p.QFrom != next {
			return fmt.Errorf("%s: starts at question %d, expected %d", p.Slug, p.QFrom, next)
		}
		next = p.QTo + 1
	}
	if next-1 != FullTestQuestions {
		return fmt.Errorf("%s %s: questions end at %d, expected %d", testSlug, skill, next-1, FullTestQuestions)
	}
	return nil
}
