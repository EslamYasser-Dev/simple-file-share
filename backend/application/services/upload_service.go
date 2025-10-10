package services

import (
	"path/filepath"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type UploadService struct {
	fileRepo ports.FileRepository
}

func NewUploadService(fileRepo ports.FileRepository) *UploadService {
	return &UploadService{fileRepo: fileRepo}
}

func (s *UploadService) Execute(parts []models.UploadPart) ([]models.FileUpload, error) {
	var uploads []models.FileUpload
	var errors []error

	for _, part := range parts {
		filename := part.Filename()
		if filename == "" {
			part.Content().Close()
			continue
		}

		// Ensure proper resource cleanup
		content := part.Content()
		defer content.Close()

		dir := filepath.Dir(filename)
		if err := s.fileRepo.CreateDirectory(dir); err != nil {
			errors = append(errors, err)
			continue
		}

		written, err := s.fileRepo.WriteFile(filename, content)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		uploads = append(uploads, models.FileUpload{
			Filename: filename,
			Size:     written,
		})
	}

	// Return partial success if some files uploaded successfully
	if len(errors) > 0 && len(uploads) == 0 {
		return nil, errors[0] // Return first error if no successful uploads
	}

	return uploads, nil
}
