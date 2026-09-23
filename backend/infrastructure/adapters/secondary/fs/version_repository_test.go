package fs

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalFileRepository_ListServeRestoreVersions(t *testing.T) {
	root := t.TempDir()
	repo := NewLocalFileRepository(root)

	write := func(content string) {
		t.Helper()
		if _, err := repo.WriteFile("doc.txt", io.NopCloser(strings.NewReader(content))); err != nil {
			t.Fatalf("write %q: %v", content, err)
		}
	}
	write("v1")
	write("v2")
	write("v3")

	versions, err := repo.ListVersions("doc.txt")
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("versions = %d, want 2", len(versions))
	}
	if versions[0].Version != 1 || versions[1].Version != 2 {
		t.Fatalf("version numbers = %d,%d want 1,2", versions[0].Version, versions[1].Version)
	}
	if versions[0].Name != "doc.txt" {
		t.Fatalf("version name = %q, want doc.txt", versions[0].Name)
	}

	stream, name, err := repo.ServeVersion("doc.txt", 1)
	if err != nil {
		t.Fatalf("serve version 1: %v", err)
	}
	data, _ := io.ReadAll(stream)
	stream.Close()
	if string(data) != "v1" {
		t.Fatalf("version 1 content = %q, want v1", data)
	}
	if name != "doc.txt" {
		t.Fatalf("serve name = %q, want doc.txt", name)
	}

	if _, _, err := repo.ServeVersion("doc.txt", 99); err == nil {
		t.Fatal("expected missing version to fail")
	}

	if err := repo.RestoreVersion("doc.txt", 1); err != nil {
		t.Fatalf("restore: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "doc.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "v1" {
		t.Fatalf("restored content = %q, want v1", got)
	}

	versions, err = repo.ListVersions("doc.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 3 {
		t.Fatalf("after restore versions = %d, want 3 (old current snapshotted)", len(versions))
	}
}

func TestLocalFileRepository_VersionRetention(t *testing.T) {
	root := t.TempDir()
	repo := NewLocalFileRepository(root)
	repo.SetVersionKeep(2)

	for i := 0; i < 6; i++ {
		if _, err := repo.WriteFile("doc.txt", io.NopCloser(strings.NewReader(strings.Repeat("x", i+1)))); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}

	versions, err := repo.ListVersions("doc.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 {
		t.Fatalf("retained versions = %d, want 2", len(versions))
	}
	if versions[0].Version != 4 || versions[1].Version != 5 {
		t.Fatalf("retained numbers = %d,%d want newest 4,5", versions[0].Version, versions[1].Version)
	}
}

func TestLocalFileRepository_ListVersionsEdgeCases(t *testing.T) {
	root := t.TempDir()
	repo := NewLocalFileRepository(root)

	if _, err := repo.ListVersions("missing.txt"); err == nil {
		t.Fatal("expected missing file to fail")
	}

	if err := repo.CreateDirectory("folder"); err != nil {
		t.Fatal(err)
	}
	versions, err := repo.ListVersions("folder")
	if err != nil {
		t.Fatalf("list versions of dir: %v", err)
	}
	if versions != nil {
		t.Fatalf("dir versions = %+v, want nil", versions)
	}

	if err := repo.RestoreVersion("../escape.txt", 1); err == nil {
		t.Fatal("expected path traversal to fail")
	}
}
