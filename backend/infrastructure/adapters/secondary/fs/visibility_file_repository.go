package fs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// VisibilityFileRepository persists per-file privacy settings as a JSON
// document under the storage root's metadata directory
// (`.file-share/visibility.json`). Writes are atomic (temp file + rename) and
// guarded by an in-process mutex, mirroring ShareFileRepository.
type VisibilityFileRepository struct {
	path string
	mu   sync.RWMutex
	// loaded guards the in-memory setting index keyed by owner+path. Reads
	// (every feed filter and serve check) are O(1) map lookups instead of a
	// file re-parse plus linear scan; the JSON file stays the source of
	// truth across restarts.
	loaded bool
	byFile map[string]*models.FileVisibility
}

var _ ports.VisibilityRepository = (*VisibilityFileRepository)(nil)

func NewVisibilityFileRepository(rootDir string) *VisibilityFileRepository {
	return &VisibilityFileRepository{
		path: filepath.Join(rootDir, ".file-share", "visibility.json"),
	}
}

type visibilityDocument struct {
	Owner       string    `json:"owner"`
	Path        string    `json:"path"`
	Level       string    `json:"level"`
	AllowStream bool      `json:"allowStream,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func fromVisibility(v *models.FileVisibility) visibilityDocument {
	return visibilityDocument{
		Owner:       v.Owner,
		Path:        v.Path,
		Level:       string(v.Level),
		AllowStream: v.AllowStream,
		UpdatedAt:   v.UpdatedAt,
	}
}

func (d visibilityDocument) toVisibility() *models.FileVisibility {
	level := models.VisibilityLevel(d.Level)
	if !level.Valid() {
		level = models.VisibilityPrivate
	}
	return &models.FileVisibility{
		Owner:       d.Owner,
		Path:        d.Path,
		Level:       level,
		AllowStream: d.AllowStream,
		UpdatedAt:   d.UpdatedAt,
	}
}

func visibilityKey(owner, path string) string { return owner + "\x00" + path }

func (r *VisibilityFileRepository) Set(vis *models.FileVisibility) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureLoadedLocked(); err != nil {
		return err
	}
	cp := *vis
	r.byFile[visibilityKey(vis.Owner, vis.Path)] = &cp
	return r.saveLocked(r.documentsLocked())
}

func (r *VisibilityFileRepository) Get(owner, path string) (*models.FileVisibility, error) {
	if err := r.ensureLoaded(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	if vis, ok := r.byFile[visibilityKey(owner, path)]; ok {
		cp := *vis
		return &cp, nil
	}
	// Absence is private: the safe default, not an error.
	return &models.FileVisibility{Owner: owner, Path: path, Level: models.VisibilityPrivate}, nil
}

func (r *VisibilityFileRepository) Delete(owner, path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureLoadedLocked(); err != nil {
		return err
	}
	delete(r.byFile, visibilityKey(owner, path))
	return r.saveLocked(r.documentsLocked())
}

// ensureLoaded loads the file into the index on first use. No lock is held
// across the upgrade, so concurrent first-use readers cannot deadlock.
func (r *VisibilityFileRepository) ensureLoaded() error {
	r.mu.RLock()
	if r.loaded {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ensureLoadedLocked()
}

// ensureLoadedLocked loads the file into the index; the caller holds the
// write lock.
func (r *VisibilityFileRepository) ensureLoadedLocked() error {
	if r.loaded {
		return nil
	}
	docs, err := r.loadLocked()
	if err != nil {
		return err
	}
	r.byFile = make(map[string]*models.FileVisibility, len(docs))
	for _, doc := range docs {
		vis := doc.toVisibility()
		r.byFile[visibilityKey(vis.Owner, vis.Path)] = vis
	}
	r.loaded = true
	return nil
}

// documentsLocked flattens the index back to documents, sorted for stable
// file output.
func (r *VisibilityFileRepository) documentsLocked() []visibilityDocument {
	docs := make([]visibilityDocument, 0, len(r.byFile))
	for _, vis := range r.byFile {
		docs = append(docs, fromVisibility(vis))
	}
	sort.Slice(docs, func(i, j int) bool {
		if docs[i].Owner != docs[j].Owner {
			return docs[i].Owner < docs[j].Owner
		}
		return docs[i].Path < docs[j].Path
	})
	return docs
}

func (r *VisibilityFileRepository) loadLocked() ([]visibilityDocument, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return []visibilityDocument{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read visibility file: %w", err)
	}

	var docs []visibilityDocument
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("parse visibility file: %w", err)
	}
	if docs == nil {
		docs = []visibilityDocument{}
	}
	return docs, nil
}

func (r *VisibilityFileRepository) saveLocked(docs []visibilityDocument) error {
	if err := os.MkdirAll(filepath.Dir(r.path), storageDirPerm); err != nil {
		return fmt.Errorf("create visibility dir: %w", err)
	}

	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode visibility: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(r.path), ".visibility-*.tmp")
	if err != nil {
		return fmt.Errorf("create visibility temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write visibility temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync visibility temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close visibility temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), r.path); err != nil {
		return fmt.Errorf("replace visibility file: %w", err)
	}
	return nil
}
