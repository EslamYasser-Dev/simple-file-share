package auth

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

func newTestUser(username, hash string, isAdmin bool) *models.User {
	return &models.User{Username: username, PasswordHash: hash, IsAdmin: isAdmin, CreatedAt: time.Now().UTC()}
}

func TestUserAuthProviderAuthenticates(t *testing.T) {
	hasher := NewPBKDF2Hasher()
	users := fs.NewUserFileRepository(t.TempDir())

	hash, err := hasher.Hash("s3cret")
	if err != nil {
		t.Fatal(err)
	}
	if err := users.CreateUser(newTestUser("alice", hash, true)); err != nil {
		t.Fatal(err)
	}

	provider := NewUserAuthProvider(users, hasher)

	user, err := provider.Authenticate("alice", "s3cret")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if user.Username != "alice" || !user.IsAdmin {
		t.Errorf("authenticated user = %+v", user)
	}

	if _, err := provider.Authenticate("alice", "wrong"); !errors.Is(err, domainerrors.ErrInvalidCredentials) {
		t.Errorf("wrong password err = %v, want ErrInvalidCredentials", err)
	}
	if _, err := provider.Authenticate("nobody", "s3cret"); err == nil {
		t.Error("unknown user authenticated successfully")
	}
}
