package usecases

import (
	"context"

	"ielts-practice/app/modules/vocab/dto"
)

func (uc *VocabUseCases) GetJob(ctx context.Context, userID, jobID string) (dto.JobDto, error) {
	job, err := uc.repo.FindJob(ctx, userID, jobID)
	if err != nil {
		return dto.JobDto{}, internalErr("Failed to load generation", err)
	}
	if job == nil {
		return dto.JobDto{}, jobNotFound()
	}
	return uc.mapper.JobEntityToDto(job), nil
}

func (uc *VocabUseCases) ListJobs(ctx context.Context, userID string) (dto.JobListDto, error) {
	jobs, err := uc.repo.ListJobs(ctx, userID, 10)
	if err != nil {
		return dto.JobListDto{}, internalErr("Failed to load generations", err)
	}
	out := dto.JobListDto{Items: make([]dto.JobSummaryDto, 0, len(jobs))}
	for i := range jobs {
		out.Items = append(out.Items, uc.mapper.JobEntityToSummaryDto(&jobs[i]))
	}
	return out, nil
}
