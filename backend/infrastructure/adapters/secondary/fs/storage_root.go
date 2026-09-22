package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const storageDirPerm os.FileMode = 0o700

// ErrUnsafeStorageRoot is returned when the configured storage root is a
// dangerous location: a filesystem root, the working directory, the user's home
// directory, a source checkout, or a symlink.
var ErrUnsafeStorageRoot = errors.New("unsafe storage root")

// sourceMarkers identify a source checkout. Storage must never point here
// because the server deletes and rewrites everything under its root.
var sourceMarkers = []string{"go.mod", "package.json", ".git"}

// PrepareStorageRoot validates rootDir, creates it when missing, and forces it
// to owner-only permissions (0700). It returns the cleaned absolute path.
//
// The storage root is the single directory the server owns and exclusively
// reads/writes, so it must be a dedicated location and not shared with the
// operating system, the source tree, or other applications.
func PrepareStorageRoot(rootDir string) (string, error) {
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		return "", fmt.Errorf("resolve storage root: %w", err)
	}
	abs = filepath.Clean(abs)

	if err := checkStorageRootSafe(abs); err != nil {
		return "", err
	}

	if err := os.MkdirAll(abs, storageDirPerm); err != nil {
		return "", fmt.Errorf("create storage root %s: %w", abs, err)
	}

	// Tighten the root to owner-only. We may not own the directory (for example
	// a mounted volume): if it is already inaccessible to group/other, accept it
	// as-is; otherwise the operator must fix the ownership and we fail loudly.
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("inspect storage root %s: %w", abs, err)
	}
	if info.Mode().Perm() != storageDirPerm {
		if err := os.Chmod(abs, storageDirPerm); err != nil && info.Mode().Perm()&0o077 != 0 {
			return "", fmt.Errorf("secure storage root %s: %w", abs, err)
		}
	}
	return abs, nil
}

func checkStorageRootSafe(abs string) error {
	// Refuse symlinks: following one could silently redirect storage elsewhere.
	if info, err := os.Lstat(abs); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: %s is a symlink", ErrUnsafeStorageRoot, abs)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect storage root %s: %w", abs, err)
	}

	// Refuse the filesystem/volume root (e.g. "/" or "C:\\").
	if filepath.Dir(abs) == abs {
		return fmt.Errorf("%w: %s is a filesystem root", ErrUnsafeStorageRoot, abs)
	}

	if home, err := os.UserHomeDir(); err == nil && samePath(abs, home) {
		return fmt.Errorf("%w: %s is the home directory", ErrUnsafeStorageRoot, abs)
	}

	if cwd, err := os.Getwd(); err == nil && samePath(abs, cwd) {
		return fmt.Errorf("%w: %s is the working directory", ErrUnsafeStorageRoot, abs)
	}

	for _, marker := range sourceMarkers {
		if _, err := os.Stat(filepath.Join(abs, marker)); err == nil {
			return fmt.Errorf("%w: %s contains %s (looks like a source tree)", ErrUnsafeStorageRoot, abs, marker)
		}
	}
	return nil
}

func samePath(a, b string) bool {
	aa, err := filepath.Abs(a)
	if err != nil {
		return false
	}
	bb, err := filepath.Abs(b)
	if err != nil {
		return false
	}
	return filepath.Clean(aa) == filepath.Clean(bb)
}
