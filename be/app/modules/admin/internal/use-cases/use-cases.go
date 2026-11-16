package usecases

import (
	"context"
	"fmt"
	"time"

	"ielts-practice/app/modules/admin/domain"
	"ielts-practice/app/modules/admin/domain/entity"
	"ielts-practice/app/modules/admin/domain/rules"
	"ielts-practice/app/modules/admin/dto"
	adminin "ielts-practice/app/modules/admin/ports/in"
	"ielts-practice/configs"
	"ielts-practice/pkg/exceptions"
	"ielts-practice/pkg/tokens"
)

const (
	defaultPageSize = 25
	recentReports   = 5
)

// IAdminUseCases is what the HTTP presenter needs. One use case per file in this folder.
type IAdminUseCases interface {
	Login(ctx context.Context, req dto.LoginRequestDto) (dto.LoginResponseDto, error)
	GetMe(ctx context.Context, adminID string) (dto.AdminDto, error)
	GetOverview(ctx context.Context) (dto.OverviewDto, error)
	ListUsers(ctx context.Context, q dto.ListUsersQueryDto) (dto.AdminUserListDto, error)
}

var (
	_ IAdminUseCases        = &AdminUseCases{}
	_ adminin.IAdminSession = &AdminUseCases{}
)

type AdminUseCases struct {
	repo     domain.IAdminRepository
	accounts domain.IAdminAccountRepository
	jwt      *tokens.AdminJWT
	loc      *time.Location
	now      func() time.Time
}

func NewAdminUseCases(repo domain.IAdminRepository, accounts domain.IAdminAccountRepository, jwt *tokens.AdminJWT, cfg *configs.Config) (*AdminUseCases, error) {
	loc, err := time.LoadLocation(cfg.Admin.Timezone)
	if err != nil {
		return nil, fmt.Errorf("ADMIN_TIMEZONE: %w", err)
	}
	return &AdminUseCases{repo: repo, accounts: accounts, jwt: jwt, loc: loc, now: time.Now}, nil
}

func (uc *AdminUseCases) window() entity.Window {
	return rules.NewWindow(uc.now(), uc.loc)
}

func internalErr(msg string, err error) *exceptions.AppError {
	return exceptions.ErrInternalServerError(msg, err)
}
