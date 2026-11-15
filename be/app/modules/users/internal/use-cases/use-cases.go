package usecases

import (
	"context"
	"time"

	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/users/domain"
	"ielts-practice/app/modules/users/dto"
	"ielts-practice/app/modules/users/internal/mapper"
	usersin "ielts-practice/app/modules/users/ports/in"

	"github.com/rs/zerolog"
)

// UserUseCases holds all user business logic, one use case per file in this folder. It serves this module's HTTP endpoints
// (IUserUseCases) and is also the implementation of usersin.IUserService used by other
// modules, so every rule lives in one place.
var (
	_ IUserUseCases        = &UserUseCases{}
	_ usersin.IUserService = &UserUseCases{}
)

type IUserUseCases interface {
	GetMe(ctx context.Context, userID string) (dto.MeResponseDto, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequestDto) (dto.MeResponseDto, error)
	UpdateStudySettings(ctx context.Context, userID string, req dto.UpdateStudySettingsRequestDto) (dto.StudySettingsDto, error)
}

type UserUseCases struct {
	userRepo    domain.IUserRepository
	transaction *database.TransactionStruct
	mapper      *mapper.UserMapper
	logger      zerolog.Logger
	now         func() time.Time
}

func NewUserUseCases(
	userRepo domain.IUserRepository,
	transaction *database.TransactionStruct,
	mapper *mapper.UserMapper,
	logger zerolog.Logger,
) *UserUseCases {
	return &UserUseCases{
		userRepo:    userRepo,
		transaction: transaction,
		mapper:      mapper,
		logger:      logger,
		now:         time.Now,
	}
}

const (
	defaultLocale   = "en"
	defaultTimezone = "Asia/Ho_Chi_Minh"
)

// IsValidTimezone accepts IANA names such as "Asia/Ho_Chi_Minh" (requires tzdata, embedded in main).
func IsValidTimezone(tz string) bool {
	if tz == "" || tz == "Local" {
		return false
	}
	_, err := time.LoadLocation(tz)
	return err == nil
}
