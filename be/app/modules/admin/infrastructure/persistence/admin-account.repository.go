package persistence

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/admin/domain"
	"ielts-practice/app/modules/admin/domain/entity"
	"ielts-practice/app/modules/database"
)

var _ domain.IAdminAccountRepository = &AdminAccountRepository{}

type AdminAccountRepository struct {
	db *gorm.DB
}

func NewAdminAccountRepository(db *gorm.DB) domain.IAdminAccountRepository {
	return &AdminAccountRepository{db: db}
}

func (r *AdminAccountRepository) conn(ctx context.Context) *gorm.DB {
	return database.GetDBFromContext(ctx, r.db).WithContext(ctx)
}

func toAccount(m *models.AdminAccount) *entity.AdminAccount {
	return &entity.AdminAccount{ID: m.ID, Username: m.Username, PasswordHash: m.PasswordHash, LastLoginAt: m.LastLoginAt, CreatedAt: m.CreatedAt}
}

func (r *AdminAccountRepository) find(ctx context.Context, where string, arg any) (*entity.AdminAccount, error) {
	var rows []models.AdminAccount
	if err := r.conn(ctx).Where(where, arg).Limit(1).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return toAccount(&rows[0]), nil
}

func (r *AdminAccountRepository) FindByUsername(ctx context.Context, username string) (*entity.AdminAccount, error) {
	return r.find(ctx, "lower(username) = lower(?)", username)
}

func (r *AdminAccountRepository) FindByID(ctx context.Context, id string) (*entity.AdminAccount, error) {
	return r.find(ctx, "id = ?", id)
}

func (r *AdminAccountRepository) Insert(ctx context.Context, username, passwordHash string) (*entity.AdminAccount, error) {
	m := models.AdminAccount{Username: username, PasswordHash: passwordHash}
	if err := r.conn(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	return toAccount(&m), nil
}

func (r *AdminAccountRepository) UpdatePasswordHash(ctx context.Context, id, passwordHash string) error {
	return r.conn(ctx).Model(&models.AdminAccount{}).Where("id = ?", id).
		Updates(map[string]any{"password_hash": passwordHash, "updated_at": time.Now()}).Error
}

func (r *AdminAccountRepository) TouchLogin(ctx context.Context, id string, at time.Time) error {
	return r.conn(ctx).Model(&models.AdminAccount{}).Where("id = ?", id).UpdateColumn("last_login_at", at).Error
}
