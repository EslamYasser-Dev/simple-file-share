package ports

import "time"

// TokenClaims are the verified fields of an access token.
type TokenClaims struct {
	Subject   string
	ID        string
	ExpiresAt time.Time
}

// TokenManager issues and verifies stateless access tokens. Verify rejects
// expired tokens and any token whose ID appears in the revocation list.
// Revoke must persist when the implementation has a store, so restarts do not
// resurrect revoked tokens.
type TokenManager interface {
	Issue(subject string) (string, TokenClaims, error)
	Verify(token string) (*TokenClaims, error)
	Revoke(id string, expiresAt time.Time)
}
