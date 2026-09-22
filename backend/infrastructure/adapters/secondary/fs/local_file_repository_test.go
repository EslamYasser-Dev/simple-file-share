package fs

import (
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestLocalFileRepository_KeepsFileVersions(t *testing.T) {
	root := t.TempDir()
	repo := NewLocalFileRepository(root)

	write := func(content string) {
		t.Helper()
		n, err := repo.WriteFile("doc.txt", io.NopCloser(strings.NewReader(content)))
		if err != nil {
			t.Fatalf("write: %v", err)
		}
		if n != int64(len(content)) {
			t.Fatalf("wrote %d bytes, want %d", n, len(content))
		}
	}

	write("v1")
	write("v2")
	write("v3")

	// The canonical file holds the latest content and every overwrite keeps a
	// numbered snapshot in the hidden `<file>.versions` sibling directory.
	info, err := repo.GetFileInfo("doc.txt")
	if err != nil {
		t.Fatalf("get info: %v", err)
	}
	if info.Version != 2 {
		t.Fatalf("version = %d, want 2", info.Version)
	}

	versionDir := filepath.Join(root, "doc.txt.versions")
	entries, err := os.ReadDir(versionDir)
	if err != nil {
		t.Fatalf("version dir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("snapshots = %d, want 2", len(entries))
	}

	// The versioned snapshots stay invisible to directory listings.
	listing, err := repo.ListDirectory("")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range listing {
		if strings.Contains(entry.Name, ".versions") {
			t.Fatalf("version directory leaked into listing: %q", entry.Name)
		}
	}

	// Deleting the file also removes its entire version history.
	if err := repo.DeletePath("doc.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(versionDir); !os.IsNotExist(err) {
		t.Fatalf("version history not cleaned up on delete: %v", err)
	}
}
