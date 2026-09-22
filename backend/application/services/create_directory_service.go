package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// CreateDirectoryService creates a directory inside the caller's namespace.
type CreateDirectoryService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
}

func NewCreateDirectoryService(fileRepo ports.FileRepository, scoper ports.PathScoper) *CreateDirectoryService {
	return &CreateDirectoryService{fileRepo: fileRepo, scoper: scoper}
}

func (s *CreateDirectoryService) Execute(user *models.User, path string) error {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return err
	}

	physical, err := s.scoper.WritePath(user, fp.Relative())
	if err != nil {
		return err
	}
	return s.fileRepo.CreateDirectory(physical)
}
