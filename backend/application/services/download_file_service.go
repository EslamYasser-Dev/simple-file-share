package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

type DownloadFileService struct {
	fileRepo ports.FileRepository
}

func NewDownloadFileService(fileRepo ports.FileRepository) *DownloadFileService {
	return &DownloadFileService{fileRepo: fileRepo}
}

func (s *DownloadFileService) Execute(path string) (models.ReadCloser, string, error) {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, "", err
	}

	rel := fp.Relative()
	exists, err := s.fileRepo.FileExists(rel)
	if err != nil {
		return nil, "", err
	}
	if !exists {
		return nil, "", &errors.NotFoundError{Path: path}
	}

	isDir, err := s.fileRepo.IsDirectory(rel)
	if err != nil {
		return nil, "", err
	}
	if isDir {
		return nil, "", nil
	}

	return s.fileRepo.ServeFile(rel)
}
