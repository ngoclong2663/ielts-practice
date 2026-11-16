package usecases

import (
	"context"
	"net/http"

	"ielts-practice/app/modules/admin/dto"
	"ielts-practice/pkg/exceptions"
)

// GetMe returns the signed-in admin.
func (uc *AdminUseCases) GetMe(ctx context.Context, adminID string) (dto.AdminDto, error) {
	account, err := uc.accounts.FindByID(ctx, adminID)
	if err != nil {
		return dto.AdminDto{}, internalErr("Failed to load admin account", err)
	}
	if account == nil {
		return dto.AdminDto{}, exceptions.NewException(http.StatusUnauthorized, exceptions.CodeAdminUnauthorized, "Admin sign-in required", nil)
	}
	return dto.AdminDto{ID: account.ID, Username: account.Username}, nil
}
