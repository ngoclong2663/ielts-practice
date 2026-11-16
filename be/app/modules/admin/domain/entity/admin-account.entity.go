package entity

// Entities: AdminAccount (the admin board sign-in, seeded by `api seed`).

import "time"

type AdminAccount struct {
	ID           string
	Username     string
	PasswordHash string // bcrypt
	LastLoginAt  *time.Time
	CreatedAt    time.Time
}
