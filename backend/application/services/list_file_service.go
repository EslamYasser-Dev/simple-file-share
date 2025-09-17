package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type ListFilesService struct {
	fileRepo ports.FileRepository
}

func NewListFilesService(fileRepo ports.FileRepository) *ListFilesService {
	return &ListFilesService{fileRepo: fileRepo}
}

func (s *ListFilesService) Execute(path string) (*models.PageData, error) {
	isDir, err := s.fileRepo.IsDirectory(path)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, nil // Not a directory → delegate to download
	}

	files, err := s.fileRepo.ListDirectory(path)
	if err != nil {
		return nil, err
	}

	return &models.PageData{Root: path, Files: files}, nil
}
