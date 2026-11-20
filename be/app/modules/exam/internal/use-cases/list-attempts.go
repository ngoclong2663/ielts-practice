package usecases

import (
	"context"
	"slices"

	"ielts-practice/app/modules/exam/domain/entity"
	"ielts-practice/app/modules/exam/dto"
)

const attemptsPageSize = 20

// ListAttempts is the learner's history for one skill, newest first, 20 per page.
func (uc *ExamUseCases) ListAttempts(ctx context.Context, userID string, q dto.ListAttemptsQueryDto) (*dto.AttemptListDto, error) {
	attempts, err := uc.repo.ListAttempts(ctx, userID, q.Skill, attemptsPageSize+1, q.Before, false)
	if err != nil {
		return nil, internal("Failed to load attempts", err)
	}
	out := &dto.AttemptListDto{Items: []dto.AttemptSummaryDto{}}
	if len(attempts) > attemptsPageSize {
		attempts = attempts[:attemptsPageSize]
		out.NextBefore = attempts[len(attempts)-1].ID
	}
	tests := map[string]*entity.Test{}
	books := map[string]entity.Localized{}
	var partIDs []string
	for _, a := range attempts {
		partIDs = append(partIDs, a.PartIDs...)
	}
	parts, err := uc.repo.ListPartsByIDs(ctx, partIDs)
	if err != nil {
		return nil, internal("Failed to load parts", err)
	}
	for i := range attempts {
		a := &attempts[i]
		t, ok := tests[a.TestID]
		if !ok {
			if t, err = uc.repo.FindTestByID(ctx, a.TestID); err != nil || t == nil {
				return nil, internal("Failed to load test", err)
			}
			tests[a.TestID] = t
		}
		title, ok := books[t.BookID]
		if !ok {
			b, err := uc.repo.FindBookByID(ctx, t.BookID)
			if err != nil || b == nil {
				return nil, internal("Failed to load book", err)
			}
			title, books[t.BookID] = b.Title, b.Title
		}
		var mine []entity.Part
		for _, p := range parts {
			if slices.Contains(a.PartIDs, p.ID) {
				mine = append(mine, p)
			}
		}
		out.Items = append(out.Items, summaryDto(a, t, title, mine))
	}
	return out, nil
}
