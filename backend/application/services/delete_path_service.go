package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// DeletePathService removes a file or directory from the caller's namespace.
type DeletePathService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
}

func NewDeletePathService(fileRepo ports.FileRepository, scoper ports.PathScoper) *DeletePathService {
	return &DeletePathService{fileRepo: fileRepo, scoper: scoper}
}

func (s *DeletePathService) Execute(user *models.User, path string) error {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return err
	}

	physical, err := s.scoper.WritePath(user, fp.Relative())
	if err != nil {
		return err
	}

	exists, err := s.fileRepo.FileExists(physical)
	if err != nil {
		return err
	}
	if !exists {
		return &errors.NotFoundError{Path: path}
	}

	return s.fileRepo.DeletePath(physical)
}
