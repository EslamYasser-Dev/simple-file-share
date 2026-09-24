package services

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

func newOAuthService(t *testing.T) (*OAuthLoginService, *fs.UserFileRepository, *TokenService) {
	t.Helper()
	dir := t.TempDir()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memory.NewFileIndexRepository(), nil)
	userRepo := fs.NewUserFileRepository(dir)
	scoper := policy.NewPathScoper()
	hasher := auth.NewPBKDF2Hasher()
	provider := auth.NewUserAuthProvider(userRepo, hasher)
	jwt := auth.NewJWTManager("oauth-test-secret", time.Hour)
	tokens := NewTokenService(NewAuthenticateService(provider), jwt, userRepo)
	svc := NewOAuthLoginService(map[string]*auth.OAuthProvider{}, userRepo, fileRepo, scoper, tokens, true, 0)
	return svc, userRepo, tokens
}

func TestOAuthUpsertRefusesCrossProviderUsernameTakeover(t *testing.T) {
	svc, users, _ := newOAuthService(t)

	// GitHub user "alice" creates an OAuth-linked account.
	if _, err := svc.upsert("github", &auth.OAuthProfile{Username: "alice", Subject: "111"}); err != nil {
		t.Fatalf("github upsert: %v", err)
	}

	// A different Google identity that sanitizes to "alice" must not take over;
	// it either fails or gets a distinct username.
	gu, gerr := svc.upsert("google", &auth.OAuthProfile{Username: "alice", Subject: "gid-1"})
	if gerr == nil && gu.Username == "alice" {
		t.Fatal("google identity must not claim github alice's username")
	}

	// Same github subject may log in again under the original name.
	u, err := svc.upsert("github", &auth.OAuthProfile{Username: "alice", Subject: "111"})
	if err != nil {
		t.Fatalf("same subject re-login: %v", err)
	}
	if u.Username != "alice" || u.OAuthProvider != "github" || u.OAuthSubject != "111" {
		t.Fatalf("user = %+v", u)
	}

	// Original account must still be bound to github subject 111.
	orig, err := users.FindByUsername("alice")
	if err != nil {
		t.Fatal(err)
	}
	if orig.OAuthProvider != "github" || orig.OAuthSubject != "111" || orig.PasswordHash != "" {
		t.Fatalf("orig = %+v", orig)
	}
	if gerr == nil && gu != nil {
		if gu.OAuthProvider != "google" || gu.OAuthSubject != "gid-1" {
			t.Fatalf("renamed google user = %+v", gu)
		}
		if gu.Username == "alice" {
			t.Fatal("google user must not share username alice")
		}
	}
}

func TestOAuthUpsertRefusesPasswordAccount(t *testing.T) {
	svc, users, _ := newOAuthService(t)
	hash, _ := auth.NewPBKDF2Hasher().Hash("secret")
	if err := users.CreateUser(&models.User{Username: "alice", PasswordHash: hash, Enabled: true, CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}

	_, err := svc.upsert("github", &auth.OAuthProfile{Username: "alice", Subject: "999"})
	if err == nil {
		t.Fatal("expected refusal for password account")
	}
	var forbidden *domainerrors.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Fatalf("err = %v, want ForbiddenError", err)
	}
}

func TestOAuthUpsertRequiresSubject(t *testing.T) {
	svc, _, _ := newOAuthService(t)
	if _, err := svc.upsert("github", &auth.OAuthProfile{Username: "alice"}); err == nil {
		t.Fatal("expected error for missing subject")
	}
}
