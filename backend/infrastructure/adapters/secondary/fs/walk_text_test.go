package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkRootTextDocuments(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "note.md"), []byte("alpha content here"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docs", "photo.jpg"), []byte("\xff\xd8 binary"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".file-share"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".file-share", "users.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	docs, err := WalkRootTextDocuments(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("docs = %+v, want only docs/note.md", docs)
	}
	if docs[0].Path != "docs/note.md" || docs[0].Content != "alpha content here" {
		t.Fatalf("doc = %+v", docs[0])
	}
}
