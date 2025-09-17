package ports

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

type FileRepository interface {
	ListDirectory(path string) ([]*models.FileInfo, error)
	IsDirectory(path string) (bool, error)
	FileExists(path string) (bool, error)
	ServeFile(path string) (models.ReadCloser, string, error)
	CreateDirectory(path string) error
	WriteFile(path string, reader models.ReadCloser) (int64, error)
	ZipDirectory(root string) (models.ReadCloser, error)
}
