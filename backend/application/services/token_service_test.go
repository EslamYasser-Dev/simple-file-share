package services

import (
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

func newTokenServiceFixture(t *testing.T) (*TokenService, *fs.UserFileRepository, *auth.PBKDF2Hasher) {
	t.Helper()
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	provider := auth.NewUserAuthProvider(userRepo, hasher)
	jwt := auth.NewJWTManager("test-secret", time.Hour)
	tokens := NewTokenService(NewAuthenticateService(provider), jwt, userRepo)
	return tokens, userRepo, hasher
}

func TestTokenServiceLoginAndAuthenticate(t *testing.T) {
	tokens, users, hasher := newTokenServiceFixture(t)
	hash, err := hasher.Hash("secret")
	if err != nil {
		t.Fatal(err)
	}
	if err := users.CreateUser(&models.User{Username: "alice", PasswordHash: hash, Enabled: true, CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}

	pair, err := tokens.Login("alice", "secret")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if pair.AccessToken == "" || pair.ExpiresAt.IsZero() {
		t.Fatalf("pair = %+v", pair)
	}

	user, err := tokens.Authenticate(pair.AccessToken)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if user.Username != "alice" {
		t.Fatalf("user = %q", user.Username)
	}
}

func TestTokenServiceLoginRejectsBadPassword(t *testing.T) {
	tokens, users, hasher := newTokenServiceFixture(t)
	hash, _ := hasher.Hash("secret")
	_ = users.CreateUser(&models.User{Username: "alice", PasswordHash: hash, Enabled: true, CreatedAt: time.Now().UTC()})

	if _, err := tokens.Login("alice", "wrong"); err == nil {
		t.Fatal("expected invalid credentials")
	}
}

func TestTokenServiceRefreshRotates(t *testing.T) {
	tokens, users, hasher := newTokenServiceFixture(t)
	hash, _ := hasher.Hash("secret")
	_ = users.CreateUser(&models.User{Username: "alice", PasswordHash: hash, Enabled: true, CreatedAt: time.Now().UTC()})

	pair, err := tokens.Login("alice", "secret")
	if err != nil {
		t.Fatal(err)
	}
	refreshed, err := tokens.Refresh(pair.AccessToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.AccessToken == pair.AccessToken {
		t.Fatal("expected a new token after refresh")
	}
	if _, err := tokens.Authenticate(pair.AccessToken); err == nil {
		t.Fatal("expected old token revoked after refresh")
	}
	if _, err := tokens.Authenticate(refreshed.AccessToken); err != nil {
		t.Fatalf("new token should authenticate: %v", err)
	}
}

func TestTokenServiceRevoke(t *testing.T) {
	tokens, users, hasher := newTokenServiceFixture(t)
	hash, _ := hasher.Hash("secret")
	_ = users.CreateUser(&models.User{Username: "alice", PasswordHash: hash, Enabled: true, CreatedAt: time.Now().UTC()})

	pair, _ := tokens.Login("alice", "secret")
	if err := tokens.Revoke(pair.AccessToken); err != nil {
		t.Fatal(err)
	}
	if _, err := tokens.Authenticate(pair.AccessToken); err == nil {
		t.Fatal("expected revoked token to fail")
	}
}

func TestOAuthLoginServiceRejectsInvalidState(t *testing.T) {
	dir := t.TempDir()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memory.NewFileIndexRepository(), nil)
	userRepo := fs.NewUserFileRepository(dir)
	scoper := policy.NewPathScoper()
	hasher := auth.NewPBKDF2Hasher()
	jwt := auth.NewJWTManager("s", time.Hour)
	providerAuth := auth.NewUserAuthProvider(userRepo, hasher)
	tokens := NewTokenService(NewAuthenticateService(providerAuth), jwt, userRepo)

	// Empty provider map: Complete must reject unknown state before exchange.
	svc := NewOAuthLoginService(map[string]*auth.OAuthProvider{}, userRepo, fileRepo, scoper, tokens, true, 0)
	if _, err := svc.Complete("github", "code", "bogus", "bogus", "http://localhost/cb"); err == nil {
		t.Fatal("expected invalid state error")
	}
	// Cookie/query state mismatch is rejected before any map lookup.
	if _, err := svc.Complete("github", "code", "aaa", "bbb", "http://localhost/cb"); err == nil {
		t.Fatal("expected state cookie mismatch error")
	}
}

func TestSanitizeOAuthUsername(t *testing.T) {
	cases := map[string]string{
		"Alice Smith": "alice-smith",
		"bob@users":   "bob-users",
		"XY":          "",
		"":            "",
		"valid.user":  "valid.user",
	}
	for in, want := range cases {
		if got := sanitizeOAuthUsername(in); got != want {
			t.Errorf("sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}
