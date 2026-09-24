package services

import (
	"errors"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

// DownloadService resolves a download request into a stream, deciding whether
// to serve a single file or a zipped directory from the requested path.
type DownloadService struct {
	files *DownloadFileService
	zips  *DownloadZipService
	bus   *events.Bus
}

func NewDownloadService(files *DownloadFileService, zips *DownloadZipService) *DownloadService {
	return &DownloadService{files: files, zips: zips}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *DownloadService) SetEventBus(bus *events.Bus) { s.bus = bus }

// Execute streams a directory as a ZIP archive, or a regular file as-is. The
// choice is made from the target's actual type — never from a ".zip" suffix,
// which would make real `.zip` files impossible to download.
func (s *DownloadService) Execute(user *models.User, path string) (*models.Download, error) {
	download, err := s.zips.Execute(user, path)
	if err == nil {
		publishEvent(s.bus, events.TypeDownload, path, user)
		return download, nil
	}

	var notDir *domainerrors.NotDirectoryError
	if !errors.As(err, &notDir) {
		return nil, err
	}
	download, err = s.files.Execute(user, path)
	if err != nil {
		return nil, err
	}
	publishEvent(s.bus, events.TypeDownload, path, user)
	return download, nil
}
