package usecases

import (
	"context"

	"ielts-practice/app/modules/exam/domain/entity"
	"ielts-practice/app/modules/exam/dto"
)

// ListBooks returns the test bank for one skill: books → tests → parts, with the learner's best score per
// part. Tests with no part of that skill are left out, and so are books with no such test.
func (uc *ExamUseCases) ListBooks(ctx context.Context, userID, skill string) ([]dto.BookDto, error) {
	books, err := uc.repo.ListPublishedBooks(ctx)
	if err != nil {
		return nil, internal("Failed to load books", err)
	}
	ids := make([]string, 0, len(books))
	for _, b := range books {
		ids = append(ids, b.ID)
	}
	tests, err := uc.repo.ListTests(ctx, ids)
	if err != nil {
		return nil, internal("Failed to load tests", err)
	}
	testIDs := make([]string, 0, len(tests))
	for _, t := range tests {
		testIDs = append(testIDs, t.ID)
	}
	parts, err := uc.repo.ListParts(ctx, testIDs, skill)
	if err != nil {
		return nil, internal("Failed to load parts", err)
	}
	best, err := uc.bestScores(ctx, userID, skill)
	if err != nil {
		return nil, err
	}

	partsOf := map[string][]dto.PartSummaryDto{}
	for _, p := range parts {
		s := dto.PartSummaryDto{PartNo: p.PartNo, Title: p.Title, QFrom: p.QFrom, QTo: p.QTo, QuestionTypes: p.QuestionTypes}
		if s.QuestionTypes == nil {
			s.QuestionTypes = []string{}
		}
		if b, ok := best[p.ID]; ok {
			s.Best = &b
		}
		partsOf[p.TestID] = append(partsOf[p.TestID], s)
	}
	testsOf := map[string][]dto.TestSummaryDto{}
	for _, t := range tests {
		if ps := partsOf[t.ID]; len(ps) > 0 {
			testsOf[t.BookID] = append(testsOf[t.BookID], dto.TestSummaryDto{Slug: t.Slug, TestNo: t.TestNo, Parts: ps})
		}
	}
	out := []dto.BookDto{}
	for _, b := range books {
		if ts := testsOf[b.ID]; len(ts) > 0 {
			out = append(out, dto.BookDto{Slug: b.Slug, Title: b.Title, Tests: ts})
		}
	}
	return out, nil
}

// bestScores is the learner's best score per part over part and full-test attempts (redo runs excluded).
func (uc *ExamUseCases) bestScores(ctx context.Context, userID, skill string) (map[string]entity.Score, error) {
	attempts, err := uc.repo.ListAttempts(ctx, userID, skill, 0, "", false)
	if err != nil {
		return nil, internal("Failed to load attempts", err)
	}
	best := map[string]entity.Score{}
	for _, a := range attempts {
		if a.Kind == entity.KindRedo {
			continue
		}
		for partID, s := range a.PartScores {
			if cur, ok := best[partID]; !ok || s.Correct > cur.Correct {
				best[partID] = s
			}
		}
	}
	return best, nil
}
