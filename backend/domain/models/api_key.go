package models

import "time"

// API key scopes. A key's scope narrows what it may do regardless of its
// owner's role — role permissions are still enforced separately, so scope
// can only ever reduce access below (never raise it above) the owner's.
const (
	APIKeyScopeRead  = "read"  // safe methods (GET/HEAD/OPTIONS)
	APIKeyScopeWrite = "write" // mutating methods outside /api/admin
	APIKeyScopeAdmin = "admin" // /api/admin/* (and everything write allows)
)

// APIKey is a long-lived credential minted by an account for programmatic
// access. The presented string is "sfs_<id>_<secret>"; only the id, a
// PBKDF2 hash of the secret, and display metadata are stored.
type APIKey struct {
	// ID is a 16-hex-character identifier embedded in the presented key.
	ID string `json:"id"`
	// Name is the owner-chosen label shown in the list.
	Name string `json:"name"`
	// Owner is the username the key authenticates as.
	Owner string `json:"owner"`
	// SecretHash is the PasswordHasher encoding of the secret half.
	SecretHash string `json:"secretHash"`
	// Scope is one of APIKeyScopeRead/Write/Admin.
	Scope string `json:"scope"`
	// CreatedAt is when the key was minted.
	CreatedAt time.Time `json:"createdAt"`
	// ExpiresAt is when the key stops working (zero = never).
	ExpiresAt time.Time `json:"expiresAt"`
	// LastUsedAt is the last successful authentication (zero = never).
	LastUsedAt time.Time `json:"lastUsedAt"`
}

// APIKeyScopeRank orders scopes so code can compare "at least this powerful".
func APIKeyScopeRank(scope string) int {
	switch scope {
	case APIKeyScopeRead:
		return 1
	case APIKeyScopeWrite:
		return 2
	case APIKeyScopeAdmin:
		return 3
	default:
		return 0
	}
}

// APIKeyScopeAllows reports whether scope satisfies the required rank
// (read <= write <= admin).
func APIKeyScopeAllows(scope, required string) bool {
	return APIKeyScopeRank(scope) >= APIKeyScopeRank(required) && APIKeyScopeRank(scope) > 0
}
