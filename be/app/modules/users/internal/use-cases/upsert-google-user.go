package usecases

import (
	"context"

	"ielts-practice/app/modules/users/domain/entity"
	usersin "ielts-practice/app/modules/users/ports/in"
	"ielts-practice/pkg/exceptions"
)

func (u *UserUseCases) UpsertGoogleUser(ctx context.Context, input usersin.GoogleUserInput) (*entity.UserEntity, error) {
	now := u.now().UTC()

	existing, err := u.userRepo.FindByGoogleSub(ctx, input.GoogleSub)
	if err != nil {
		return nil, exceptions.ErrInternalServerError("Failed to load user", err)
	}

	if existing != nil {
		existing.Email = input.Email
		if input.DisplayName != "" {
			existing.DisplayName = input.DisplayName
		}
		existing.AvatarURL = input.AvatarURL
		existing.LastLoginAt = &now
		if err := u.userRepo.Update(ctx, existing); err != nil {
			return nil, exceptions.ErrInternalServerError("Failed to update user", err)
		}
		return existing, nil
	}

	locale := input.UILocale
	if locale != "vi" && locale != "en" {
		locale = defaultLocale
	}
	timezone := input.Timezone
	if !IsValidTimezone(timezone) {
		timezone = defaultTimezone
	}

	created, err := u.userRepo.Create(ctx, &entity.UserEntity{
		GoogleSub:    input.GoogleSub,
		Email:        input.Email,
		DisplayName:  input.DisplayName,
		AvatarURL:    input.AvatarURL,
		UILocale:     locale,
		Timezone:     timezone,
		DayStartHour: 4,
		LastLoginAt:  &now,
	})
	if err != nil {
		return nil, exceptions.ErrInternalServerError("Failed to create user", err)
	}
	if err := u.userRepo.CreateDefaultStudySettings(ctx, created.ID); err != nil {
		return nil, exceptions.ErrInternalServerError("Failed to create study settings", err)
	}
	return created, nil
}
