package usecases

import (
	"context"

	"ielts-practice/app/modules/users/dto"
)

func (u *UserUseCases) GetMe(ctx context.Context, userID string) (dto.MeResponseDto, error) {
	user, settings, err := u.GetStudyProfile(ctx, userID)
	if err != nil {
		return dto.MeResponseDto{}, err
	}
	return dto.MeResponseDto{
		User:          u.mapper.MapUserEntityToProfileDto(user),
		StudySettings: u.mapper.MapSettingsEntityToDto(settings),
	}, nil
}
