package fs

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

func TestUserFileRepositoryCreateAndFind(t *testing.T) {
	repo := NewUserFileRepository(t.TempDir())

	user := &models.User{
		Username:     "alice",
		PasswordHash: "hash",
		IsAdmin:      true,
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
	}
	if err := repo.CreateUser(user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	got, err := repo.FindByUsername("alice")
	if err != nil {
		t.Fatalf("FindByUsername: %v", err)
	}
	if got.Username != "alice" || !got.IsAdmin || got.PasswordHash != "hash" {
		t.Errorf("round-tripped user = %+v", got)
	}
}

func TestUserFileRepositoryDuplicate(t *testing.T) {
	repo := NewUserFileRepository(t.TempDir())
	user := &models.User{Username: "alice", PasswordHash: "hash", CreatedAt: time.Now()}

	if err := repo.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateUser(user); !errors.Is(err, domainerrors.ErrUserAlreadyExists) {
		t.Errorf("err = %v, want ErrUserAlreadyExists", err)
	}
}

func TestUserFileRepositoryMissing(t *testing.T) {
	repo := NewUserFileRepository(t.TempDir())

	if _, err := repo.FindByUsername("nobody"); !errors.Is(err, domainerrors.ErrUserNotFound) {
		t.Errorf("err = %v, want ErrUserNotFound", err)
	}
	count, err := repo.CountUsers()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestUserFileRepositoryPersists(t *testing.T) {
	dir := t.TempDir()
	first := NewUserFileRepository(dir)
	if err := first.CreateUser(&models.User{Username: "bob", PasswordHash: "h", Enabled: true, CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	second := NewUserFileRepository(dir)
	users, err := second.ListUsers()
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].Username != "bob" {
		t.Errorf("reloaded users = %+v", users)
	}
}

func TestUserFileRepositoryQuotaRoundTrip(t *testing.T) {
	dir := t.TempDir()
	repo := NewUserFileRepository(dir)
	if err := repo.CreateUser(&models.User{Username: "quota", PasswordHash: "h", Enabled: true, CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	q, err := repo.GetQuotaBytes("quota")
	if err != nil || q != 0 {
		t.Fatalf("GetQuotaBytes() = %d, %v; want 0, nil", q, err)
	}

	const quota = 2 << 30
	if err := repo.SetQuotaBytes("quota", quota); err != nil {
		t.Fatal(err)
	}

	reloaded := NewUserFileRepository(dir)
	q, err = reloaded.GetQuotaBytes("quota")
	if err != nil {
		t.Fatal(err)
	}
	if q != quota {
		t.Errorf("GetQuotaBytes after reload = %d, want %d", q, quota)
	}
}

func TestUserFileRepositoryQuotaMissingUser(t *testing.T) {
	repo := NewUserFileRepository(t.TempDir())
	if _, err := repo.GetQuotaBytes("nobody"); !errors.Is(err, domainerrors.ErrUserNotFound) {
		t.Errorf("GetQuotaBytes err = %v, want ErrUserNotFound", err)
	}
	if err := repo.SetQuotaBytes("nobody", 100); !errors.Is(err, domainerrors.ErrUserNotFound) {
		t.Errorf("SetQuotaBytes err = %v, want ErrUserNotFound", err)
	}
}
