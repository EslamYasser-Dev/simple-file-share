package fs

import (
	"errors"
	"os"
	"testing"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

func sampleKey(id, owner, scope string) models.APIKey {
	return models.APIKey{
		ID:         id,
		Name:       "ci",
		Owner:      owner,
		SecretHash: "pbkdf2$hash",
		Scope:      scope,
		CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestAPIKeyRepositoryCRUDAndPersistence(t *testing.T) {
	dir := t.TempDir()
	repo := NewAPIKeyFileRepository(dir)

	key := sampleKey("0123456789abcdef", "alice", models.APIKeyScopeWrite)
	if err := repo.Create(key); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(key); err == nil {
		t.Fatal("duplicate id must fail")
	}
	if err := repo.Create(models.APIKey{Owner: "alice"}); err == nil {
		t.Fatal("missing id must fail")
	}

	found, err := repo.Find(key.ID)
	if err != nil {
		t.Fatal(err)
	}
	if found.SecretHash != key.SecretHash || found.Owner != "alice" {
		t.Fatalf("find returned %+v", found)
	}
	if _, err := repo.Find("ffffffffffffffff"); err == nil {
		t.Fatal("unknown id must fail")
	}

	// Save updates LastUsedAt in place.
	updated := *found
	updated.LastUsedAt = time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if err := repo.Save(updated); err != nil {
		t.Fatal(err)
	}
	got, _ := repo.Find(key.ID)
	if !got.LastUsedAt.Equal(updated.LastUsedAt) {
		t.Fatalf("last used = %v, want %v", got.LastUsedAt, updated.LastUsedAt)
	}
	if err := repo.Save(sampleKey("ffffffffffffffff", "x", models.APIKeyScopeRead)); err == nil {
		t.Fatal("save of unknown id must fail")
	}

	// Owner filter.
	other := sampleKey("aaaabbbbccccdddd", "bob", models.APIKeyScopeRead)
	other.CreatedAt = key.CreatedAt.Add(time.Hour)
	if err := repo.Create(other); err != nil {
		t.Fatal(err)
	}
	aliceKeys, _ := repo.List("alice")
	if len(aliceKeys) != 1 || aliceKeys[0].Owner != "alice" {
		t.Fatalf("alice list = %+v", aliceKeys)
	}
	allKeys, _ := repo.List("")
	if len(allKeys) != 2 {
		t.Fatalf("all list = %d keys, want 2", len(allKeys))
	}
	if allKeys[0].ID != other.ID {
		t.Fatalf("list must be newest first, got %s first", allKeys[0].ID)
	}

	// Persistence across instances: same root reloads, other root is empty.
	same := NewAPIKeyFileRepository(dir)
	if keys, err := same.List(""); err != nil || len(keys) != 2 {
		t.Fatalf("reload: %v %d keys", err, len(keys))
	}
	fresh := NewAPIKeyFileRepository(t.TempDir())
	if keys, _ := fresh.List(""); len(keys) != 0 {
		t.Fatalf("fresh root has %d keys", len(keys))
	}

	// Delete.
	if err := repo.Delete(key.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Find(key.ID); err == nil {
		t.Fatal("deleted key must be gone")
	}
	if err := repo.Delete(key.ID); err == nil {
		t.Fatal("deleting a missing key must fail")
	}
	if keys, _ := repo.List(""); len(keys) != 1 {
		t.Fatalf("after delete: %d keys", len(keys))
	}
}

func TestAPIKeyRepositoryCorruptFileResetsEmpty(t *testing.T) {
	dir := t.TempDir()
	repo := NewAPIKeyFileRepository(dir)
	if err := repo.Create(sampleKey("0123456789abcdef", "alice", models.APIKeyScopeRead)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(repo.path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	keys, err := repo.List("")
	if err != nil || len(keys) != 0 {
		t.Fatalf("corrupt file: err=%v keys=%d", err, len(keys))
	}
	// And the store recovers on the next write.
	if err := repo.Create(sampleKey("aaaabbbbccccdddd", "bob", models.APIKeyScopeRead)); err != nil {
		t.Fatal(err)
	}
	if keys, _ := repo.List(""); len(keys) != 1 {
		t.Fatalf("recovery write: %d keys", len(keys))
	}
}

func TestAPIKeyRepositoryErrorsAreNotFound(t *testing.T) {
	repo := NewAPIKeyFileRepository(t.TempDir())
	if _, err := repo.Find("0000000000000000"); !isNotFound(err) {
		t.Fatalf("Find err = %v, want ErrNotFound", err)
	}
	if err := repo.Delete("0000000000000000"); !isNotFound(err) {
		t.Fatalf("Delete err = %v, want ErrNotFound", err)
	}
	if err := repo.Save(sampleKey("0000000000000000", "a", models.APIKeyScopeRead)); !isNotFound(err) {
		t.Fatalf("Save err = %v, want ErrNotFound", err)
	}
}

func isNotFound(err error) bool {
	return errors.Is(err, domainerrors.ErrNotFound)
}
