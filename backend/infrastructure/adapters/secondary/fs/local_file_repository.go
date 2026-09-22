package fs

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/utils"
)

// LocalFileRepository implements ports.FileRepository using the local filesystem.
type LocalFileRepository struct {
	rootDir string
}

func NewLocalFileRepository(rootDir string) *LocalFileRepository {
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		abs = rootDir
	}
	return &LocalFileRepository{rootDir: abs}
}

func (r *LocalFileRepository) resolve(path string) (string, error) {
	cleaned := filepath.FromSlash(strings.TrimPrefix(path, "/"))

	// Defense in depth: reject any exact ".." segment before joining.
	for _, segment := range strings.Split(filepath.ToSlash(cleaned), "/") {
		if segment == ".." {
			return "", errors.NewValidationError("path", path, "path traversal detected")
		}
	}

	fullPath := filepath.Join(r.rootDir, cleaned)

	// Ensure the joined path stays within the root directory.
	if fullPath != r.rootDir &&
		!strings.HasPrefix(fullPath, r.rootDir+string(os.PathSeparator)) {
		return "", errors.NewValidationError("path", path, "path escapes root directory")
	}

	return fullPath, nil
}

func (r *LocalFileRepository) ListDirectory(path string) ([]*models.FileInfo, error) {
	fullPath, err := r.resolve(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	parentPath := "/" + filepath.ToSlash(strings.TrimPrefix(path, "/"))
	if parentPath == "/." {
		parentPath = "/"
	}

	var files []*models.FileInfo
	for _, entry := range entries {
		if shouldSkipIndexPath(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}

		childPath := parentPath
		if childPath == "/" {
			childPath = "/" + entry.Name()
		} else {
			childPath = parentPath + "/" + entry.Name()
		}

		resolved := filepath.Join(fullPath, entry.Name())
		var version uint64
		if !entry.IsDir() {
			version = versionCount(resolved)
		}

		files = append(files, &models.FileInfo{
			Name:     entry.Name(),
			Path:     strings.TrimPrefix(childPath, "/"),
			Size:     info.Size(),
			IsDir:    entry.IsDir(),
			Modified: info.ModTime(),
			Version:  version,
		})
	}
	return files, nil
}

func (r *LocalFileRepository) GetFileInfo(path string) (*models.FileInfo, error) {
	fullPath, err := r.resolve(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	displayPath := "/" + filepath.ToSlash(strings.TrimPrefix(path, "/"))
	if displayPath == "/." {
		displayPath = "/"
	}

	return &models.FileInfo{
		Name:     info.Name(),
		Path:     strings.TrimPrefix(displayPath, "/"),
		Size:     info.Size(),
		IsDir:    info.IsDir(),
		Modified: info.ModTime(),
		Version:  versionCount(fullPath),
	}, nil
}

// versionCount returns how many historical versions exist for a canonical
// file path (stored as the monotonically numbered contents of `<file>.versions/`).
func versionCount(fullPath string) uint64 {
	entries, err := os.ReadDir(fullPath + ".versions")
	if err != nil {
		return 0
	}
	return uint64(len(entries))
}

// nextVersionNumber returns the first free slot in the version directory.
func nextVersionNumber(fullPath string) int {
	versionDir := fullPath + ".versions"
	entries, err := os.ReadDir(versionDir)
	if err != nil {
		return 1
	}
	max := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if ver, err := strconv.Atoi(entry.Name()); err == nil && ver > max {
			max = ver
		}
	}
	return max + 1
}

func (r *LocalFileRepository) IsDirectory(path string) (bool, error) {
	fullPath, err := r.resolve(path)
	if err != nil {
		return false, err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func (r *LocalFileRepository) FileExists(path string) (bool, error) {
	fullPath, err := r.resolve(path)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (r *LocalFileRepository) ServeFile(path string) (io.ReadCloser, string, error) {
	fullPath, err := r.resolve(path)
	if err != nil {
		return nil, "", err
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, "", err
	}
	return file, filepath.Base(fullPath), nil
}

func (r *LocalFileRepository) CreateDirectory(path string) error {
	fullPath, err := r.resolve(path)
	if err != nil {
		return err
	}
	return os.MkdirAll(fullPath, storageDirPerm)
}

func (r *LocalFileRepository) DeletePath(path string) error {
	fullPath, err := r.resolve(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return os.RemoveAll(fullPath)
	}
	// Deleting a file also removes its version history so no orphaned
	// snapshots linger on disk.
	_ = os.RemoveAll(fullPath + ".versions")
	return os.Remove(fullPath)
}

func (r *LocalFileRepository) WriteFile(path string, reader io.ReadCloser) (int64, error) {
	defer func() { _ = reader.Close() }()

	fullPath, err := r.resolve(path)
	if err != nil {
		return 0, err
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), storageDirPerm); err != nil {
		return 0, err
	}

	// If the file already exists, snapshot it into `.versions/` so overwrites
	// keep a recoverable history. The version directory is a sibling hidden
	// path (`<file>.versions/`) containing monotonically numbered copies.
	if _, statErr := os.Stat(fullPath); statErr == nil {
		versionDir := fullPath + ".versions"
		if err := os.MkdirAll(versionDir, storageDirPerm); err != nil {
			return 0, err
		}
		verPath := filepath.Join(versionDir, strconv.Itoa(nextVersionNumber(fullPath)))
		if err := os.Rename(fullPath, verPath); err != nil {
			return 0, err
		}
	}

	// Create with owner-only permissions so only the server can read uploads.
	dst, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return 0, err
	}

	written, copyErr := io.Copy(dst, reader)
	closeErr := dst.Close()
	if copyErr != nil {
		// Remove the partial file and restore the previous version.
		_ = os.Remove(fullPath)
		_ = restoreLatestVersion(fullPath)
		return written, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(fullPath)
		_ = restoreLatestVersion(fullPath)
		return written, closeErr
	}
	return written, nil
}

// restoreLatestVersion moves the newest numbered copy back onto the canonical
// path. Used to roll saved history back when a fresh write fails.
func restoreLatestVersion(fullPath string) error {
	versionDir := fullPath + ".versions"
	entries, err := os.ReadDir(versionDir)
	if err != nil {
		return nil
	}
	maxVer := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if ver, err := strconv.Atoi(entry.Name()); err == nil && ver > maxVer {
			maxVer = ver
		}
	}
	if maxVer == 0 {
		return nil
	}
	return os.Rename(filepath.Join(versionDir, strconv.Itoa(maxVer)), fullPath)
}

func (r *LocalFileRepository) ZipDirectory(root string) (io.ReadCloser, error) {
	fullPath, err := r.resolve(root)
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		if err := utils.ZipDirectory(fullPath, pw); err != nil {
			pw.CloseWithError(err)
		}
	}()

	return pr, nil
}
