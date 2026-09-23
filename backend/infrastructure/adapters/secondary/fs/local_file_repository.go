package fs

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/utils"
)

// LocalFileRepository implements ports.FileRepository using the local filesystem.
type LocalFileRepository struct {
	rootDir     string
	versionKeep int
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
	r.pruneVersions(fullPath)
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

func (r *LocalFileRepository) ListVersions(path string) ([]*models.FileInfo, error) {
	fullPath, err := r.resolve(path)
	if err != nil {
		return nil, err
	}
	if info, statErr := os.Stat(fullPath); statErr != nil {
		return nil, errors.ErrNotFound
	} else if info.IsDir() {
		return nil, nil
	}

	nums, err := versionNumbers(fullPath)
	if err != nil {
		return nil, err
	}

	versionDir := fullPath + ".versions"
	versions := make([]*models.FileInfo, 0, len(nums))
	for _, n := range nums {
		verPath := filepath.Join(versionDir, strconv.Itoa(n))
		fi, statErr := os.Stat(verPath)
		if statErr != nil {
			continue
		}
		versions = append(versions, &models.FileInfo{
			Name:     filepath.Base(fullPath),
			Path:     path,
			Size:     fi.Size(),
			IsDir:    false,
			Modified: fi.ModTime(),
			Version:  uint64(n),
		})
	}
	return versions, nil
}

func (r *LocalFileRepository) ServeVersion(path string, n int) (io.ReadCloser, string, error) {
	if n < 1 {
		return nil, "", errors.ErrNotFound
	}
	fullPath, err := r.resolve(path)
	if err != nil {
		return nil, "", err
	}
	verPath := filepath.Join(fullPath+".versions", strconv.Itoa(n))
	file, err := os.Open(verPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", errors.ErrNotFound
		}
		return nil, "", err
	}
	return file, filepath.Base(fullPath), nil
}

func (r *LocalFileRepository) RestoreVersion(path string, n int) error {
	if n < 1 {
		return errors.ErrNotFound
	}
	fullPath, err := r.resolve(path)
	if err != nil {
		return err
	}
	verPath := filepath.Join(fullPath+".versions", strconv.Itoa(n))
	src, err := os.Open(verPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.ErrNotFound
		}
		return err
	}
	defer src.Close()

	versionDir := fullPath + ".versions"
	if _, statErr := os.Stat(fullPath); statErr == nil {
		if err := os.MkdirAll(versionDir, storageDirPerm); err != nil {
			return err
		}
		snapPath := filepath.Join(versionDir, strconv.Itoa(nextVersionNumber(fullPath)))
		if err := os.Rename(fullPath, snapPath); err != nil {
			return err
		}
	}

	dst, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		_ = restoreLatestVersion(fullPath)
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		_ = os.Remove(fullPath)
		_ = restoreLatestVersion(fullPath)
		return err
	}
	if err := dst.Close(); err != nil {
		_ = os.Remove(fullPath)
		_ = restoreLatestVersion(fullPath)
		return err
	}
	r.pruneVersions(fullPath)
	return nil
}

// SetVersionKeep bounds how many historical snapshots are retained per file.
// A value of 0 (the default) keeps every snapshot.
func (r *LocalFileRepository) SetVersionKeep(n int) {
	if n < 0 {
		n = 0
	}
	r.versionKeep = n
}

func (r *LocalFileRepository) pruneVersions(fullPath string) {
	if r.versionKeep <= 0 {
		return
	}
	nums, err := versionNumbers(fullPath)
	if err != nil || len(nums) <= r.versionKeep {
		return
	}
	versionDir := fullPath + ".versions"
	for _, n := range nums[:len(nums)-r.versionKeep] {
		_ = os.Remove(filepath.Join(versionDir, strconv.Itoa(n)))
	}
}

func versionNumbers(fullPath string) ([]int, error) {
	entries, err := os.ReadDir(fullPath + ".versions")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var nums []int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if n, convErr := strconv.Atoi(entry.Name()); convErr == nil && n > 0 {
			nums = append(nums, n)
		}
	}
	sort.Ints(nums)
	return nums, nil
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
