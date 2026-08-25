package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalFileRepository_PathTraversalBlocked(t *testing.T) {
	root := t.TempDir()
	repo := NewLocalFileRepository(root)

	_, _, err := repo.ServeFile("../../etc/passwd")
	if err == nil {
		t.Fatal("expected path traversal to be blocked")
	}

	safeFile := filepath.Join(root, "safe.txt")
	if err := os.WriteFile(safeFile, []byte("ok"), 0644); err != nil {
		t.Fatal(err)
	}

	stream, name, err := repo.ServeFile("safe.txt")
	if err != nil {
		t.Fatalf("expected safe read: %v", err)
	}
	stream.Close()
	if name != "safe.txt" {
		t.Fatalf("name = %q", name)
	}
}

func TestLocalFileRepository_ListAndDelete(t *testing.T) {
	root := t.TempDir()
	repo := NewLocalFileRepository(root)

	if err := repo.CreateDirectory("nested"); err != nil {
		t.Fatal(err)
	}

	files, err := repo.ListDirectory("")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || !files[0].IsDir {
		t.Fatalf("unexpected listing: %+v", files)
	}

	if err := repo.DeletePath("nested"); err != nil {
		t.Fatal(err)
	}
}
