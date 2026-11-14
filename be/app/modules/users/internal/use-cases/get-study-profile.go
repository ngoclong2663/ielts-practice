package usecases

import (
	"context"

	"ielts-practice/app/modules/users/domain/entity"
	"ielts-practice/pkg/exceptions"
)

func (u *UserUseCases) GetStudyProfile(ctx context.Context, userID string) (*entity.UserEntity, *entity.StudySettingsEntity, error) {
	user, err := u.GetUserByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	settings, err := u.userRepo.GetStudySettings(ctx, userID)
	if err != nil {
		return nil, nil, exceptions.ErrInternalServerError("Failed to load study settings", err)
	}
	if settings == nil {
		if err := u.userRepo.CreateDefaultStudySettings(ctx, userID); err != nil {
			return nil, nil, exceptions.ErrInternalServerError("Failed to create study settings", err)
		}
		if settings, err = u.userRepo.GetStudySettings(ctx, userID); err != nil || settings == nil {
			return nil, nil, exceptions.ErrInternalServerError("Failed to load study settings", err)
		}
	}
	return user, settings, nil
}
