package usecases

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	"ielts-practice/app/modules/admin/domain"
	adminin "ielts-practice/app/modules/admin/ports/in"
	"ielts-practice/configs"
)

const (
	bcryptCost             = 12
	minProductionPassword  = 12
	maxUsernameRunes       = 64
	maxPasswordBytesBcrypt = 72
)

var _ adminin.IAdminSeeder = &AdminSeeder{}

// AdminSeeder upserts the admin account (`api seed`). It needs only the database.
type AdminSeeder struct {
	accounts domain.IAdminAccountRepository
	cfg      *configs.Config
	logger   zerolog.Logger
}

func NewAdminSeeder(accounts domain.IAdminAccountRepository, cfg *configs.Config, logger zerolog.Logger) *AdminSeeder {
	return &AdminSeeder{accounts: accounts, cfg: cfg, logger: logger.With().Str("module", "admin").Logger()}
}

// SeedAdmin creates the account, or replaces its hash when the password no longer matches
// (changing ADMIN_PASSWORD rotates it). The password is never logged.
func (s *AdminSeeder) SeedAdmin(ctx context.Context, username, password string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		s.logger.Info().Msg("Admin seed skipped: ADMIN_USERNAME or ADMIN_PASSWORD is empty")
		return adminin.SeedAdminSkipped, nil
	}
	if utf8.RuneCountInString(username) > maxUsernameRunes {
		return "", fmt.Errorf("ADMIN_USERNAME must be at most %d characters", maxUsernameRunes)
	}
	if len(password) > maxPasswordBytesBcrypt {
		return "", fmt.Errorf("ADMIN_PASSWORD must be at most %d bytes", maxPasswordBytesBcrypt)
	}
	if s.cfg.IsProduction() && utf8.RuneCountInString(password) < minProductionPassword {
		return "", fmt.Errorf("ADMIN_PASSWORD must be at least %d characters in production", minProductionPassword)
	}

	account, err := s.accounts.FindByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("load admin account: %w", err)
	}
	if account != nil && bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)) == nil {
		s.logger.Info().Str("username", account.Username).Msg("Admin account unchanged")
		return adminin.SeedAdminUnchanged, nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash admin password: %w", err)
	}
	if account == nil {
		if _, err := s.accounts.Insert(ctx, username, string(hash)); err != nil {
			return "", fmt.Errorf("create admin account: %w", err)
		}
		s.logger.Info().Str("username", username).Msg("Admin account created")
		return adminin.SeedAdminCreated, nil
	}
	if err := s.accounts.UpdatePasswordHash(ctx, account.ID, string(hash)); err != nil {
		return "", fmt.Errorf("update admin password: %w", err)
	}
	s.logger.Info().Str("username", account.Username).Msg("Admin password updated")
	return adminin.SeedAdminUpdated, nil
}
