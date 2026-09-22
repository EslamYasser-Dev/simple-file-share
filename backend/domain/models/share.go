package models

import "time"

// Share is a public, unguessable link to a file or directory. The token is the
// public identifier embedded in the share URL; it is never derived from the
// path, so links can only be forged with access to the token.
type Share struct {
	Token     string
	Path      string
	Owner     string
	CreatedAt time.Time
	// IsAdmin records whether the owner created this link from the un-scoped
	// system view. It lets the serve path re-scope the stored virtual path with
	// the same privileges the link was created under.
	IsAdmin bool
	// ExpiresAt is the instant after which the link stops working. The zero
	// value means the link never expires.
	ExpiresAt time.Time
}

// Expired reports whether the share has passed its validity window.
func (s *Share) Expired(now time.Time) bool {
	return !s.ExpiresAt.IsZero() && now.After(s.ExpiresAt)
}
