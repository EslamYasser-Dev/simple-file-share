package ports

import "time"

// JWTClaims represents the claims stored in a JWT token
type JWTClaims struct {
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
}

type JWTProvider interface {
	// GenerateToken creates a new JWT token for the given username
	GenerateToken(username string) (string, error)

	// ValidateToken validates a JWT token and returns the claims
	ValidateToken(tokenString string) (*JWTClaims, error)
}
