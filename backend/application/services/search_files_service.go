package services

import (
	"path/filepath"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type SearchFilesService struct {
	index  ports.FileIndexRepository
	scoper ports.PathScoper
}

// DefaultSearchLimit is applied by the use case when a client does not ask
// for a specific result count, so every adapter speaks the same contract.
const DefaultSearchLimit = 50

func NewSearchFilesService(index ports.FileIndexRepository, scoper ports.PathScoper) *SearchFilesService {
	return &SearchFilesService{index: index, scoper: scoper}
}

func (s *SearchFilesService) Execute(user *models.User, query string, limit int) ([]*models.FileInfo, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.NewValidationError("query", query, "query cannot be empty")
	}
	if limit <= 0 {
		limit = DefaultSearchLimit
	}

	// Sysadmins see everything; regular users are confined to their private
	// area plus the global shared folder. The index is a superset, so widen
	// the fetch and filter down.
	allowed := s.scoper.ReadPrefixes(user)

	fetchLimit := limit
	if fetched := len(allowed) * limit; fetched > fetchLimit {
		fetchLimit = fetched
	}
	// A small overspill keeps remapped results from silently dropping matches.
	fetchLimit += 100

	results, err := s.index.Search(query, fetchLimit)
	if err != nil {
		return nil, err
	}

	out := make([]*models.FileInfo, 0, len(results))
	for _, info := range results {
		rel := strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(info.Path), "/"), "/")
		if len(allowed) > 0 && !matchesAnyPrefix(rel, allowed...) {
			continue
		}
		// Hide the user's home directory itself from results; only its
		// contents are meaningful to the client.
		if s.scoper.IsPrivateRoot(user, rel) {
			continue
		}
		info.Path = s.scoper.PhysicalToVirtual(user, rel)
		out = append(out, info)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func matchesAnyPrefix(path string, prefixes ...string) bool {
	for _, p := range prefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}
