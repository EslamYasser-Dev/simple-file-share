package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// RebuildIndexService rebuilds the search index from the current storage state.
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
