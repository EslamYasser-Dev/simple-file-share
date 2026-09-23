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
	text   ports.TextIndex
}

// DefaultSearchLimit is applied by the use case when a client does not ask
// for a specific result count, so every adapter speaks the same contract.
const DefaultSearchLimit = 50

// NewSearchFilesService builds the search use case. text may be nil, which
// disables full-text content matching and keeps name/path-only results.
func NewSearchFilesService(index ports.FileIndexRepository, scoper ports.PathScoper, text ports.TextIndex) *SearchFilesService {
	return &SearchFilesService{index: index, scoper: scoper, text: text}
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

	nameResults, err := s.index.Search(query, fetchLimit)
	if err != nil {
		return nil, err
	}

	// Full-text hits (content match) are appended after name/path matches so
	// filename relevance keeps priority; they are still scope-filtered.
	var contentPaths []string
	if s.text != nil {
		hits, hitErr := s.text.Search(query, fetchLimit)
		if hitErr != nil {
			return nil, hitErr
		}
		contentPaths = make([]string, 0, len(hits))
		for _, hit := range hits {
			contentPaths = append(contentPaths, hit.Path)
		}
	}

	seen := make(map[string]struct{}, len(nameResults)+len(contentPaths))
	out := make([]*models.FileInfo, 0, len(nameResults)+len(contentPaths))

	appendIfVisible := func(info *models.FileInfo) bool {
		physical := strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(info.Path), "/"), "/")
		if _, ok := seen[physical]; ok {
			return len(out) < limit
		}
		if len(allowed) > 0 && !matchesAnyPrefix(physical, allowed...) {
			return len(out) < limit
		}
		// Hide the user's home directory itself from results; only its
		// contents are meaningful to the client.
		if s.scoper.IsPrivateRoot(user, physical) {
			return len(out) < limit
		}
		seen[physical] = struct{}{}
		cloned := *info
		cloned.Path = s.scoper.PhysicalToVirtual(user, physical)
		out = append(out, &cloned)
		return len(out) < limit
	}

	for _, info := range nameResults {
		if !appendIfVisible(info) {
			return out, nil
		}
	}

	for _, path := range contentPaths {
		if len(out) >= limit {
			break
		}
		info, getErr := s.index.Get(path)
		if getErr != nil {
			return nil, getErr
		}
		if info == nil {
			continue
		}
		appendIfVisible(info)
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
