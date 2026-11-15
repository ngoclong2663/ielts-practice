package usecases

import (
	"context"
	"net/http"
	"slices"

	"ielts-practice/app/modules/writing/domain"
	"ielts-practice/app/modules/writing/domain/entity"
	"ielts-practice/app/modules/writing/dto"
	"ielts-practice/pkg/exceptions"
)

// StartAttempt begins a guided attempt. The lesson for the task and position is written on first use
// (one shared AI call; concurrent starts wait for it) and reused by every learner afterwards.
func (uc *WritingUseCases) StartAttempt(ctx context.Context, userID, taskID string, req dto.StartAttemptRequestDto) (dto.AttemptDto, error) {
	task, err := uc.repo.FindTaskByID(ctx, taskID)
	if err != nil {
		return dto.AttemptDto{}, internalErr("Failed to load task", err)
	}
	if task == nil {
		return dto.AttemptDto{}, taskNotFound()
	}
	position := req.Position
	if task.TaskNo == 1 {
		position = ""
	} else if !slices.Contains(domain.Positions(task), position) {
		return dto.AttemptDto{}, appErr(http.StatusBadRequest, exceptions.CodeWritingInvalidPosition, "Choose one of the task's positions")
	}

	lesson, err := uc.lesson(ctx, userID, task, position)
	if err != nil {
		return dto.AttemptDto{}, err
	}
	attempt := &entity.Attempt{UserID: userID, TaskID: task.ID, LessonID: lesson.ID, Level: req.Level, Position: position}
	if err := uc.repo.CreateAttempt(ctx, attempt); err != nil {
		return dto.AttemptDto{}, internalErr("Failed to start the attempt", err)
	}
	return uc.attemptView(ctx, attempt, task, lesson)
}

func (uc *WritingUseCases) lesson(ctx context.Context, userID string, task *entity.Task, position string) (*entity.Lesson, error) {
	if l, err := uc.repo.FindLesson(ctx, task.ID, position, task.ContentHash); err != nil || l != nil {
		if err != nil {
			return nil, internalErr("Failed to load the lesson", err)
		}
		return l, nil
	}
	key := task.ID + "|" + position + "|" + task.ContentHash
	v, err, _ := uc.lessons.Do(key, func() (any, error) {
		// Detached from the request: a learner closing the tab should not waste a lesson others will use.
		gctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), lessonTimeout)
		defer cancel()
		content, usage, err := uc.coach.WriteLesson(gctx, domain.LessonRequest{Task: task, Position: position})
		if err != nil {
			return nil, coachErr(err)
		}
		uc.record(gctx, userID, nil, domain.KindLesson, usage)
		stored, err := uc.repo.InsertLesson(gctx, &entity.Lesson{
			TaskID: task.ID, Position: position, TaskHash: task.ContentHash, Content: content,
			Model: usage.Model, InputTokens: usage.InputTokens, OutputTokens: usage.OutputTokens,
		})
		if err != nil {
			return nil, internalErr("Failed to store the lesson", err)
		}
		uc.logger.Info().Str("task", task.Slug).Str("position", position).Str("model", usage.Model).
			Int("output_tokens", usage.OutputTokens).Msg("Writing lesson created")
		return stored, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*entity.Lesson), nil
}
