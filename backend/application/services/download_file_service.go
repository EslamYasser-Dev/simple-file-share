package services

import (
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

type DownloadFileService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
}

func NewDownloadFileService(fileRepo ports.FileRepository, scoper ports.PathScoper) *DownloadFileService {
	return &DownloadFileService{fileRepo: fileRepo, scoper: scoper}
}

func (s *DownloadFileService) Execute(user *models.User, path string) (*models.Download, error) {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, err
	}

	physical, err := s.scoper.ReadPath(user, fp.Relative())
	if err != nil {
		return nil, err
	}

	exists, err := s.fileRepo.FileExists(physical)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &domainerrors.NotFoundError{Path: path}
	}

	isDir, err := s.fileRepo.IsDirectory(physical)
	if err != nil {
		return nil, err
	}
	if isDir {
		return nil, &domainerrors.IsDirectoryError{Path: path}
	}

	stream, filename, err := s.fileRepo.ServeFile(physical)
	if err != nil {
		return nil, err
	}
	return &models.Download{
		Stream:      stream,
		Filename:    filename,
		ContentType: "application/octet-stream",
	}, nil
}
