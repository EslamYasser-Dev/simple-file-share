package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// RebuildIndexService rebuilds the metadata search index from the current
// storage state, and optionally the full-text content index.
type RebuildIndexService struct {
	index ports.FileIndexRepository
	text  ports.TextIndex
}

func NewRebuildIndexService(index ports.FileIndexRepository, text ports.TextIndex) *RebuildIndexService {
	return &RebuildIndexService{index: index, text: text}
}

// Execute rebuilds metadata from walk. When textIndexWalk is non-nil it also
// rebuilds the full-text index (startup path).
func (s *RebuildIndexService) Execute(
	rootDir string,
	walk func(string) ([]*models.FileInfo, error),
	textWalk func(string) ([]ports.TextDocument, error),
) error {
	entries, err := walk(rootDir)
	if err != nil {
		return err
	}
	if err := s.index.Rebuild(entries); err != nil {
		return err
	}
	if s.text != nil && textWalk != nil {
		docs, err := textWalk(rootDir)
		if err != nil {
			return err
		}
		return s.text.Rebuild(docs)
	}
	return nil
}
