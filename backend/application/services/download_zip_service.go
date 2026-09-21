package services

import (
	"io"
	"path/filepath"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

type DownloadZipService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
}

func NewDownloadZipService(fileRepo ports.FileRepository, scoper ports.PathScoper) *DownloadZipService {
	return &DownloadZipService{fileRepo: fileRepo, scoper: scoper}
}

func (s *DownloadZipService) Execute(user *models.User, path string) (io.ReadCloser, string, error) {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, "", err
	}

	physical, err := s.scoper.ReadPath(user, fp.Relative())
	if err != nil {
		return nil, "", err
	}

	isDir, err := s.fileRepo.IsDirectory(physical)
	if err != nil {
		return nil, "", err
	}
	if !isDir {
		return nil, "", nil
	}

	zipStream, err := s.fileRepo.ZipDirectory(physical)
	if err != nil {
		return nil, "", err
	}

	name := filepath.Base(physical)
	if name == "." || name == "" {
		name = "root"
	}
	return zipStream, name + ".zip", nil
}
