package usecases

import (
	"context"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ielts-practice/app/modules/admin/dto"
	"ielts-practice/pkg/exceptions"
)

// dummyHash (cost 12, like real hashes) is compared when the username is unknown, so a wrong
// username takes as long as a wrong password.
const dummyHash = "$2a$12$2Pd2Cst1xelJRxCKzyIJD.H40GaJ8EPVotFHfOjoNjAYSqSZdJjHq"

// Login checks the admin username (case-insensitive) and password and issues an admin token.
// The Next.js server stores the token in its own cookie.
func (uc *AdminUseCases) Login(ctx context.Context, req dto.LoginRequestDto) (dto.LoginResponseDto, error) {
	account, err := uc.accounts.FindByUsername(ctx, strings.TrimSpace(req.Username))
	if err != nil {
		return dto.LoginResponseDto{}, internalErr("Failed to load admin account", err)
	}
	hash := dummyHash
	if account != nil {
		hash = account.PasswordHash
	}
	// Always exactly one bcrypt compare.
	match := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) == nil
	if account == nil || !match {
		return dto.LoginResponseDto{}, exceptions.NewException(http.StatusUnauthorized, exceptions.CodeAdminInvalidCreds, "Wrong username or password", nil)
	}

	token, expiresAt, err := uc.jwt.Issue(account.ID)
	if err != nil {
		return dto.LoginResponseDto{}, internalErr("Failed to issue admin token", err)
	}
	if err := uc.accounts.TouchLogin(ctx, account.ID, uc.now()); err != nil {
		return dto.LoginResponseDto{}, internalErr("Failed to save admin sign-in", err)
	}
	return dto.LoginResponseDto{
		AccessToken: token,
		ExpiresAt:   expiresAt.UTC().Format(time.RFC3339),
		Admin:       dto.AdminDto{ID: account.ID, Username: account.Username},
	}, nil
}
