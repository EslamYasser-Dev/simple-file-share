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

	for _, part := range parts {
		filename := part.Filename()
		if filename == "" {
			part.Content().Close()
			continue
		}

		dir := filepath.Dir(filename)
		if err := s.fileRepo.CreateDirectory(dir); err != nil {
			part.Content().Close()
			continue
		}

		written, err := s.fileRepo.WriteFile(filename, part.Content())
		if err != nil {
			part.Content().Close()
			continue
		}

		uploads = append(uploads, models.FileUpload{
			Filename: filename,
			Size:     written,
		})
	}

	return uploads, nil
}
