package domain

import (
	"context"
	"time"

	"ielts-practice/app/modules/admin/domain/entity"
)

// IAdminRepository: read-only reporting queries over every module's tables (Epic 6 §6).
type IAdminRepository interface {
	Totals(ctx context.Context, w entity.Window) (entity.Totals, error)
	// Daily returns the chart values by date (YYYY-MM-DD); days without data are missing.
	Daily(ctx context.Context, w entity.Window) (map[string]entity.DayStats, error)
	RecentReports(ctx context.Context, limit int) ([]entity.ContentReport, error)
	ListUsers(ctx context.Context, f entity.UserFilter) ([]entity.UserActivity, int64, error)
}

// IAdminAccountRepository: admin sign-in accounts. Usernames match case-insensitively.
// Find methods return nil when there is no such account.
type IAdminAccountRepository interface {
	FindByUsername(ctx context.Context, username string) (*entity.AdminAccount, error)
	FindByID(ctx context.Context, id string) (*entity.AdminAccount, error)
	Insert(ctx context.Context, username, passwordHash string) (*entity.AdminAccount, error)
	UpdatePasswordHash(ctx context.Context, id, passwordHash string) error
	TouchLogin(ctx context.Context, id string, at time.Time) error
}
