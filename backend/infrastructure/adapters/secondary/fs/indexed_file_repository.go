package fs

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// IndexedFileRepository wraps filesystem storage and keeps metadata and
// full-text indexes in sync on every mutation.
type IndexedFileRepository struct {
	fs    ports.FileRepository
	index ports.FileIndexRepository
	text  ports.TextIndex
}

// NewIndexedFileRepository wires the metadata index. text may be nil to skip
// content indexing (unit tests, read-only tools).
func NewIndexedFileRepository(fs ports.FileRepository, index ports.FileIndexRepository, text ports.TextIndex) *IndexedFileRepository {
	return &IndexedFileRepository{fs: fs, index: index, text: text}
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

func (r *IndexedFileRepository) ServeFile(path string) (io.ReadCloser, string, error) {
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
		if r.text != nil {
			_ = r.text.RemovePrefix(normalized)
		}
	} else {
		_ = r.index.Remove(normalized)
		if r.text != nil {
			_ = r.text.Remove(normalized)
		}
	}
	return nil
}

func (r *IndexedFileRepository) WriteFile(path string, reader io.ReadCloser) (int64, error) {
	written, err := r.fs.WriteFile(path, reader)
	if err != nil {
		return written, err
	}
	r.syncPath(path)
	return written, nil
}

func (r *IndexedFileRepository) ZipDirectory(root string) (io.ReadCloser, error) {
	return r.fs.ZipDirectory(root)
}

// SyncPath refreshes metadata and (when available) content for path after an
// external write such as a version restore.
func (r *IndexedFileRepository) SyncPath(path string) {
	r.syncPath(path)
}

func (r *IndexedFileRepository) syncPath(path string) {
	info, err := r.fs.GetFileInfo(path)
	if err != nil {
		return
	}
	_ = r.index.Upsert(info)
	r.syncContent(path, info)
}

// syncContent re-extracts text for indexable files and updates the inverted
// index. Directories and binary/non-indexable files drop any prior content.
func (r *IndexedFileRepository) syncContent(path string, info *models.FileInfo) {
	if r.text == nil {
		return
	}
	normalized := strings.TrimPrefix(filepath.ToSlash(path), "/")
	if info == nil || info.IsDir || !policy.IsIndexableTextPath(normalized) {
		_ = r.text.Remove(normalized)
		return
	}
	rc, _, err := r.fs.ServeFile(normalized)
	if err != nil {
		_ = r.text.Remove(normalized)
		return
	}
	defer rc.Close()
	content, ok := policy.ReadIndexableText(normalized, rc)
	if !ok {
		_ = r.text.Remove(normalized)
		return
	}
	_ = r.text.Index(normalized, content)
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
			Version:  versionCount(fullPath),
		})
		return nil
	})
	return entries, err
}

// WalkRootTextDocuments extracts searchable text from indexable files under
// rootDir for a full-text index rebuild at startup.
func WalkRootTextDocuments(rootDir string) ([]ports.TextDocument, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	var docs []ports.TextDocument
	err = filepath.WalkDir(absRoot, func(fullPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, relErr := filepath.Rel(absRoot, fullPath)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if shouldSkipIndexPath(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkipIndexPath(rel) {
			return nil
		}
		if !policy.IsIndexableTextPath(rel) {
			return nil
		}

		f, openErr := os.Open(fullPath)
		if openErr != nil {
			return nil
		}
		content, ok := policy.ReadIndexableText(rel, f)
		_ = f.Close()
		if !ok {
			return nil
		}
		docs = append(docs, ports.TextDocument{Path: rel, Content: content})
		return nil
	})
	return docs, err
}

func shouldSkipIndexPath(path string) bool {
	if strings.HasPrefix(path, ".file-share") {
		return true
	}
	base := filepath.Base(path)
	// Hidden metadata (dotfiles) and version-history siblings of files are
	// never surfaced to users or the search index.
	return strings.HasPrefix(base, ".") || strings.HasSuffix(base, ".versions")
}

var _ ports.FileRepository = (*IndexedFileRepository)(nil)

// Ensure LocalFileRepository satisfies FileRepository when passed as fs.
var _ ports.FileRepository = (*LocalFileRepository)(nil)
var _ ports.VersionRepository = (*LocalFileRepository)(nil)
