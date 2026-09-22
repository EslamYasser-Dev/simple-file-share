package fs

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

func newShare(t *testing.T, token string, expiresAt time.Time) *models.Share {
	t.Helper()
	return &models.Share{
		Token:     token,
		Path:      "users/alice/notes.txt",
		Owner:     "alice",
		CreatedAt: time.Now().Add(-time.Hour),
		ExpiresAt: expiresAt,
	}
}

func TestShareFileRepositoryCRUD(t *testing.T) {
	repo := NewShareFileRepository(t.TempDir())

	if _, err := repo.FindByToken("nope"); !errors.Is(err, domainerrors.ErrShareNotFound) {
		t.Fatalf("FindByToken(missing) = %v, want ErrShareNotFound", err)
	}
	if err := repo.Delete("nope"); !errors.Is(err, domainerrors.ErrShareNotFound) {
		t.Fatalf("Delete(missing) = %v, want ErrShareNotFound", err)
	}

	a := newShare(t, "tok-a", time.Time{})
	b := newShare(t, "tok-b", time.Now().Add(24*time.Hour))
	b.Owner = "bob"
	b.Path = "users/bob/file.txt"
	if err := repo.Create(a); err != nil {
		t.Fatalf("create a: %v", err)
	}
	if err := repo.Create(b); err != nil {
		t.Fatalf("create b: %v", err)
	}

	got, err := repo.FindByToken("tok-a")
	if err != nil {
		t.Fatalf("find a: %v", err)
	}
	if got.Path != "users/alice/notes.txt" || got.Owner != "alice" {
		t.Errorf("share a = %+v", got)
	}
	if !got.ExpiresAt.IsZero() {
		t.Errorf("share a should not expire, got %v", got.ExpiresAt)
	}

	alice, err := repo.ListByOwner("alice")
	if err != nil {
		t.Fatalf("list alice: %v", err)
	}
	if len(alice) != 1 || alice[0].Token != "tok-a" {
		t.Errorf("alice shares = %v, want [tok-a]", alice)
	}

	all, err := repo.ListAll()
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("all shares = %d, want 2", len(all))
	}

	if err := repo.Delete("tok-b"); err != nil {
		t.Fatalf("delete b: %v", err)
	}
	if _, err := repo.FindByToken("tok-b"); !errors.Is(err, domainerrors.ErrShareNotFound) {
		t.Fatalf("find deleted b = %v, want ErrShareNotFound", err)
	}
}

func TestShareFileRepositoryReplacesSameToken(t *testing.T) {
	repo := NewShareFileRepository(t.TempDir())

	original := newShare(t, "same", time.Time{})
	if err := repo.Create(original); err != nil {
		t.Fatalf("create: %v", err)
	}

	replacement := newShare(t, "same", time.Now().Add(time.Hour))
	replacement.Path = "users/alice/updated.txt"
	if err := repo.Create(replacement); err != nil {
		t.Fatalf("replace: %v", err)
	}

	all, err := repo.ListAll()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("shares = %d, want 1", len(all))
	}
	if all[0].Path != "users/alice/updated.txt" {
		t.Errorf("path = %q, want users/alice/updated.txt", all[0].Path)
	}
}

func TestShareFileRepositoryPurgeExpired(t *testing.T) {
	repo := NewShareFileRepository(t.TempDir())
	now := time.Now()

	for i, expires := range []time.Time{
		now.Add(-time.Minute), // expired
		now.Add(time.Minute),  // still valid
		time.Time{},           // never expires
	} {
		late := expires
		if err := repo.Create(newShare(t, string(rune('a'+i)), late)); err != nil {
			t.Fatalf("create share: %v", err)
		}
	}

	removed, err := repo.PurgeExpired(now)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	all, err := repo.ListAll()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("shares remaining = %d, want 2", len(all))
	}
}
