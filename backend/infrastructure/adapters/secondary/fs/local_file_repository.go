package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	if strings.Contains(cleaned, "..") {
		return "", errors.NewValidationError("path", path, "path traversal detected")
	}

	fullPath := filepath.Join(r.rootDir, cleaned)

	absRoot, err := filepath.Abs(r.rootDir)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot+string(os.PathSeparator)) {
		return "", errors.NewValidationError("path", path, "path escapes root directory")
	}

	return absPath, nil
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

		files = append(files, &models.FileInfo{
			Name:     entry.Name(),
			Path:     strings.TrimPrefix(childPath, "/"),
			Size:     info.Size(),
			IsDir:    entry.IsDir(),
			Modified: info.ModTime(),
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
		return nil, err
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
	}, nil
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

func (r *LocalFileRepository) ServeFile(path string) (models.ReadCloser, string, error) {
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
	return os.MkdirAll(fullPath, 0755)
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
	return os.Remove(fullPath)
}

func (r *LocalFileRepository) WriteFile(path string, reader models.ReadCloser) (int64, error) {
	defer reader.Close()

	fullPath, err := r.resolve(path)
	if err != nil {
		return 0, err
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return 0, err
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	return io.Copy(dst, reader)
}

func (r *LocalFileRepository) ZipDirectory(root string) (models.ReadCloser, error) {
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
