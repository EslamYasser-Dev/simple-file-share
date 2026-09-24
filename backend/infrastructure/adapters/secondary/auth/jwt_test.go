package auth

import (
	"path/filepath"
	"testing"
	"time"
)

func TestJWTIssueAndVerify(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	token, claims, err := m.Issue("alice")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if claims.Subject != "alice" || claims.ID == "" {
		t.Fatalf("claims = %+v", claims)
	}

	got, err := m.Verify(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.Subject != "alice" || got.ID != claims.ID {
		t.Fatalf("verified = %+v, want %+v", got, claims)
	}
}

func TestJWTRejectsTamperedToken(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	token, _, err := m.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}

	other := NewJWTManager("other-secret", time.Hour)
	if _, err := other.Verify(token); err == nil {
		t.Fatal("expected wrong secret to fail")
	}

	if _, err := m.Verify(token[:len(token)-2] + "xx"); err == nil {
		t.Fatal("expected tampered signature to fail")
	}
	if _, err := m.Verify("not-a-token"); err == nil {
		t.Fatal("expected malformed token to fail")
	}
}

func TestJWTRejectsExpired(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	base := time.Now()
	m.now = func() time.Time { return base }
	token, _, err := m.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}

	m.now = func() time.Time { return base.Add(2 * time.Hour) }
	if _, err := m.Verify(token); err != ErrExpiredToken {
		t.Fatalf("err = %v, want ErrExpiredToken", err)
	}
}

func TestJWTRevoke(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	token, claims, err := m.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Verify(token); err != nil {
		t.Fatalf("pre-revoke verify: %v", err)
	}

	m.Revoke(claims.ID, claims.ExpiresAt)
	if _, err := m.Verify(token); err == nil {
		t.Fatal("expected revoked token to fail")
	}
}

func TestJWTRevocationSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "revocations.json")
	m1 := NewJWTManagerWithStore("test-secret", time.Hour, path)
	token, claims, err := m1.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}
	m1.Revoke(claims.ID, claims.ExpiresAt)

	m2 := NewJWTManagerWithStore("test-secret", time.Hour, path)
	if _, err := m2.Verify(token); err == nil {
		t.Fatal("revocation must persist across manager restart")
	}
}

func TestJWTRevokeSubjectBlocksOldAllowsNew(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	old, _, err := m.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}
	m.RevokeSubject("alice", time.Now().Add(24*time.Hour))
	if _, err := m.Verify(old); err == nil {
		t.Fatal("token issued before RevokeSubject must fail")
	}
	fresh, _, err := m.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Verify(fresh); err != nil {
		t.Fatalf("token issued after RevokeSubject must verify: %v", err)
	}
}
