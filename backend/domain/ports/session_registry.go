package ports

import "time"

// Session is one issued access token tracked for the session list.
type Session struct {
	JTI       string    `json:"jti"`
	Subject   string    `json:"subject"`
	IssuedAt  time.Time `json:"issuedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	Remote    string    `json:"remote,omitempty"`
	UserAgent string    `json:"userAgent,omitempty"`
}

// SessionRegistry tracks issued access tokens so users can see and revoke
// their active sessions. It is a convenience view over the stateless JWT —
// revocation still goes through TokenManager.Revoke.
type SessionRegistry interface {
	// Record upserts a session by JTI and prunes expired ones. Implementations
	// bound the number of retained sessions per subject.
	Record(sess Session) error
	// List returns a subject's sessions newest first; empty subject returns
	// every tracked session (system view).
	List(subject string) ([]Session, error)
	// Find returns the session with the given JTI (nil, false when unknown).
	Find(jti string) (*Session, bool, error)
	// Remove drops a session from the list (idempotent).
	Remove(jti string) error
}
