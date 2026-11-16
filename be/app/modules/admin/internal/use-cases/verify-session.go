package usecases

import "context"

// VerifyAdminToken backs AdminAuthMiddleware: a valid admin token whose account still exists
// (deleting the row revokes access at once).
func (uc *AdminUseCases) VerifyAdminToken(ctx context.Context, raw string) (string, bool, error) {
	id, err := uc.jwt.Verify(raw)
	if err != nil {
		return "", false, nil
	}
	account, err := uc.accounts.FindByID(ctx, id)
	if err != nil {
		return "", false, err
	}
	if account == nil {
		return "", false, nil
	}
	return account.ID, true, nil
}
