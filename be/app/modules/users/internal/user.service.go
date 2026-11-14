package internal

import (
	"context"
	"time"

	"ielts-practice/app/modules/users/domain"
	"ielts-practice/app/modules/users/domain/entity"
	"ielts-practice/app/modules/users/ports"
	"ielts-practice/pkg/exceptions"
)

var _ ports.IUserService = &UserService{}

const (
	defaultLocale   = "en"
	defaultTimezone = "Asia/Ho_Chi_Minh"
)

type UserService struct {
	userRepo domain.IUserRepository
	now      func() time.Time
}

func NewUserService(userRepo domain.IUserRepository) ports.IUserService {
	return &UserService{userRepo: userRepo, now: time.Now}
}

func (us *UserService) UpsertGoogleUser(ctx context.Context, input ports.GoogleUserInput) (*entity.UserEntity, error) {
	now := us.now().UTC()

	existing, err := us.userRepo.FindByGoogleSub(ctx, input.GoogleSub)
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
		if err := us.userRepo.Update(ctx, existing); err != nil {
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

	created, err := us.userRepo.Create(ctx, &entity.UserEntity{
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
	if err := us.userRepo.CreateDefaultStudySettings(ctx, created.ID); err != nil {
		return nil, exceptions.ErrInternalServerError("Failed to create study settings", err)
	}
	return created, nil
}

func (us *UserService) GetUserByID(ctx context.Context, id string) (*entity.UserEntity, error) {
	user, err := us.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.ErrInternalServerError("Failed to load user", err)
	}
	if user == nil {
		return nil, exceptions.ErrNotFound("User not found", nil)
	}
	return user, nil
}

// IsValidTimezone accepts IANA names such as "Asia/Ho_Chi_Minh" (requires tzdata, embedded in main).
func IsValidTimezone(tz string) bool {
	if tz == "" || tz == "Local" {
		return false
	}
	_, err := time.LoadLocation(tz)
	return err == nil
}
