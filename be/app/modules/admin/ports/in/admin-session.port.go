// Package in is what the admin module offers: admin session checks for other modules' admin
// routes, and the admin account seed (used by `api seed`).
package in

import "context"

// IAdminSession verifies admin tokens (implements middlewares.AdminVerifier).
type IAdminSession interface {
	// VerifyAdminToken: ok is false for an invalid or expired token or a deleted account.
	VerifyAdminToken(ctx context.Context, raw string) (adminID string, ok bool, err error)
}
