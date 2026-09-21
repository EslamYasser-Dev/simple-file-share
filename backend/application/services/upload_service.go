package services

import (
	"path/filepath"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// UploadService stores one or more uploaded parts into the caller's namespace.
type UploadService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
}

func NewUploadService(fileRepo ports.FileRepository, scoper ports.PathScoper) *UploadService {
	return &UploadService{fileRepo: fileRepo, scoper: scoper}
}

func (s *UploadService) Execute(user *models.User, parts []models.UploadPart) ([]models.FileUpload, error) {
	var uploads []models.FileUpload
	var execErrors []error

	for _, part := range parts {
		filename := part.Filename()
		content := part.Content()

		if filename == "" {
			content.Close()
			continue
		}

		if _, err := valueobjects.NewFilePath(filename); err != nil {
			content.Close()
			execErrors = append(execErrors, err)
			continue
		}

		physical, err := s.scoper.WritePath(user, filepath.ToSlash(filename))
		if err != nil {
			content.Close()
			execErrors = append(execErrors, err)
			continue
		}

		dir := filepath.Dir(physical)
		if dir != "." && dir != "/" {
			if err := s.fileRepo.CreateDirectory(dir); err != nil {
				content.Close()
				execErrors = append(execErrors, err)
				continue
			}
		}

		written, err := s.fileRepo.WriteFile(physical, content)
		content.Close()
		if err != nil {
			execErrors = append(execErrors, err)
			continue
		}

		uploads = append(uploads, models.FileUpload{
			Filename: s.scoper.PhysicalToVirtual(user, physical),
			Size:     written,
		})
	}

	if len(execErrors) > 0 && len(uploads) == 0 {
		return nil, execErrors[0]
	}

	return uploads, nil
}
