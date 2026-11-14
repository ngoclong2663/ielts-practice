package usecases

import (
	"context"
	"net/http"
	"strings"

	"ielts-practice/app/modules/users/domain/entity"
	"ielts-practice/app/modules/users/dto"
	"ielts-practice/pkg/exceptions"
	"ielts-practice/pkg/steps"
)

func invalidSteps(field string) *exceptions.AppError {
	appErr := exceptions.NewException(http.StatusBadRequest, exceptions.CodeUserInvalidSteps, "Invalid steps, use e.g. \"1m 10m\"", nil)
	appErr.Fields = map[string]string{field: exceptions.CodeUserInvalidSteps}
	return appErr
}

func (u *UserUseCases) UpdateStudySettings(ctx context.Context, userID string, req dto.UpdateStudySettingsRequestDto) (dto.StudySettingsDto, error) {
	if req.LearningSteps != nil {
		if _, err := steps.ParseMinutes(*req.LearningSteps); err != nil {
			return dto.StudySettingsDto{}, invalidSteps("learning_steps")
		}
	}
	if req.RelearningSteps != nil {
		if _, err := steps.ParseMinutes(*req.RelearningSteps); err != nil {
			return dto.StudySettingsDto{}, invalidSteps("relearning_steps")
		}
	}
	var result *entity.StudySettingsEntity
	err := u.transaction.Execute(ctx, func(txCtx context.Context) error {
		_, settings, err := u.GetStudyProfile(txCtx, userID)
		if err != nil {
			return err
		}
		if req.NewCardsPerDay != nil {
			settings.NewCardsPerDay = *req.NewCardsPerDay
		}
		if req.MaxReviewsPerDay != nil {
			settings.MaxReviewsPerDay = *req.MaxReviewsPerDay
		}
		if req.DesiredRetention != nil {
			settings.DesiredRetention = *req.DesiredRetention
		}
		if req.AudioAutoplay != nil {
			settings.AudioAutoplay = *req.AudioAutoplay
		}
		if req.Accent != nil {
			settings.Accent = *req.Accent
		}
		if req.ShowImageOnFront != nil {
			settings.ShowImageOnFront = *req.ShowImageOnFront
		}
		if req.LearningSteps != nil {
			settings.LearningSteps = strings.Join(strings.Fields(*req.LearningSteps), " ")
		}
		if req.RelearningSteps != nil {
			settings.RelearningSteps = strings.Join(strings.Fields(*req.RelearningSteps), " ")
		}
		if req.MaxIntervalDays != nil {
			settings.MaxIntervalDays = *req.MaxIntervalDays
		}
		if err := u.userRepo.UpdateStudySettings(txCtx, settings); err != nil {
			return exceptions.ErrInternalServerError("Failed to update study settings", err)
		}
		result = settings
		return nil
	})
	if err != nil {
		return dto.StudySettingsDto{}, err
	}
	return u.mapper.MapSettingsEntityToDto(result), nil
}
