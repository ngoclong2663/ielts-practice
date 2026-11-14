package infrastructure

import (
	"context"
	"errors"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/users/domain"
	"ielts-practice/app/modules/users/domain/entity"
	"ielts-practice/app/modules/users/internal/mapper"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var _ domain.IUserRepository = &UserRepository{}

type UserRepository struct {
	db         *gorm.DB
	userMapper *mapper.UserMapper
	logger     zerolog.Logger
}

func NewUserRepository(db *gorm.DB, userMapper *mapper.UserMapper, logger zerolog.Logger) domain.IUserRepository {
	return &UserRepository{db: db, userMapper: userMapper, logger: logger}
}

func (r *UserRepository) findOne(ctx context.Context, query string, arg any) (*entity.UserEntity, error) {
	var user models.User
	err := database.GetDBFromContext(ctx, r.db).WithContext(ctx).
		Where("deleted_at IS NULL").
		First(&user, query, arg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.userMapper.MapUserModelToEntity(&user), nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.UserEntity, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *UserRepository) FindByGoogleSub(ctx context.Context, googleSub string) (*entity.UserEntity, error) {
	return r.findOne(ctx, "google_sub = ?", googleSub)
}

func (r *UserRepository) Create(ctx context.Context, user *entity.UserEntity) (*entity.UserEntity, error) {
	model := r.userMapper.MapUserEntityToModel(user)
	if err := database.GetDBFromContext(ctx, r.db).WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}
	return r.userMapper.MapUserModelToEntity(model), nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.UserEntity) error {
	model := r.userMapper.MapUserEntityToModel(user)
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx).
		Model(&models.User{ID: user.ID}).
		Select("email", "display_name", "avatar_url", "ui_locale", "timezone", "day_start_hour", "last_login_at", "updated_at").
		Updates(model).Error
}

func (r *UserRepository) GetStudySettings(ctx context.Context, userID string) (*entity.StudySettingsEntity, error) {
	var settings models.UserStudySettings
	err := database.GetDBFromContext(ctx, r.db).WithContext(ctx).First(&settings, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return r.userMapper.MapSettingsModelToEntity(&settings), nil
}

func (r *UserRepository) CreateDefaultStudySettings(ctx context.Context, userID string) error {
	// Column defaults in the migration are the single source of truth for default values.
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx).
		Exec("INSERT INTO user_study_settings (user_id) VALUES (?) ON CONFLICT (user_id) DO NOTHING", userID).Error
}

func (r *UserRepository) UpdateStudySettings(ctx context.Context, settings *entity.StudySettingsEntity) error {
	model := r.userMapper.MapSettingsEntityToModel(settings)
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx).
		Model(&models.UserStudySettings{UserID: settings.UserID}).
		Select("new_cards_per_day", "max_reviews_per_day", "desired_retention", "audio_autoplay", "accent", "show_image_on_front",
			"learning_steps", "relearning_steps", "max_interval_days", "updated_at").
		Updates(model).Error
}
