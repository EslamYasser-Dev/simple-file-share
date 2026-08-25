package services

import (
	"os"
	"path/filepath"

	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

type UploadService struct {
	fileRepo ports.FileRepository
}

func NewUploadService(fileRepo ports.FileRepository) *UploadService {
	return &UploadService{fileRepo: fileRepo}
}

func (s *UploadService) Execute(parts []models.UploadPart) ([]models.FileUpload, error) {
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

		dir := filepath.Dir(filename)
		if dir != "." && dir != "/" {
			if err := s.fileRepo.CreateDirectory(dir); err != nil {
				content.Close()
				execErrors = append(execErrors, err)
				continue
			}
		}

		written, err := s.fileRepo.WriteFile(filename, content)
		content.Close()
		if err != nil {
			execErrors = append(execErrors, err)
			continue
		}

		uploads = append(uploads, models.FileUpload{
			Filename: filename,
			Size:     written,
		})
	}

	if len(execErrors) > 0 && len(uploads) == 0 {
		return nil, execErrors[0]
	}

	return uploads, nil
}

type CreateDirectoryService struct {
	fileRepo ports.FileRepository
}

func NewCreateDirectoryService(fileRepo ports.FileRepository) *CreateDirectoryService {
	return &CreateDirectoryService{fileRepo: fileRepo}
}

func (s *CreateDirectoryService) Execute(path string) error {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return err
	}
	return s.fileRepo.CreateDirectory(fp.Relative())
}

type DeletePathService struct {
	fileRepo ports.FileRepository
}

func NewDeletePathService(fileRepo ports.FileRepository) *DeletePathService {
	return &DeletePathService{fileRepo: fileRepo}
}

func (s *DeletePathService) Execute(path string) error {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return err
	}

	exists, err := s.fileRepo.FileExists(fp.Relative())
	if err != nil {
		return err
	}
	if !exists {
		return &errors.NotFoundError{Path: path}
	}

	return s.fileRepo.DeletePath(fp.Relative())
}

type GetFileInfoService struct {
	fileRepo ports.FileRepository
}

func NewGetFileInfoService(fileRepo ports.FileRepository) *GetFileInfoService {
	return &GetFileInfoService{fileRepo: fileRepo}
}

func (s *GetFileInfoService) Execute(path string) (*models.FileInfo, error) {
	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, err
	}

	info, err := s.fileRepo.GetFileInfo(fp.Relative())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &errors.NotFoundError{Path: path}
		}
		return nil, err
	}
	return info, nil
}
