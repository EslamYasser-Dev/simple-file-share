package services

import (
	"path/filepath"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

type DownloadZipService struct {
	fileRepo ports.FileRepository
}

func NewDownloadZipService(fileRepo ports.FileRepository) *DownloadZipService {
	return &DownloadZipService{fileRepo: fileRepo}
}

func (s *DownloadZipService) Execute(path string) (models.ReadCloser, string, error) {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, "", err
	}

	rel := fp.Relative()
	isDir, err := s.fileRepo.IsDirectory(rel)
	if err != nil {
		return nil, "", err
	}
	if !isDir {
		return nil, "", nil
	}

	zipStream, err := s.fileRepo.ZipDirectory(rel)
	if err != nil {
		return nil, "", err
	}

	name := filepath.Base(rel)
	if name == "." || name == "" {
		name = "root"
	}
	return zipStream, name + ".zip", nil
}
