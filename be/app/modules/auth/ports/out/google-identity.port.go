// Package out lists what the auth module needs from outside: Google's OAuth identity check.
package out

import (
	"context"
	"errors"

	"ielts-practice/app/modules/auth/domain/entity"
)

// Errors returned by IGoogleIdentityProvider.
var (
	ErrGoogleNotConfigured = errors.New("google oauth is not configured")
	ErrGoogleExchange      = errors.New("google code exchange failed")
	ErrGoogleIDToken       = errors.New("google id token invalid")
)

// IGoogleIdentityProvider exchanges an authorization code (PKCE) and returns the verified identity.
type IGoogleIdentityProvider interface {
	Exchange(ctx context.Context, code string, codeVerifier string) (*entity.GoogleIdentity, error)
}
