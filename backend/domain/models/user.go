package models

import "time"

// User is an authenticated account. PasswordHash holds the encoded password
// hash and is never serialized to API responses. QuotaBytes is the account's
// storage allowance in bytes; 0 means unlimited. OAuthProvider/OAuthSubject
// bind the account to a stable IdP user id so display-name collisions cannot
// take over another person's account.
type User struct {
	Username      string
	PasswordHash  string
	IsAdmin       bool
	QuotaBytes    int64
	CreatedAt     time.Time
	OAuthProvider string
	OAuthSubject  string
}

// IsSystemView reports whether the user has unrestricted (admin/system)
// access. A nil user is the system view used when auth is disabled.
func (u *User) IsSystemView() bool {
	return u == nil || u.IsAdmin
}

// UserStats is a read model that combines an account with its storage usage,
// used by the admin console.
type UserStats struct {
	Username   string
	IsAdmin    bool
	QuotaBytes int64
	CreatedAt  time.Time
	Files      int
	Size       int64
}
