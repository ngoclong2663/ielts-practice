package in

import "context"

// What one admin seed run did.
const (
	SeedAdminSkipped   = "skipped"   // ADMIN_USERNAME or ADMIN_PASSWORD is empty
	SeedAdminCreated   = "created"   // the account did not exist
	SeedAdminUpdated   = "updated"   // the password changed: new hash
	SeedAdminUnchanged = "unchanged" // the password still matches
)

// IAdminSeeder upserts the admin account from ADMIN_USERNAME / ADMIN_PASSWORD. Used by `api seed`.
type IAdminSeeder interface {
	SeedAdmin(ctx context.Context, username, password string) (string, error)
}
