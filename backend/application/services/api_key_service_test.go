package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

func apiKeyFixture(t *testing.T) (*APIKeyService, *fs.UserFileRepository, *models.User) {
	t.Helper()
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	svc := NewAPIKeyService(fs.NewAPIKeyFileRepository(dir), userRepo, hasher)
	alice := &models.User{Username: "alice", Role: models.RoleMember, Enabled: true, CreatedAt: time.Now().UTC()}
	if err := userRepo.CreateUser(alice); err != nil {
		t.Fatal(err)
	}
	return svc, userRepo, alice
}

func TestAPIKeyIssueAuthenticateRoundtrip(t *testing.T) {
	svc, _, alice := apiKeyFixture(t)

	plaintext, key, err := svc.Issue(alice, "ci bot", "write", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(plaintext, APIKeyPrefix+key.ID+"_") {
		t.Fatalf("plaintext %q does not embed id %q", plaintext, key.ID)
	}
	parts := strings.Split(plaintext, "_")
	if len(parts) != 3 || len(parts[1]) != 16 || len(parts[2]) != 32 {
		t.Fatalf("key shape = %q (id %d chars, secret %d chars)", plaintext, len(parts[1]), len(parts[2]))
	}
	if key.SecretHash == "" || strings.Contains(key.SecretHash, parts[2]) {
		t.Fatal("secret must be stored only as a non-leaky hash")
	}

	user, scope, err := svc.Authenticate(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if user.Username != "alice" || scope != models.APIKeyScopeWrite {
		t.Fatalf("resolved user=%s scope=%s", user.Username, scope)
	}
	// LastUsedAt gets recorded on first use.
	stored, _ := svc.keys.Find(key.ID)
	if stored.LastUsedAt.IsZero() {
		t.Fatal("first use must record LastUsedAt")
	}
}

func TestAPIKeyAuthenticateRejects(t *testing.T) {
	svc, userRepo, alice := apiKeyFixture(t)
	plaintext, key, err := svc.Issue(alice, "ci", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	id := key.ID
	secret := strings.Split(plaintext, "_")[2]

	cases := map[string]string{
		"garbage":         "not-a-key",
		"empty":           "",
		"jwt-shaped":      "eyJhbGciOi.secret",
		"wrong-prefix":    "xyz_" + id + "_" + secret,
		"short-id":        APIKeyPrefix + "abc_" + secret,
		"non-hex-id":      APIKeyPrefix + "zzzzzzzzzzzzzzzz_" + secret,
		"missing-secret":  APIKeyPrefix + id + "_",
		"unknown-id":      APIKeyPrefix + "ffffffffffffffff_" + secret,
		"wrong-secret":    APIKeyPrefix + id + "_" + strings.Repeat("a", 32),
		"trailing-garish": plaintext + "x",
	}
	for name, presented := range cases {
		if _, _, err := svc.Authenticate(presented); !errors.Is(err, ports.ErrUnauthorized) {
			t.Errorf("%s: err = %v, want ErrUnauthorized", name, err)
		}
	}
	if _, _, err := svc.Authenticate(plaintext); err != nil {
		t.Fatalf("valid key rejected after negatives: %v", err)
	}

	// Deleted owner kills the key.
	if err := userRepo.DeleteUser("alice"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Authenticate(plaintext); !errors.Is(err, ports.ErrUnauthorized) {
		t.Fatalf("deleted owner: err = %v, want ErrUnauthorized", err)
	}
}

func TestAPIKeyExpiryAndDisabledOwner(t *testing.T) {
	svc, userRepo, alice := apiKeyFixture(t)

	// Already-expired key.
	_, key, err := svc.Issue(alice, "old", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	expired := key
	expired.ExpiresAt = time.Now().Add(-time.Minute)
	if err := svc.keys.Save(expired); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Authenticate(APIKeyPrefix + key.ID + "_whatever"); !errors.Is(err, ports.ErrUnauthorized) {
		t.Fatalf("expired key: %v", err)
	}

	// Future expiry is fine.
	plaintext, key2, err := svc.Issue(alice, "fresh", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if key2.ExpiresAt.Before(time.Now()) {
		t.Fatalf("expiresAt = %v, want future", key2.ExpiresAt)
	}
	if _, _, err := svc.Authenticate(plaintext); err != nil {
		t.Fatalf("live key: %v", err)
	}

	// Disabled owner.
	alice.Enabled = false
	if err := userRepo.UpdateUser(alice, alice.Username); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Authenticate(plaintext); !errors.Is(err, ports.ErrUnauthorized) {
		t.Fatalf("disabled owner: %v", err)
	}
}

func TestAPIKeyIssueValidation(t *testing.T) {
	svc, _, alice := apiKeyFixture(t)
	if _, _, err := svc.Issue(nil, "x", "", 0); err == nil {
		t.Fatal("nil actor must fail")
	}
	if _, _, err := svc.Issue(alice, "", "", 0); err == nil {
		t.Fatal("empty name must fail")
	}
	if _, _, err := svc.Issue(alice, strings.Repeat("n", 81), "", 0); err == nil {
		t.Fatal("oversized name must fail")
	}
	if _, _, err := svc.Issue(alice, "x", "root", 0); err == nil {
		t.Fatal("unknown scope must fail")
	}
	// Empty scope defaults to read.
	_, key, err := svc.Issue(alice, "x", "", 0)
	if err != nil || key.Scope != models.APIKeyScopeRead {
		t.Fatalf("default scope = %q err=%v", key.Scope, err)
	}
}

func TestAPIKeyListAndRevokeOwnership(t *testing.T) {
	svc, _, alice := apiKeyFixture(t)
	plaintextOne, key, err := svc.Issue(alice, "one", "admin", 0)
	if err != nil {
		t.Fatal(err)
	}
	plaintext, _, err := svc.Issue(alice, "two", "", 0)
	if err != nil {
		t.Fatal(err)
	}

	// List hides hashes.
	keys, err := svc.List(alice)
	if err != nil || len(keys) != 2 {
		t.Fatalf("list: %v %d keys", err, len(keys))
	}
	for _, k := range keys {
		if k.SecretHash == "" {
			t.Fatal("list must return stored records (hash present for service use)")
		}
	}
	if _, err := svc.List(nil); err == nil {
		t.Fatal("nil actor list must fail")
	}

	// Another owner cannot revoke (reported as not found), and cannot probe.
	mallory := &models.User{Username: "mallory", Enabled: true}
	if err := svc.Revoke(mallory, key.ID); err == nil {
		t.Fatal("cross-owner revoke must fail")
	}
	if _, _, err := svc.Authenticate(APIKeyPrefix + key.ID + "_anything"); err == nil {
		t.Fatal("key must survive a foreign revoke attempt")
	}
	if err := svc.Revoke(mallory, "ffffffffffffffff"); err == nil {
		t.Fatal("unknown id revoke must fail")
	}

	// The owner revokes; the key dies immediately (the other key lives).
	if err := svc.Revoke(alice, key.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Authenticate(plaintext); err != nil {
		t.Fatalf("second key must survive: %v", err)
	}
	if _, _, err := svc.Authenticate(plaintextOne); !errors.Is(err, ports.ErrUnauthorized) {
		t.Fatalf("revoked key authenticates: %v", err)
	}
	if keys, _ := svc.List(alice); len(keys) != 1 {
		t.Fatalf("after revoke: %d keys", len(keys))
	}
}

func TestAPIKeyTouchThrottling(t *testing.T) {
	svc, _, alice := apiKeyFixture(t)
	plaintext, key, err := svc.Issue(alice, "hot", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Authenticate(plaintext); err != nil {
		t.Fatal(err)
	}
	first, _ := svc.keys.Find(key.ID)

	// Within the window LastUsedAt must not move.
	svc.now = func() time.Time { return first.LastUsedAt.Add(time.Minute) }
	if _, _, err := svc.Authenticate(plaintext); err != nil {
		t.Fatal(err)
	}
	second, _ := svc.keys.Find(key.ID)
	if !second.LastUsedAt.Equal(first.LastUsedAt) {
		t.Fatalf("touch not throttled: %v -> %v", first.LastUsedAt, second.LastUsedAt)
	}

	// Past the window it moves.
	svc.now = func() time.Time { return first.LastUsedAt.Add(10 * time.Minute) }
	if _, _, err := svc.Authenticate(plaintext); err != nil {
		t.Fatal(err)
	}
	third, _ := svc.keys.Find(key.ID)
	if !third.LastUsedAt.After(first.LastUsedAt) {
		t.Fatalf("expected touch after window, got %v", third.LastUsedAt)
	}
}
