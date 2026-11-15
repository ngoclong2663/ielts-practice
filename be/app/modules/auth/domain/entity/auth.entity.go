package entity

// Entities (have an identity): RefreshTokenEntity.
// Value objects (defined only by their values): GoogleIdentity, ClientMeta.

import "time"

type RefreshTokenEntity struct {
	ID         string
	UserID     string
	TokenHash  string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	ReplacedBy *string
	UserAgent  string
	IP         string
	CreatedAt  time.Time
}

// GoogleIdentity is the verified content of a Google ID token.
type GoogleIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
	Nonce         string
}

// ClientMeta describes the device that requested a session (stored with the refresh token).
type ClientMeta struct {
	UserAgent string
	IP        string
}
