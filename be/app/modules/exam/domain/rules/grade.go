package rules

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"

	"ielts-practice/app/modules/exam/domain/entity"
)

// GradePart marks the learner's answers for one part. only limits marking to some questions (a redo run);
// nil marks them all. Results are in question order.
func GradePart(p *entity.Part, key entity.Key, answers entity.Answers, only map[int]bool) []entity.Result {
	var out []entity.Result
	include := func(n int) bool { return only == nil || only[n] }
	for _, g := range p.Content.Groups {
		mode := entity.AnswerMode(g.Type)
		if mode == entity.AnswerChooseN {
			out = append(out, gradeChooseN(p.PartNo, g, key, answers, include)...)
			continue
		}
		for n := g.From; n <= g.To; n++ {
			if !include(n) {
				continue
			}
			given := answerString(answers, n)
			r := entity.Result{N: n, PartNo: p.PartNo, Type: g.Type, Given: given}
			k := key[strconv.Itoa(n)]
			if mode == entity.AnswerGap {
				if given != "" && !WithinLimit(given, g.WordLimit, g.AllowNumber) {
					r.OverLimit = true
				} else {
					r.Correct = MatchesAny(given, k.Answer)
				}
			} else {
				r.Correct = given != "" && slices.ContainsFunc(k.Answer, func(a string) bool { return strings.EqualFold(strings.TrimSpace(a), given) })
			}
			out = append(out, r)
		}
	}
	return out
}

// gradeChooseN marks a "Choose TWO letters" group: each correct letter scores one of its questions, in any
// order; extra or repeated letters score nothing.
func gradeChooseN(partNo int, g entity.Group, key entity.Key, answers entity.Answers, include func(int) bool) []entity.Result {
	var chosen []string
	if raw, ok := answers[strconv.Itoa(g.From)]; ok {
		_ = json.Unmarshal(raw, &chosen)
	}
	correctSet := key[strconv.Itoa(g.From)].Answer
	var hits, misses []string
	seen := map[string]bool{}
	for _, c := range chosen {
		c = strings.ToUpper(strings.TrimSpace(c))
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		if slices.Contains(correctSet, c) {
			hits = append(hits, c)
		} else {
			misses = append(misses, c)
		}
	}
	var out []entity.Result
	for n := g.From; n <= g.To; n++ {
		i := n - g.From
		r := entity.Result{N: n, PartNo: partNo, Type: g.Type}
		switch {
		case i < len(hits):
			r.Given, r.Correct = hits[i], true
		case i-len(hits) < len(misses):
			r.Given = misses[i-len(hits)]
		}
		if include(n) {
			out = append(out, r)
		}
	}
	return out
}

// answerString reads one question's answer; anything that is not a JSON string counts as blank.
func answerString(answers entity.Answers, n int) string {
	raw, ok := answers[strconv.Itoa(n)]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return strings.TrimSpace(s)
}

// CountCorrect returns the number of correct results.
func CountCorrect(results []entity.Result) int {
	c := 0
	for _, r := range results {
		if r.Correct {
			c++
		}
	}
	return c
}
