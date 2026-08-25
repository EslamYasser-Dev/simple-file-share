package ports

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// FileIndexRepository indexes file metadata for fast search.
type FileIndexRepository interface {
	Search(query string, limit int) ([]*models.FileInfo, error)
	Upsert(info *models.FileInfo) error
	Remove(path string) error
	RemovePrefix(pathPrefix string) error
	Rebuild(entries []*models.FileInfo) error
	Clear() error
}
