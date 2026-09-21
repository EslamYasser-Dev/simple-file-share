package services

import (
	"strings"

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

func (s *DownloadService) Execute(user *models.User, path string) (*models.Download, error) {
	if strings.HasSuffix(path, ".zip") {
		return s.zips.Execute(user, strings.TrimSuffix(path, ".zip"))
	}
	return s.files.Execute(user, path)
}
