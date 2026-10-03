package services

import (
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

func newDirectoryFixture(t *testing.T) (*DirectoryService, *models.User) {
	t.Helper()
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	for _, name := range []string{"alice", "bob", "bobby", "carol"} {
		if err := userRepo.CreateUser(&models.User{Username: name}); err != nil {
			t.Fatalf("create user: %v", err)
		}
	}
	alice, _ := userRepo.FindByUsername("alice")
	return NewDirectoryService(userRepo), alice
}

// TestDirectorySearch verifies substring matching, self-exclusion,
// validation, and that only usernames leak.
func TestDirectorySearch(t *testing.T) {
	svc, alice := newDirectoryFixture(t)

	got, err := svc.Search(alice, "bob")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 2 || got[0] != "bob" || got[1] != "bobby" {
		t.Fatalf("search = %v", got)
	}

	got, err = svc.Search(alice, "ALICE")
	if err != nil || len(got) != 0 {
		t.Fatalf("self must be excluded: %v err=%v", got, err)
	}

	if _, err := svc.Search(alice, ""); err == nil {
		t.Fatal("empty query must be rejected")
	}
	if _, err := svc.Search(alice, "CAR"); err != nil {
		t.Fatalf("case-insensitive: %v", err)
	}
}
