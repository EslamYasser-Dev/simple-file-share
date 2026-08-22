package fs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// IndexedFileRepository wraps filesystem storage and keeps a search index in sync.
type IndexedFileRepository struct {
	fs    ports.FileRepository
	index ports.FileIndexRepository
}

func NewIndexedFileRepository(fs ports.FileRepository, index ports.FileIndexRepository) *IndexedFileRepository {
	return &IndexedFileRepository{fs: fs, index: index}
}

func (r *IndexedFileRepository) ListDirectory(path string) ([]*models.FileInfo, error) {
	return r.fs.ListDirectory(path)
}

func (r *IndexedFileRepository) GetFileInfo(path string) (*models.FileInfo, error) {
	return r.fs.GetFileInfo(path)
}

func (r *IndexedFileRepository) IsDirectory(path string) (bool, error) {
	return r.fs.IsDirectory(path)
}

func (r *IndexedFileRepository) FileExists(path string) (bool, error) {
	return r.fs.FileExists(path)
}

func (r *IndexedFileRepository) ServeFile(path string) (models.ReadCloser, string, error) {
	return r.fs.ServeFile(path)
}

func (r *IndexedFileRepository) CreateDirectory(path string) error {
	if err := r.fs.CreateDirectory(path); err != nil {
		return err
	}
	r.syncPath(path)
	return nil
}

func (r *IndexedFileRepository) DeletePath(path string) error {
	isDir, err := r.fs.IsDirectory(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := r.fs.DeletePath(path); err != nil {
		return err
	}

	normalized := strings.TrimPrefix(filepath.ToSlash(path), "/")
	if isDir {
		_ = r.index.RemovePrefix(normalized)
	} else {
		_ = r.index.Remove(normalized)
	}
	return nil
}

func (r *IndexedFileRepository) WriteFile(path string, reader models.ReadCloser) (int64, error) {
	written, err := r.fs.WriteFile(path, reader)
	if err != nil {
		return written, err
	}
	r.syncPath(path)
	return written, nil
}

func (r *IndexedFileRepository) ZipDirectory(root string) (models.ReadCloser, error) {
	return r.fs.ZipDirectory(root)
}

func (r *IndexedFileRepository) syncPath(path string) {
	info, err := r.fs.GetFileInfo(path)
	if err != nil {
		return
	}
	_ = r.index.Upsert(info)
}

// WalkRoot indexes every file and directory under rootDir.
func WalkRoot(rootDir string) ([]*models.FileInfo, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	var entries []*models.FileInfo
	err = filepath.WalkDir(absRoot, func(fullPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(absRoot, fullPath)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		rel = filepath.ToSlash(rel)
		if shouldSkipIndexPath(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		entries = append(entries, &models.FileInfo{
			Name:     d.Name(),
			Path:     rel,
			Size:     info.Size(),
			IsDir:    d.IsDir(),
			Modified: info.ModTime(),
		})
		return nil
	})
	return entries, err
}

func shouldSkipIndexPath(path string) bool {
	if strings.HasPrefix(path, ".file-share") {
		return true
	}
	base := filepath.Base(path)
	return strings.HasPrefix(base, ".")
}

var _ ports.FileRepository = (*IndexedFileRepository)(nil)

// Ensure LocalFileRepository satisfies FileRepository when passed as fs.
var _ ports.FileRepository = (*LocalFileRepository)(nil)
