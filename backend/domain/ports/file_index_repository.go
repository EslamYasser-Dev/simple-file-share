package ports

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// FileIndexRepository indexes file metadata for fast search.
type FileIndexRepository interface {
	Search(query string, limit int) ([]*models.FileInfo, error)
	PrefixStats(prefix string) (files int, size int64, err error)
	// Get returns a clone of the entry for path, or nil when absent.
	Get(path string) (*models.FileInfo, error)
	Upsert(info *models.FileInfo) error
	Remove(path string) error
	RemovePrefix(pathPrefix string) error
	Rebuild(entries []*models.FileInfo) error
	Clear() error
}
