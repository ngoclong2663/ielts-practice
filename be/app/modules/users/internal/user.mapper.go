package internal

import (
	"ielts-practice/app/models"
	"ielts-practice/app/modules/users/domain/entity"
	"ielts-practice/app/modules/users/dto"
)

type UserMapper struct{}

func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

func (m *UserMapper) MapUserEntityToModel(e *entity.UserEntity) *models.User {
	return &models.User{
		ID:           e.ID,
		GoogleSub:    e.GoogleSub,
		Email:        e.Email,
		DisplayName:  e.DisplayName,
		AvatarURL:    e.AvatarURL,
		UILocale:     e.UILocale,
		Timezone:     e.Timezone,
		DayStartHour: int16(e.DayStartHour),
		LastLoginAt:  e.LastLoginAt,
	}
}

func (m *UserMapper) MapUserModelToEntity(model *models.User) *entity.UserEntity {
	return &entity.UserEntity{
		ID:           model.ID,
		GoogleSub:    model.GoogleSub,
		Email:        model.Email,
		DisplayName:  model.DisplayName,
		AvatarURL:    model.AvatarURL,
		UILocale:     model.UILocale,
		Timezone:     model.Timezone,
		DayStartHour: int(model.DayStartHour),
		LastLoginAt:  model.LastLoginAt,
		CreatedAt:    model.CreatedAt,
	}
}

func (m *UserMapper) MapSettingsModelToEntity(model *models.UserStudySettings) *entity.StudySettingsEntity {
	return &entity.StudySettingsEntity{
		UserID:           model.UserID,
		NewCardsPerDay:   model.NewCardsPerDay,
		MaxReviewsPerDay: model.MaxReviewsPerDay,
		DesiredRetention: model.DesiredRetention,
		LearningSteps:    model.LearningSteps,
		RelearningSteps:  model.RelearningSteps,
		AudioAutoplay:    model.AudioAutoplay,
		Accent:           model.Accent,
		ShowImageOnFront: model.ShowImageOnFront,
	}
}

func (m *UserMapper) MapSettingsEntityToModel(e *entity.StudySettingsEntity) *models.UserStudySettings {
	return &models.UserStudySettings{
		UserID:           e.UserID,
		NewCardsPerDay:   e.NewCardsPerDay,
		MaxReviewsPerDay: e.MaxReviewsPerDay,
		DesiredRetention: e.DesiredRetention,
		LearningSteps:    e.LearningSteps,
		RelearningSteps:  e.RelearningSteps,
		AudioAutoplay:    e.AudioAutoplay,
		Accent:           e.Accent,
		ShowImageOnFront: e.ShowImageOnFront,
	}
}

func (m *UserMapper) MapUserEntityToProfileDto(e *entity.UserEntity) dto.UserProfileDto {
	return dto.UserProfileDto{
		ID:           e.ID,
		Email:        e.Email,
		DisplayName:  e.DisplayName,
		AvatarURL:    e.AvatarURL,
		UILocale:     e.UILocale,
		Timezone:     e.Timezone,
		DayStartHour: e.DayStartHour,
		CreatedAt:    e.CreatedAt,
	}
}

func (m *UserMapper) MapSettingsEntityToDto(e *entity.StudySettingsEntity) dto.StudySettingsDto {
	return dto.StudySettingsDto{
		NewCardsPerDay:   e.NewCardsPerDay,
		MaxReviewsPerDay: e.MaxReviewsPerDay,
		DesiredRetention: e.DesiredRetention,
		LearningSteps:    e.LearningSteps,
		RelearningSteps:  e.RelearningSteps,
		AudioAutoplay:    e.AudioAutoplay,
		Accent:           e.Accent,
		ShowImageOnFront: e.ShowImageOnFront,
	}
}
