package services

import (
	"errors"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

// DownloadService resolves a download request into a stream, deciding whether
// to serve a single file or a zipped directory from the requested path.
type DownloadService struct {
	files *DownloadFileService
	zips  *DownloadZipService
}

func NewDownloadService(files *DownloadFileService, zips *DownloadZipService) *DownloadService {
	return &DownloadService{files: files, zips: zips}
}

// Execute streams a directory as a ZIP archive, or a regular file as-is. The
// choice is made from the target's actual type — never from a ".zip" suffix,
// which would make real `.zip` files impossible to download.
func (s *DownloadService) Execute(user *models.User, path string) (*models.Download, error) {
	download, err := s.zips.Execute(user, path)
	if err == nil {
		return download, nil
	}

	var notDir *domainerrors.NotDirectoryError
	if !errors.As(err, &notDir) {
		return nil, err
	}
	return s.files.Execute(user, path)
}
