package usecases

import (
	"context"
	"net/http"

	"ielts-practice/app/modules/writing/domain/entity"
	"ielts-practice/app/modules/writing/domain/rules"
	"ielts-practice/app/modules/writing/dto"
	"ielts-practice/pkg/exceptions"
)

// GetVocab returns the useful vocabulary of the attempt's task (Epic 7; per task since Epic 8). Seeded tasks
// have a reviewed set whose words are ready-made cards; otherwise it is written on first request (one shared AI
// call, from the attempt's lesson; concurrent requests wait for it) and reused by every learner of the task.
func (uc *WritingUseCases) GetVocab(ctx context.Context, userID, attemptID string) (dto.VocabDto, error) {
	attempt, task, lesson, err := uc.loadAttempt(ctx, userID, attemptID)
	if err != nil {
		return dto.VocabDto{}, err
	}
	if !rules.VocabVisible(attempt.Level, attempt.Status == entity.StatusSubmitted) {
		return dto.VocabDto{}, appErr(http.StatusForbidden, exceptions.CodeWritingVocabLocked,
			"Free writing shows the vocabulary after you submit")
	}
	if v, err := uc.repo.FindTaskVocab(ctx, task.ID); err != nil || v != nil {
		if err != nil {
			return dto.VocabDto{}, internalErr("Failed to load the vocabulary", err)
		}
		return dto.VocabDto{Groups: v.Groups}, nil
	}
	v, err, _ := uc.lessons.Do("vocab|"+task.ID, func() (any, error) {
		// Detached from the request, like lessons: the result is shared, so a closed tab should not waste it.
		gctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), vocabTimeout)
		defer cancel()
		content, usage, err := uc.coach.WriteVocab(gctx, entity.VocabRequest{
			Task: task, Position: lesson.Position, ModelAnswer: modelAnswer(lesson.Content),
		})
		if err != nil {
			return nil, coachErr(err)
		}
		uc.record(gctx, userID, &attempt.ID, entity.KindVocab, usage)
		stored, err := uc.repo.InsertTaskVocab(gctx, &entity.StoredVocab{TaskID: task.ID, Content: content, Usage: usage})
		if err != nil {
			return nil, internalErr("Failed to store the vocabulary", err)
		}
		uc.logger.Info().Str("task", task.Slug).Str("lesson", lesson.ID).Str("model", usage.Model).
			Int("output_tokens", usage.OutputTokens).Msg("Writing vocabulary created")
		return stored, nil
	})
	if err != nil {
		return dto.VocabDto{}, err
	}
	return dto.VocabDto{Groups: v.(*entity.LessonVocab).Groups}, nil
}
