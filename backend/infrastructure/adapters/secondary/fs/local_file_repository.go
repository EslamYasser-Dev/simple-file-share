package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/utils"
)

// LocalFileRepository implements domain.FileRepository using local filesystem.
type LocalFileRepository struct {
	rootDir string
}

// NewLocalFileRepository creates a new file repository adapter.
func NewLocalFileRepository(rootDir string) *LocalFileRepository {
	return &LocalFileRepository{rootDir: rootDir}
}

// resolve converts a relative path to absolute within rootDir.
func (r *LocalFileRepository) resolve(path string) string {
	cleaned := filepath.FromSlash(strings.TrimPrefix(path, "/"))
	return filepath.Join(r.rootDir, cleaned)
}

// ListDirectory returns metadata for all entries in a directory.
func (r *LocalFileRepository) ListDirectory(path string) ([]*models.FileInfo, error) {
	fullPath := r.resolve(path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var files []*models.FileInfo
	for _, entry := range entries {
		name := entry.Name()
		url := "/" + filepath.ToSlash(filepath.Join(strings.TrimPrefix(path, "/"), name))
		zipURL := url + ".zip"
		if path == "/" || path == "" {
			url = "/" + name
			zipURL = "/" + name + ".zip"
		}

		fileInfo, _ := entry.Info()
		size := formatFileSize(fileInfo.Size(), entry.IsDir())

		files = append(files, &models.FileInfo{
			Name:   name,
			URL:    url,
			ZipURL: zipURL,
			Size:   size,
			IsDir:  entry.IsDir(),
		})
	}
	return files, nil
}

// IsDirectory checks if the path is a directory.
func (r *LocalFileRepository) IsDirectory(path string) (bool, error) {
	info, err := os.Stat(r.resolve(path))
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// FileExists checks if a file or directory exists at the given path.
func (r *LocalFileRepository) FileExists(path string) (bool, error) {
	_, err := os.Stat(r.resolve(path))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// ServeFile opens a file for reading and returns its stream and name.
func (r *LocalFileRepository) ServeFile(path string) (models.ReadCloser, string, error) {
	fullPath := r.resolve(path)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, "", err
	}
	return file, filepath.Base(fullPath), nil
}

// CreateDirectory creates all directories in the given path.
func (r *LocalFileRepository) CreateDirectory(path string) error {
	return os.MkdirAll(r.resolve(path), 0755)
}

// WriteFile writes content from reader to the specified file path.
func (r *LocalFileRepository) WriteFile(path string, reader models.ReadCloser) (int64, error) {
	defer reader.Close()

	fullPath := r.resolve(path)
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

// ZipDirectory returns a streaming ZIP archive of the directory.
func (r *LocalFileRepository) ZipDirectory(root string) (models.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		if err := utils.ZipDirectory(r.resolve(root), pw); err != nil {
			pw.CloseWithError(err)
		}
	}()

	return pr, nil
}

// formatFileSize returns human-readable size string.
func formatFileSize(size int64, isDir bool) string {
	if isDir {
		return "[Directory]"
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}
