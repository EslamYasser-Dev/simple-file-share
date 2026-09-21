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

func fileInfo(path string, size int64, isDir bool) *models.FileInfo {
	return &models.FileInfo{Name: path, Path: path, Size: size, IsDir: isDir, Modified: time.Now()}
}

func newUserFixture(t *testing.T, signupEnabled bool) (*RegisterUserService, *fs.UserFileRepository, *fs.IndexedFileRepository, *SeedAdminService) {
	t.Helper()
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index)
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	scoper := policy.NewPathScoper()

	register := NewRegisterUserService(userRepo, hasher, fileRepo, scoper, signupEnabled)
	seed := NewSeedAdminService(userRepo, hasher, fileRepo, scoper)
	return register, userRepo, fileRepo, seed
}

func TestRegisterFirstUserBecomesAdmin(t *testing.T) {
	register, _, _, _ := newUserFixture(t, true)

	user, err := register.Execute("alice", "secret")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if !user.IsAdmin {
		t.Error("first user should be admin")
	}
	if user.PasswordHash != "" {
		t.Error("returned user should not expose the password hash")
	}

	second, err := register.Execute("bob", "secret")
	if err != nil {
		t.Fatalf("register second: %v", err)
	}
	if second.IsAdmin {
		t.Error("second user should not be admin")
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	register, _, _, _ := newUserFixture(t, true)

	if _, err := register.Execute("ab", "secret"); err == nil {
		t.Error("short username should be rejected")
	}
	if _, err := register.Execute("bad name", "secret"); err == nil {
		t.Error("username with a space should be rejected")
	}
	if _, err := register.Execute("carol", "123"); err == nil {
		t.Error("short password should be rejected")
	}
}

func TestRegisterRejectsDuplicate(t *testing.T) {
	register, _, _, _ := newUserFixture(t, true)
	if _, err := register.Execute("alice", "secret"); err != nil {
		t.Fatal(err)
	}
	_, err := register.Execute("alice", "other")
	if !errors.Is(err, domainerrors.ErrUserAlreadyExists) {
		t.Errorf("err = %v, want ErrUserAlreadyExists", err)
	}
}

func TestRegisterDisabled(t *testing.T) {
	register, _, _, _ := newUserFixture(t, false)
	_, err := register.Execute("alice", "secret")
	var forbidden *domainerrors.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Errorf("err = %v, want ForbiddenError", err)
	}
}

func TestRegisterCreatesPrivateHome(t *testing.T) {
	register, _, fileRepo, _ := newUserFixture(t, true)
	if _, err := register.Execute("alice", "secret"); err != nil {
		t.Fatal(err)
	}
	isDir, err := fileRepo.IsDirectory("users/alice")
	if err != nil {
		t.Fatal(err)
	}
	if !isDir {
		t.Error("expected private home directory to be created")
	}
}

func TestSeedAdminOnlyOnce(t *testing.T) {
	_, _, _, seed := newUserFixture(t, true)

	seeded, err := seed.Execute("admin", "admin123")
	if err != nil {
		t.Fatal(err)
	}
	if !seeded {
		t.Fatal("expected seed to create the admin")
	}

	seeded, err = seed.Execute("admin", "admin123")
	if err != nil {
		t.Fatal(err)
	}
	if seeded {
		t.Error("second seed should be a no-op")
	}
}

func TestListUsersReturnsStorageStats(t *testing.T) {
	register, userRepo, _, _ := newUserFixture(t, true)
	if _, err := register.Execute("alice", "secret"); err != nil {
		t.Fatal(err)
	}

	index := memory.NewFileIndexRepository()
	if err := index.Upsert(fileInfo("users/alice/a.txt", 10, false)); err != nil {
		t.Fatal(err)
	}
	if err := index.Upsert(fileInfo("users/alice/b.txt", 5, false)); err != nil {
		t.Fatal(err)
	}
	service := NewListUsersService(userRepo, index, policy.NewPathScoper())

	stats, err := service.Execute()
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 {
		t.Fatalf("got %d users, want 1", len(stats))
	}
	if stats[0].Files != 2 || stats[0].Size != 15 {
		t.Errorf("stats = %+v, want 2 files / 15 bytes", stats[0])
	}
}
