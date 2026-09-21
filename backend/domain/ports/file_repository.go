package ports

import (
	"io"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

type FileRepository interface {
	ListDirectory(path string) ([]*models.FileInfo, error)
	GetFileInfo(path string) (*models.FileInfo, error)
	IsDirectory(path string) (bool, error)
	FileExists(path string) (bool, error)
	ServeFile(path string) (io.ReadCloser, string, error)
	CreateDirectory(path string) error
	DeletePath(path string) error
	WriteFile(path string, reader io.ReadCloser) (int64, error)
	ZipDirectory(root string) (io.ReadCloser, error)
}
