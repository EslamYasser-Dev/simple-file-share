package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

type ListFilesService struct {
	fileRepo ports.FileRepository
}

func NewListFilesService(fileRepo ports.FileRepository) *ListFilesService {
	return &ListFilesService{fileRepo: fileRepo}
}

func (s *ListFilesService) Execute(path string) (*models.PageData, error) {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, err
	}

	rel := fp.Relative()
	isDir, err := s.fileRepo.IsDirectory(rel)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, nil
	}

	files, err := s.fileRepo.ListDirectory(rel)
	if err != nil {
		return nil, err
	}

	return &models.PageData{Root: fp.String(), Files: files}, nil
}
