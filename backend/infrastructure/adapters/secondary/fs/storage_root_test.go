package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareStorageRootCreatesOwnerOnly(t *testing.T) {
	root := filepath.Join(t.TempDir(), "storage")

	got, err := PrepareStorageRoot(root)
	if err != nil {
		t.Fatalf("PrepareStorageRoot() error = %v", err)
	}
	if got != root {
		t.Fatalf("PrepareStorageRoot() = %q, want %q", got, root)
	}

	info, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != storageDirPerm {
		t.Fatalf("permissions = %o, want %o", perm, storageDirPerm)
	}
}

func TestPrepareStorageRootRepairsPermissions(t *testing.T) {
	root := filepath.Join(t.TempDir(), "storage")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := PrepareStorageRoot(root); err != nil {
		t.Fatalf("PrepareStorageRoot() error = %v", err)
	}

	info, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != storageDirPerm {
		t.Fatalf("permissions = %o, want %o", perm, storageDirPerm)
	}
}

func TestPrepareStorageRootRejectsSymlink(t *testing.T) {
	base := t.TempDir()
	target := filepath.Join(base, "target")
	if err := os.MkdirAll(target, storageDirPerm); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if _, err := PrepareStorageRoot(link); !errors.Is(err, ErrUnsafeStorageRoot) {
		t.Fatalf("PrepareStorageRoot(symlink) error = %v, want ErrUnsafeStorageRoot", err)
	}
}

func TestPrepareStorageRootRejectsFilesystemRoot(t *testing.T) {
	if _, err := PrepareStorageRoot(string(os.PathSeparator)); !errors.Is(err, ErrUnsafeStorageRoot) {
		t.Fatalf("PrepareStorageRoot(/) error = %v, want ErrUnsafeStorageRoot", err)
	}
}

func TestPrepareStorageRootRejectsWorkingDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if _, err := PrepareStorageRoot("."); !errors.Is(err, ErrUnsafeStorageRoot) {
		t.Fatalf("PrepareStorageRoot(cwd) error = %v, want ErrUnsafeStorageRoot", err)
	}
}

func TestPrepareStorageRootRejectsHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if _, err := PrepareStorageRoot(home); !errors.Is(err, ErrUnsafeStorageRoot) {
		t.Fatalf("PrepareStorageRoot(home) error = %v, want ErrUnsafeStorageRoot", err)
	}
}

func TestPrepareStorageRootRejectsSourceTree(t *testing.T) {
	for _, marker := range sourceMarkers {
		t.Run(marker, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, marker), []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := PrepareStorageRoot(dir); !errors.Is(err, ErrUnsafeStorageRoot) {
				t.Fatalf("PrepareStorageRoot(source tree) error = %v, want ErrUnsafeStorageRoot", err)
			}
		})
	}
}
