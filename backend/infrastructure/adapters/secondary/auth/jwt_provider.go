package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// HMACJWTProvider implements ports.JWTProvider using HMAC-SHA256 (HS256).
type HMACJWTProvider struct {
	secretKey []byte
	expiry    time.Duration
}

func NewJWTProvider(secretKey string, expiry time.Duration) *HMACJWTProvider {
	return &HMACJWTProvider{
		secretKey: []byte(secretKey),
		expiry:    expiry,
	}
}

func (p *HMACJWTProvider) GenerateToken(username string) (string, error) {
	now := time.Now()
	claims := map[string]any{
		"username": username,
		"exp":      now.Add(p.expiry).Unix(),
		"iat":      now.Unix(),
	}

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Encode header
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode claims
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerEncoded + "." + claimsEncoded
	h := hmac.New(sha256.New, p.secretKey)
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	// Combine all parts
	token := message + "." + signature
	return token, nil
}

// ValidateToken validates a JWT token and returns the claims.
func (p *HMACJWTProvider) ValidateToken(tokenString string) (*ports.JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerEncoded, claimsEncoded, signatureEncoded := parts[0], parts[1], parts[2]

	// Verify the header declares HS256 before trusting the signature.
	headerJSON, err := base64.RawURLEncoding.DecodeString(headerEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}
	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil || header.Alg != "HS256" {
		return nil, errors.New("unsupported token algorithm")
	}

	// Verify signature in constant time to prevent timing attacks.
	message := headerEncoded + "." + claimsEncoded
	mac := hmac.New(sha256.New, p.secretKey)
	mac.Write([]byte(message))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signatureEncoded), []byte(expectedSignature)) {
		return nil, errors.New("invalid signature")
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(claimsEncoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Extract username
	username, ok := claims["username"].(string)
	if !ok {
		return nil, errors.New("invalid username claim")
	}

	// Extract expiration time
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("invalid exp claim")
	}
	expiresAt := time.Unix(int64(expFloat), 0)

	// Check if token is expired
	if time.Now().After(expiresAt) {
		return nil, errors.New("token expired")
	}

	// Extract issued at time
	iatFloat, ok := claims["iat"].(float64)
	if !ok {
		return nil, errors.New("invalid iat claim")
	}
	issuedAt := time.Unix(int64(iatFloat), 0)

	return &ports.JWTClaims{
		Username:  username,
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  issuedAt.Unix(),
	}, nil
}

var _ ports.JWTProvider = (*HMACJWTProvider)(nil)
