package services

import (
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

type ListFilesService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
}

func NewListFilesService(fileRepo ports.FileRepository, scoper ports.PathScoper) *ListFilesService {
	return &ListFilesService{fileRepo: fileRepo, scoper: scoper}
}

func (s *ListFilesService) Execute(user *models.User, path string) (*models.PageData, error) {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, err
	}

	physical, err := s.scoper.ReadPath(user, fp.Relative())
	if err != nil {
		return nil, err
	}

	isDir, err := s.fileRepo.IsDirectory(physical)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, &domainerrors.NotDirectoryError{Path: path}
	}

	files, err := s.fileRepo.ListDirectory(physical)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		f.Path = s.scoper.PhysicalToVirtual(user, f.Path)
	}

	return &models.PageData{Files: files}, nil
}
