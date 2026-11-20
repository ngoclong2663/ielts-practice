package usecases

import (
	"context"

	"ielts-practice/app/modules/exam/dto"
)

// GetTest returns every part of one skill of a test for the player, without the key or the transcript. The
// player shows the parts the learner chose.
func (uc *ExamUseCases) GetTest(ctx context.Context, slug, skill string) (*dto.TestDto, error) {
	t, err := uc.repo.FindTestBySlug(ctx, slug)
	if err != nil {
		return nil, internal("Failed to load test", err)
	}
	if t == nil {
		return nil, errTestNotFound()
	}
	parts, err := uc.repo.ListParts(ctx, []string{t.ID}, skill)
	if err != nil {
		return nil, internal("Failed to load parts", err)
	}
	if len(parts) == 0 {
		return nil, errTestNotFound()
	}
	return uc.testDto(ctx, t, skill, parts, nil)
}
