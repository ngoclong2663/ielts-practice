package domain

import (
	"context"

	"ielts-practice/app/modules/users/domain/entity"
)

// IUserRepository: Find* methods return (nil, nil) when no row matches.
type IUserRepository interface {
	FindByID(ctx context.Context, id string) (*entity.UserEntity, error)
	FindByGoogleSub(ctx context.Context, googleSub string) (*entity.UserEntity, error)
	Create(ctx context.Context, user *entity.UserEntity) (*entity.UserEntity, error)
	Update(ctx context.Context, user *entity.UserEntity) error

	GetStudySettings(ctx context.Context, userID string) (*entity.StudySettingsEntity, error)
	CreateDefaultStudySettings(ctx context.Context, userID string) error
	UpdateStudySettings(ctx context.Context, settings *entity.StudySettingsEntity) error
}
