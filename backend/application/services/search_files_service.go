package services

import (
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type SearchFilesService struct {
	index ports.FileIndexRepository
}

func NewSearchFilesService(index ports.FileIndexRepository) *SearchFilesService {
	return &SearchFilesService{index: index}
}

func (s *SearchFilesService) Execute(query string, limit int) ([]*models.FileInfo, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.NewValidationError("query", query, "query cannot be empty")
	}
	return s.index.Search(query, limit)
}

type RebuildIndexService struct {
	index ports.FileIndexRepository
}

func NewRebuildIndexService(index ports.FileIndexRepository) *RebuildIndexService {
	return &RebuildIndexService{index: index}
}

func (s *RebuildIndexService) Execute(rootDir string, walk func(string) ([]*models.FileInfo, error)) error {
	entries, err := walk(rootDir)
	if err != nil {
		return err
	}
	return s.index.Rebuild(entries)
}
