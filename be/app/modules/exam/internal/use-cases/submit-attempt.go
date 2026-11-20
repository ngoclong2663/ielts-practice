package usecases

import (
	"context"
	"encoding/json"
	"slices"
	"strconv"
	"unicode/utf8"

	"ielts-practice/app/modules/exam/domain/entity"
	"ielts-practice/app/modules/exam/domain/rules"
	"ielts-practice/app/modules/exam/dto"
)

// maxAnswerLen caps one stored answer; real answers are a few words.
const maxAnswerLen = 120

// SubmitAttempt marks a finished run and stores it (Epic 10 §4, §7.3). The run itself (answers in progress,
// flags, highlights, timer) lived in the browser until now.
//   - part: any parts of one test; full: every part of the skill (a band is given); redo: some questions of
//     an earlier attempt, chosen by number.
func (uc *ExamUseCases) SubmitAttempt(ctx context.Context, userID string, req dto.SubmitAttemptRequestDto) (*dto.SubmitAttemptResponseDto, error) {
	t, err := uc.repo.FindTestBySlug(ctx, req.Test)
	if err != nil {
		return nil, internal("Failed to load test", err)
	}
	if t == nil {
		return nil, errTestNotFound()
	}
	all, err := uc.repo.ListParts(ctx, []string{t.ID}, req.Skill)
	if err != nil {
		return nil, internal("Failed to load parts", err)
	}
	if len(all) == 0 {
		return nil, errTestNotFound()
	}

	var chosen []entity.Part
	var only map[int]bool
	parentID := ""
	switch req.Kind {
	case entity.KindFull:
		chosen = all
	case entity.KindPart:
		if len(req.Parts) == 0 {
			return nil, errSelection("Choose at least one part")
		}
		for _, p := range all {
			if slices.Contains(req.Parts, p.PartNo) {
				chosen = append(chosen, p)
			}
		}
		if len(chosen) != len(uniqueInts(req.Parts)) {
			return nil, errSelection("A chosen part does not exist in this test")
		}
	case entity.KindRedo:
		if req.ParentAttemptID == "" {
			return nil, errSelection("A redo needs an earlier attempt of this test")
		}
		parent, err := uc.repo.FindAttempt(ctx, userID, req.ParentAttemptID)
		if err != nil {
			return nil, internal("Failed to load attempt", err)
		}
		if parent == nil || parent.TestID != t.ID || parent.Skill != req.Skill {
			return nil, errSelection("A redo needs an earlier attempt of this test")
		}
		qs := uniqueInts(req.Questions)
		if len(qs) == 0 {
			return nil, errSelection("Choose the questions to redo")
		}
		only = map[int]bool{}
		for _, n := range qs {
			if !slices.Contains(parent.Questions, n) {
				return nil, errSelection("A redo can only repeat questions of the earlier attempt")
			}
			only[n] = true
		}
		for _, p := range all {
			for n := p.QFrom; n <= p.QTo; n++ {
				if only[n] {
					chosen = append(chosen, p)
					break
				}
			}
		}
		parentID = parent.ID
	}

	partIDs := make([]string, 0, len(chosen))
	keys, err := uc.repo.FindKeys(ctx, idsOf(chosen))
	if err != nil {
		return nil, internal("Failed to load answer keys", err)
	}
	a := &entity.Attempt{UserID: userID, Skill: req.Skill, TestID: t.ID, Kind: req.Kind, ParentAttemptID: parentID,
		Mode: req.Mode, TimeLimitSec: req.TimeLimitSec, DurationSec: req.DurationSec, Answers: entity.Answers{},
		PartScores: map[string]entity.Score{}}
	if a.Mode == "" {
		a.Mode = entity.ModePractice
	}
	for i := range chosen {
		p := &chosen[i]
		partIDs = append(partIDs, p.ID)
		results := rules.GradePart(p, keys[p.ID], req.Answers, only)
		a.Results = append(a.Results, results...)
		a.PartScores[p.ID] = entity.Score{Correct: rules.CountCorrect(results), Total: len(results)}
		for _, r := range results {
			a.Questions = append(a.Questions, r.N)
		}
	}
	a.PartIDs = partIDs
	a.Correct, a.Total = rules.CountCorrect(a.Results), len(a.Results)
	a.Answers = keepAnswers(req.Answers, chosen, a.Questions)
	if a.Kind == entity.KindFull && a.Total == rules.FullTestQuestions {
		a.Band = rules.Band(a.Skill, a.Correct)
	}
	if err := uc.repo.CreateAttempt(ctx, a); err != nil {
		return nil, internal("Failed to save attempt", err)
	}
	return &dto.SubmitAttemptResponseDto{ID: a.ID}, nil
}

// keepAnswers stores only the answers of the marked questions (a choose_n group's answer is keyed by its
// first question), each capped in length.
func keepAnswers(in entity.Answers, parts []entity.Part, questions []int) entity.Answers {
	out := entity.Answers{}
	keep := map[string]bool{}
	for _, n := range questions {
		keep[strconv.Itoa(n)] = true
	}
	for _, p := range parts {
		for _, g := range p.Content.Groups {
			if g.Type == entity.TypeChooseN {
				for n := g.From; n <= g.To; n++ {
					if keep[strconv.Itoa(n)] {
						keep[strconv.Itoa(g.From)] = true
					}
				}
			}
		}
	}
	for k, raw := range in {
		if !keep[k] {
			continue
		}
		var s string
		if json.Unmarshal(raw, &s) == nil {
			if utf8.RuneCountInString(s) > maxAnswerLen {
				s = string([]rune(s)[:maxAnswerLen])
			}
			b, _ := json.Marshal(s)
			out[k] = b
			continue
		}
		var list []string
		if json.Unmarshal(raw, &list) == nil && len(list) <= 5 {
			b, _ := json.Marshal(list)
			out[k] = b
		}
	}
	return out
}

func idsOf(parts []entity.Part) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, p.ID)
	}
	return out
}

func uniqueInts(in []int) []int {
	var out []int
	for _, n := range in {
		if !slices.Contains(out, n) {
			out = append(out, n)
		}
	}
	return out
}
