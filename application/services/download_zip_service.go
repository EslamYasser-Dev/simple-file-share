package services

import (
	"path/filepath"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type DownloadZipService struct {
	fileRepo ports.FileRepository
}

func NewDownloadZipService(fileRepo ports.FileRepository) *DownloadZipService {
	return &DownloadZipService{fileRepo: fileRepo}
}

func (s *DownloadZipService) Execute(path string) (models.ReadCloser, string, error) {
	isDir, err := s.fileRepo.IsDirectory(path)
	if err != nil {
		return nil, "", err
	}
	if !isDir {
		return nil, "", nil // Not a dir → not zip
	}

	zipStream, err := s.fileRepo.ZipDirectory(path)
	if err != nil {
		return nil, "", err
	}

	return zipStream, filepath.Base(path) + ".zip", nil
}
