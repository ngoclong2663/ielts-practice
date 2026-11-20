package usecases

import (
	"context"

	"ielts-practice/app/modules/exam/domain/entity"
	"ielts-practice/app/modules/exam/dto"
)

// GetAttempt returns a submitted attempt with its marking and, for its parts only, the content with the key,
// explanations and (listening) transcript, for the result and review screens.
func (uc *ExamUseCases) GetAttempt(ctx context.Context, userID, id string) (*dto.AttemptDto, error) {
	a, err := uc.repo.FindAttempt(ctx, userID, id)
	if err != nil {
		return nil, internal("Failed to load attempt", err)
	}
	if a == nil {
		return nil, errAttemptNotFound()
	}
	t, err := uc.repo.FindTestByID(ctx, a.TestID)
	if err != nil || t == nil {
		return nil, internal("Failed to load test", err)
	}
	parts, err := uc.repo.ListPartsByIDs(ctx, a.PartIDs)
	if err != nil {
		return nil, internal("Failed to load parts", err)
	}
	keys, err := uc.repo.FindKeys(ctx, a.PartIDs)
	if err != nil {
		return nil, internal("Failed to load answer keys", err)
	}
	test, err := uc.testDto(ctx, t, a.Skill, parts, keys)
	if err != nil {
		return nil, err
	}
	out := &dto.AttemptDto{
		AttemptSummaryDto: summaryDto(a, t, test.BookTitle, parts),
		ParentAttemptID:   a.ParentAttemptID,
		Questions:         a.Questions,
		Answers:           a.Answers,
		Results:           a.Results,
		Test:              *test,
	}
	if out.Questions == nil {
		out.Questions = []int{}
	}
	if out.Results == nil {
		out.Results = []entity.Result{}
	}
	if out.Answers == nil {
		out.Answers = entity.Answers{}
	}
	return out, nil
}

func summaryDto(a *entity.Attempt, t *entity.Test, bookTitle entity.Localized, parts []entity.Part) dto.AttemptSummaryDto {
	nos := []int{}
	for _, p := range parts {
		nos = append(nos, p.PartNo)
	}
	return dto.AttemptSummaryDto{ID: a.ID, Skill: a.Skill, TestSlug: t.Slug, TestNo: t.TestNo, BookTitle: bookTitle, PartNos: nos,
		Kind: a.Kind, Mode: a.Mode, Correct: a.Correct, Total: a.Total, Band: a.Band, DurationSec: a.DurationSec,
		TimeLimitSec: a.TimeLimitSec, CreatedAt: a.CreatedAt}
}
