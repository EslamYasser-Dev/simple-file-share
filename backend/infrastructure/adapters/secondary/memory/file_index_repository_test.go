package memory

import (
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

func TestFileIndexRepository_Search(t *testing.T) {
	repo := NewFileIndexRepository()

	entries := []*models.FileInfo{
		{Name: "report.pdf", Path: "docs/report.pdf", Size: 100, Modified: time.Now()},
		{Name: "notes", Path: "docs/notes", Size: 0, IsDir: true, Modified: time.Now()},
		{Name: "photo.jpg", Path: "images/photo.jpg", Size: 200, Modified: time.Now()},
	}
	if err := repo.Rebuild(entries); err != nil {
		t.Fatal(err)
	}

	results, err := repo.Search("report", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "report.pdf" {
		t.Fatalf("unexpected search results: %+v", results)
	}

	results, err = repo.Search("docs", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 matches for docs, got %d", len(results))
	}
}

func TestFileIndexRepository_RemovePrefix(t *testing.T) {
	repo := NewFileIndexRepository()

	entries := []*models.FileInfo{
		{Name: "a.txt", Path: "docs/a.txt", Size: 1, Modified: time.Now()},
		{Name: "b.txt", Path: "docs/nested/b.txt", Size: 1, Modified: time.Now()},
		{Name: "c.txt", Path: "other/c.txt", Size: 1, Modified: time.Now()},
	}
	if err := repo.Rebuild(entries); err != nil {
		t.Fatal(err)
	}

	if err := repo.RemovePrefix("docs"); err != nil {
		t.Fatal(err)
	}

	results, err := repo.Search("txt", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Path != "other/c.txt" {
		t.Fatalf("unexpected remaining entries: %+v", results)
	}
}
