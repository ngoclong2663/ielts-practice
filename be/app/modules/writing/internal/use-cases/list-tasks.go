package usecases

import (
	"context"

	"ielts-practice/app/modules/writing/domain"
	"ielts-practice/app/modules/writing/dto"
)

func (uc *WritingUseCases) ListTasks(ctx context.Context, userID string, q dto.TaskListQueryDto) ([]dto.TaskSummaryDto, error) {
	tasks, err := uc.repo.ListTasks(ctx, domain.TaskFilter{TaskNo: q.TaskNo, TaskType: q.Type, Topic: q.Topic})
	if err != nil {
		return nil, internalErr("Failed to load tasks", err)
	}
	progress, err := uc.repo.TaskProgress(ctx, userID)
	if err != nil {
		return nil, internalErr("Failed to load progress", err)
	}
	out := make([]dto.TaskSummaryDto, 0, len(tasks))
	for i := range tasks {
		out = append(out, taskSummary(&tasks[i], progress[tasks[i].ID]))
	}
	return out, nil
}

func (uc *WritingUseCases) GetTask(ctx context.Context, userID, taskID string) (dto.TaskDetailDto, error) {
	task, err := uc.repo.FindTaskByID(ctx, taskID)
	if err != nil {
		return dto.TaskDetailDto{}, internalErr("Failed to load task", err)
	}
	if task == nil {
		return dto.TaskDetailDto{}, taskNotFound()
	}
	history, err := uc.repo.ListAttempts(ctx, userID, taskID, historyLimit)
	if err != nil {
		return dto.TaskDetailDto{}, internalErr("Failed to load attempts", err)
	}
	progress := domain.TaskProgress{Attempts: len(history)}
	out := dto.TaskDetailDto{History: make([]dto.AttemptSummaryDto, 0, len(history))}
	for _, a := range history {
		out.History = append(out.History, attemptSummary(a))
		if a.BandHigh != nil && (progress.BestBand == nil || *a.BandHigh > *progress.BestBand) {
			progress.BestBand = a.BandHigh
		}
	}
	out.TaskDto = taskDto(task, progress)
	return out, nil
}
