package services

import (
	"errors"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// GetFileInfoService returns metadata for a single file or directory.
type GetFileInfoService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
}

func NewGetFileInfoService(fileRepo ports.FileRepository, scoper ports.PathScoper) *GetFileInfoService {
	return &GetFileInfoService{fileRepo: fileRepo, scoper: scoper}
}

func (s *GetFileInfoService) Execute(user *models.User, path string) (*models.FileInfo, error) {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, err
	}

	physical, err := s.scoper.ReadPath(user, fp.Relative())
	if err != nil {
		return nil, err
	}

	info, err := s.fileRepo.GetFileInfo(physical)
	if err != nil {
		if errors.Is(err, domainerrors.ErrNotFound) {
			return nil, &domainerrors.NotFoundError{Path: path}
		}
		return nil, err
	}
	info.Path = s.scoper.PhysicalToVirtual(user, info.Path)
	return info, nil
}
