package services

import (
	"io"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// UpdateFileContentService overwrites a file's contents with the supplied text
// (used by the markdown editor). Like uploads, it keeps the search index in
// sync through the port.
type UpdateFileContentService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
}

func NewUpdateFileContentService(fileRepo ports.FileRepository, scoper ports.PathScoper) *UpdateFileContentService {
	return &UpdateFileContentService{fileRepo: fileRepo, scoper: scoper}
}

func (s *UpdateFileContentService) Execute(user *models.User, path, content string) (int64, error) {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return 0, err
	}

	physical, err := s.scoper.WritePath(user, fp.Relative())
	if err != nil {
		return 0, err
	}

	exists, err := s.fileRepo.FileExists(physical)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, &errors.NotFoundError{Path: path}
	}

	isDir, err := s.fileRepo.IsDirectory(physical)
	if err != nil {
		return 0, err
	}
	if isDir {
		return 0, errors.NewValidationError("path", path, "cannot edit a directory")
	}

	reader := io.NopCloser(strings.NewReader(content))
	return s.fileRepo.WriteFile(physical, reader)
}
