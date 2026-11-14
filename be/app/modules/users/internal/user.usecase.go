package internal

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/users/domain"
	"ielts-practice/app/modules/users/domain/entity"
	"ielts-practice/app/modules/users/dto"
	"ielts-practice/app/modules/users/ports"
	"ielts-practice/pkg/exceptions"
)

var _ IUserUseCase = &UserUseCase{}

type IUserUseCase interface {
	GetMe(ctx context.Context, userID string) (dto.MeResponseDto, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequestDto) (dto.MeResponseDto, error)
	UpdateStudySettings(ctx context.Context, userID string, req dto.UpdateStudySettingsRequestDto) (dto.StudySettingsDto, error)
}

type UserUseCase struct {
	userRepo    domain.IUserRepository
	userService ports.IUserService
	transaction *database.TransactionStruct
	mapper      *UserMapper
	logger      zerolog.Logger
}

func NewUserUseCase(
	userRepo domain.IUserRepository,
	userService ports.IUserService,
	transaction *database.TransactionStruct,
	mapper *UserMapper,
	logger zerolog.Logger,
) IUserUseCase {
	return &UserUseCase{
		userRepo:    userRepo,
		userService: userService,
		transaction: transaction,
		mapper:      mapper,
		logger:      logger,
	}
}

func (u *UserUseCase) GetMe(ctx context.Context, userID string) (dto.MeResponseDto, error) {
	user, err := u.userService.GetUserByID(ctx, userID)
	if err != nil {
		return dto.MeResponseDto{}, err
	}
	settings, err := u.loadSettings(ctx, userID)
	if err != nil {
		return dto.MeResponseDto{}, err
	}
	return dto.MeResponseDto{
		User:          u.mapper.MapUserEntityToProfileDto(user),
		StudySettings: u.mapper.MapSettingsEntityToDto(settings),
	}, nil
}

func (u *UserUseCase) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequestDto) (dto.MeResponseDto, error) {
	if req.Timezone != nil && !IsValidTimezone(*req.Timezone) {
		appErr := exceptions.NewException(http.StatusBadRequest, exceptions.CodeUserInvalidTimezone, "Unknown timezone", nil)
		appErr.Fields = map[string]string{"timezone": exceptions.CodeUserInvalidTimezone}
		return dto.MeResponseDto{}, appErr
	}

	user, err := u.userService.GetUserByID(ctx, userID)
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

func (u *UserUseCase) UpdateStudySettings(ctx context.Context, userID string, req dto.UpdateStudySettingsRequestDto) (dto.StudySettingsDto, error) {
	var result *entity.StudySettingsEntity
	err := u.transaction.Execute(ctx, func(txCtx context.Context) error {
		settings, err := u.loadSettings(txCtx, userID)
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

// loadSettings returns the user's study settings, creating the defaults row if it is missing.
func (u *UserUseCase) loadSettings(ctx context.Context, userID string) (*entity.StudySettingsEntity, error) {
	settings, err := u.userRepo.GetStudySettings(ctx, userID)
	if err != nil {
		return nil, exceptions.ErrInternalServerError("Failed to load study settings", err)
	}
	if settings != nil {
		return settings, nil
	}
	if err := u.userRepo.CreateDefaultStudySettings(ctx, userID); err != nil {
		return nil, exceptions.ErrInternalServerError("Failed to create study settings", err)
	}
	settings, err = u.userRepo.GetStudySettings(ctx, userID)
	if err != nil || settings == nil {
		return nil, exceptions.ErrInternalServerError("Failed to load study settings", err)
	}
	return settings, nil
}
