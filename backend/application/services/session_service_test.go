package services

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/sessions"
)

func sessionFixture(t *testing.T) (*SessionService, ports.TokenManager) {
	t.Helper()
	jwt := auth.NewJWTManager("test-secret", time.Hour)
	svc := NewSessionService(sessions.NewRegistry(t.TempDir()), jwt)
	return svc, jwt
}

func TestSessionObserveNilSafe(t *testing.T) {
	var nilSvc *SessionService
	nilSvc.Observe(&TokenPair{ID: "x", Subject: "alice"}, "", "") // must not panic

	disabled := NewSessionService(nil, nil)
	if disabled.Enabled() {
		t.Fatal("nil store must disable the feature")
	}
	disabled.Observe(&TokenPair{ID: "x", Subject: "alice"}, "", "") // must not panic
	disabled.Observe(nil, "", "")                                   // must not panic
	disabled.Observe(&TokenPair{ID: "", Subject: "alice"}, "", "")  // no jti → ignore

	if list, err := disabled.List(nil); err != nil || len(list) != 0 {
		t.Fatalf("disabled list: err=%v list=%d", err, len(list))
	}
}

func TestSessionObserveRecordsAndEnriches(t *testing.T) {
	svc, _ := sessionFixture(t)
	pair := &TokenPair{ID: "jti-1", Subject: "alice", ExpiresAt: time.Now().Add(time.Hour)}

	svc.Observe(pair, "", "")
	svc.Observe(pair, "203.0.113.7", "Mozilla/5.0")

	list, err := svc.List(&models.User{Username: "alice", Role: models.RoleMember, Enabled: true})
	if err != nil || len(list) != 1 {
		t.Fatalf("list: err=%v len=%d", err, len(list))
	}
	if list[0].Remote != "203.0.113.7" || list[0].UserAgent != "Mozilla/5.0" {
		t.Fatalf("enrichment lost: %+v", list[0])
	}
}

func TestSessionListScoping(t *testing.T) {
	svc, jwt := sessionFixture(t)
	token, claims, err := jwt.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}
	_ = token
	svc.Observe(&TokenPair{ID: claims.ID, Subject: "alice", ExpiresAt: claims.ExpiresAt}, "", "")

	// Member sees only own sessions.
	alice := &models.User{Username: "alice", Role: models.RoleMember, Enabled: true}
	list, err := svc.List(alice)
	if err != nil || len(list) != 1 {
		t.Fatalf("alice list: err=%v len=%d", err, len(list))
	}
	// Another member sees nothing of alice's.
	bob := &models.User{Username: "bob", Role: models.RoleMember, Enabled: true}
	list, err = svc.List(bob)
	if err != nil || len(list) != 0 {
		t.Fatalf("bob must see no sessions: err=%v len=%d", err, len(list))
	}
	// System view (auth disabled) sees everything.
	if list, err = svc.List(nil); err != nil || len(list) != 1 {
		t.Fatalf("system view: err=%v len=%d", err, len(list))
	}
}

func TestSessionCurrentJTI(t *testing.T) {
	svc, jwt := sessionFixture(t)
	if svc.CurrentJTI("") != "" || svc.CurrentJTI("garbage") != "" {
		t.Fatal("missing/invalid token must yield empty jti")
	}
	token, claims, err := jwt.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}
	if got := svc.CurrentJTI(token); got != claims.ID {
		t.Fatalf("CurrentJTI = %q, want %q", got, claims.ID)
	}
}

func TestSessionRevokeOwn(t *testing.T) {
	svc, jwt := sessionFixture(t)
	token, claims, err := jwt.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}
	svc.Observe(&TokenPair{ID: claims.ID, Subject: "alice", ExpiresAt: claims.ExpiresAt}, "", "")

	alice := &models.User{Username: "alice", Role: models.RoleMember, Enabled: true}
	if err := svc.Revoke(alice, claims.ID); err != nil {
		t.Fatalf("revoke own: %v", err)
	}
	// Session list no longer shows it and the token itself is dead.
	if list, _ := svc.List(alice); len(list) != 0 {
		t.Fatalf("revoked session still listed: %+v", list)
	}
	if _, err := jwt.Verify(token); err == nil {
		t.Fatal("revoked token must fail verification")
	}
}

func TestSessionRevokeOwnershipAndNotFound(t *testing.T) {
	svc, _ := sessionFixture(t)
	svc.Observe(&TokenPair{ID: "jti-a", Subject: "alice", ExpiresAt: time.Now().Add(time.Hour)}, "", "")

	bob := &models.User{Username: "bob", Role: models.RoleMember, Enabled: true}
	err := svc.Revoke(bob, "jti-a")
	var forbidden *domainerrors.ForbiddenError
	if err == nil || !errors.As(err, &forbidden) {
		t.Fatalf("foreign revoke must be forbidden, got %v", err)
	}
	// The session is untouched.
	if list, _ := svc.List(&models.User{Username: "alice", Role: models.RoleMember, Enabled: true}); len(list) != 1 {
		t.Fatalf("foreign revoke must not remove the session: %+v", list)
	}

	err = svc.Revoke(bob, "does-not-exist")
	var notFound *domainerrors.NotFoundError
	if err == nil || !errors.As(err, &notFound) {
		t.Fatalf("unknown jti must be not found, got %v", err)
	}
}
