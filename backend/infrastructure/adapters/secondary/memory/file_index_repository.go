package memory

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

const defaultSearchLimit = 50

type FileIndexRepository struct {
	mu      sync.RWMutex
	entries map[string]*models.FileInfo
}

func NewFileIndexRepository() *FileIndexRepository {
	return &FileIndexRepository{
		entries: make(map[string]*models.FileInfo),
	}
}

func (r *FileIndexRepository) Search(query string, limit int) ([]*models.FileInfo, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	terms := strings.Fields(strings.ToLower(query))

	r.mu.RLock()
	defer r.mu.RUnlock()

	type scored struct {
		info  *models.FileInfo
		score int
	}
	var matches []scored

	for _, info := range r.entries {
		name := strings.ToLower(info.Name)
		path := strings.ToLower(info.Path)
		score := matchScore(name, path, terms)
		if score > 0 {
			matches = append(matches, scored{info: info, score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].info.Path < matches[j].info.Path
	})

	if len(matches) > limit {
		matches = matches[:limit]
	}

	results := make([]*models.FileInfo, len(matches))
	for i, m := range matches {
		copy := *m.info
		results[i] = &copy
	}
	return results, nil
}

func matchScore(name, path string, terms []string) int {
	score := 0
	for _, term := range terms {
		switch {
		case name == term:
			score += 100
		case strings.HasPrefix(name, term):
			score += 50
		case strings.Contains(name, term):
			score += 30
		case strings.Contains(path, term):
			score += 10
		default:
			return 0
		}
	}
	return score
}

func (r *FileIndexRepository) Upsert(info *models.FileInfo) error {
	if info == nil {
		return nil
	}

	path := normalizeIndexPath(info.Path)
	copy := *info
	copy.Path = path

	r.mu.Lock()
	r.entries[path] = &copy
	r.mu.Unlock()
	return nil
}

func (r *FileIndexRepository) Remove(path string) error {
	path = normalizeIndexPath(path)
	r.mu.Lock()
	delete(r.entries, path)
	r.mu.Unlock()
	return nil
}

func (r *FileIndexRepository) RemovePrefix(pathPrefix string) error {
	pathPrefix = normalizeIndexPath(pathPrefix)

	r.mu.Lock()
	for path := range r.entries {
		if path == pathPrefix || strings.HasPrefix(path, pathPrefix+"/") {
			delete(r.entries, path)
		}
	}
	r.mu.Unlock()
	return nil
}

func (r *FileIndexRepository) Clear() error {
	r.mu.Lock()
	r.entries = make(map[string]*models.FileInfo)
	r.mu.Unlock()
	return nil
}

func (r *FileIndexRepository) Rebuild(entries []*models.FileInfo) error {
	next := make(map[string]*models.FileInfo, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		path := normalizeIndexPath(entry.Path)
		copy := *entry
		copy.Path = path
		next[path] = &copy
	}

	r.mu.Lock()
	r.entries = next
	r.mu.Unlock()
	return nil
}

func normalizeIndexPath(path string) string {
	return strings.TrimPrefix(filepath.ToSlash(strings.TrimSpace(path)), "/")
}

var _ ports.FileIndexRepository = (*FileIndexRepository)(nil)
