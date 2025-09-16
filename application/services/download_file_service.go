package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type DownloadFileService struct {
	fileRepo ports.FileRepository
}

func NewDownloadFileService(fileRepo ports.FileRepository) *DownloadFileService {
	return &DownloadFileService{fileRepo: fileRepo}
}

func (s *DownloadFileService) Execute(path string) (models.ReadCloser, string, error) {
	exists, err := s.fileRepo.FileExists(path)
	if err != nil {
		return nil, "", err
	}
	if !exists {
		return nil, "", &errors.NotFoundError{Path: path}
	}

	isDir, err := s.fileRepo.IsDirectory(path)
	if err != nil {
		return nil, "", err
	}
	if isDir {
		return nil, "", nil // Delegate to list or zip
	}

	return s.fileRepo.ServeFile(path)
}
