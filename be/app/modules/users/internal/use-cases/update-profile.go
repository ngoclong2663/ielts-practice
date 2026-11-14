package usecases

import (
	"context"
	"net/http"

	"ielts-practice/app/modules/users/dto"
	"ielts-practice/pkg/exceptions"
)

func (u *UserUseCases) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequestDto) (dto.MeResponseDto, error) {
	if req.Timezone != nil && !IsValidTimezone(*req.Timezone) {
		appErr := exceptions.NewException(http.StatusBadRequest, exceptions.CodeUserInvalidTimezone, "Unknown timezone", nil)
		appErr.Fields = map[string]string{"timezone": exceptions.CodeUserInvalidTimezone}
		return dto.MeResponseDto{}, appErr
	}

	user, err := u.GetUserByID(ctx, userID)
	if err != nil {
		return dto.MeResponseDto{}, err
	}
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.UILocale != nil {
		user.UILocale = *req.UILocale
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}
	if req.DayStartHour != nil {
		user.DayStartHour = *req.DayStartHour
	}
	if err := u.userRepo.Update(ctx, user); err != nil {
		return dto.MeResponseDto{}, exceptions.ErrInternalServerError("Failed to update profile", err)
	}
	return u.GetMe(ctx, userID)
}
