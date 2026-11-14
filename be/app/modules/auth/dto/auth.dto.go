package dto

import "time"

type GoogleSignInRequestDto struct {
	Code         string `json:"code" binding:"required,max=2048"`
	CodeVerifier string `json:"code_verifier" binding:"required,min=43,max=128"`
	Nonce        string `json:"nonce" binding:"required,max=256"`
	UILocale     string `json:"ui_locale" binding:"omitempty,oneof=vi en"`
	Timezone     string `json:"timezone" binding:"omitempty,max=64"`
}

type RefreshRequestDto struct {
	RefreshToken string `json:"refresh_token" binding:"required,max=256"`
}

type LogoutRequestDto struct {
	RefreshToken string `json:"refresh_token" binding:"omitempty,max=256"`
}

// TestLoginRequestDto is only accepted when APP_ENV=test (end-to-end tests without Google).
type TestLoginRequestDto struct {
	Email       string `json:"email" binding:"required,email"`
	DisplayName string `json:"display_name" binding:"required,max=120"`
	UILocale    string `json:"ui_locale" binding:"omitempty,oneof=vi en"`
}

type SessionUserDto struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	UILocale    string `json:"ui_locale"`
}

type SessionResponseDto struct {
	AccessToken           string         `json:"access_token"`
	AccessTokenExpiresAt  time.Time      `json:"access_token_expires_at"`
	RefreshToken          string         `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time      `json:"refresh_token_expires_at"`
	User                  SessionUserDto `json:"user"`
}
