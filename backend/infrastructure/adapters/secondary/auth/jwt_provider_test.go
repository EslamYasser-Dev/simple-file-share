package auth

import (
	"testing"
	"time"
)

func TestJWTProvider(t *testing.T) {
	secretKey := "test-secret-key"
	expiry := 1 * time.Hour

	provider := NewJWTProvider(secretKey, expiry)

	// Test token generation
	username := "testuser"
	token, err := provider.GenerateToken(username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Test token validation
	claims, err := provider.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.Username != username {
		t.Errorf("Expected username %s, got %s", username, claims.Username)
	}

	// Test expired token
	shortExpiry := 1 * time.Millisecond
	shortProvider := NewJWTProvider(secretKey, shortExpiry)
	expiredToken, err := shortProvider.GenerateToken(username)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err = shortProvider.ValidateToken(expiredToken)
	if err == nil {
		t.Fatal("Expected error for expired token")
	}

	// Test invalid token
	_, err = provider.ValidateToken("invalid.token.here")
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}
}
