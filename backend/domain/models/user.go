package models

import "time"

// User is an authenticated account. PasswordHash holds the encoded password
// hash and is never serialized to API responses.
type User struct {
	Username     string
	PasswordHash string
	IsAdmin      bool
	CreatedAt    time.Time
}

// UserStats is a read model that combines an account with its storage usage,
// used by the admin console.
type UserStats struct {
	Username  string
	IsAdmin   bool
	CreatedAt time.Time
	Files     int
	Size      int64
}
