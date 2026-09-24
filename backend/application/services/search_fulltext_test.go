package services

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

func TestSearchIncludesFullTextContentMatches(t *testing.T) {
	index := memory.NewFileIndexRepository()
	text := memory.NewTextIndex()
	scoper := policy.NewPathScoper()

	// Filename does not contain the query; content does.
	if err := index.Upsert(&models.FileInfo{
		Name: "notes.md", Path: "notes.md", Size: 20, Modified: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := text.Index("notes.md", "the quarterly kalmar merger closed"); err != nil {
		t.Fatal(err)
	}
	// Unrelated file that should not match.
	if err := index.Upsert(&models.FileInfo{
		Name: "other.txt", Path: "other.txt", Size: 5, Modified: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}
	_ = text.Index("other.txt", "unrelated groceries")

	svc := NewSearchFilesService(index, scoper, text)
	results, err := svc.Execute(nil, "kalmar", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Path != "notes.md" {
		t.Fatalf("content search results = %+v, want notes.md", results)
	}
}

func TestSearchNameMatchesStillRankFirst(t *testing.T) {
	index := memory.NewFileIndexRepository()
	text := memory.NewTextIndex()
	scoper := policy.NewPathScoper()

	_ = index.Upsert(&models.FileInfo{Name: "kalmar.txt", Path: "kalmar.txt", Size: 1, Modified: time.Now()})
	_ = text.Index("kalmar.txt", "alpha")
	_ = index.Upsert(&models.FileInfo{Name: "notes.md", Path: "notes.md", Size: 1, Modified: time.Now()})
	_ = text.Index("notes.md", "mentions kalmar inside content")

	svc := NewSearchFilesService(index, scoper, text)
	results, err := svc.Execute(nil, "kalmar", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) < 2 {
		t.Fatalf("results = %+v, want name + content hits", results)
	}
	if results[0].Path != "kalmar.txt" {
		t.Fatalf("first result = %s, want name match kalmar.txt", results[0].Path)
	}
}

func TestSearchScopeFiltersContentHits(t *testing.T) {
	index := memory.NewFileIndexRepository()
	text := memory.NewTextIndex()
	scoper := policy.NewPathScoper()

	_ = index.Upsert(&models.FileInfo{
		Name: "alice-secret.txt", Path: "users/alice/alice-secret.txt", Size: 1, Modified: time.Now(),
	})
	_ = text.Index("users/alice/alice-secret.txt", "classified alpha content")
	_ = index.Upsert(&models.FileInfo{
		Name: "shared-note.txt", Path: "shared/shared-note.txt", Size: 1, Modified: time.Now(),
	})
	_ = text.Index("shared/shared-note.txt", "classified shared content")

	alice := &models.User{Username: "alice"}
	svc := NewSearchFilesService(index, scoper, text)
	results, err := svc.Execute(alice, "classified", 10)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, info := range results {
		got[info.Path] = true
	}
	if !got["alice-secret.txt"] || !got["shared/shared-note.txt"] {
		t.Fatalf("alice should see own+shared content hits, got %v", got)
	}
	if got["users/alice/alice-secret.txt"] && !got["alice-secret.txt"] {
		// PhysicalToVirtual should remap — assert virtual form present.
		t.Fatalf("paths should be virtual: %v", got)
	}
}

func TestSearchWithoutTextIndexIsNameOnly(t *testing.T) {
	index := memory.NewFileIndexRepository()
	scoper := policy.NewPathScoper()
	_ = index.Upsert(&models.FileInfo{Name: "a.txt", Path: "a.txt", Size: 1, Modified: time.Now()})

	svc := NewSearchFilesService(index, scoper, nil)
	results, err := svc.Execute(nil, "a", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %+v", results)
	}
	// Content-only term with nil text index must not invent hits.
	results, err = svc.Execute(nil, "nonexistentterm", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("nil text index still returned %d hits", len(results))
	}
}

func TestIndexedRepositoryWriteIndexesContent(t *testing.T) {
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	text := memory.NewTextIndex()
	repo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index, text)

	body := "important zanzibar keyword"
	if _, err := repo.WriteFile("docs/note.txt", io.NopCloser(strings.NewReader(body))); err != nil {
		t.Fatal(err)
	}

	hits, err := text.Search("zanzibar", 10)
	if err != nil || len(hits) != 1 || hits[0].Path != "docs/note.txt" {
		t.Fatalf("content hits = %+v err=%v", hits, err)
	}

	if err := repo.DeletePath("docs/note.txt"); err != nil {
		t.Fatal(err)
	}
	hits, _ = text.Search("zanzibar", 10)
	if len(hits) != 0 {
		t.Fatalf("delete left content hits: %+v", hits)
	}
}
