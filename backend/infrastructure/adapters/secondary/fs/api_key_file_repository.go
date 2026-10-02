// Package fs persists API keys as JSON under the storage root's metadata
// directory (.file-share/api_keys.json). The file holds only hashes — a
// leaked copy does not yield usable keys. Writes are atomic and mutex
// guarded; a corrupt file resets to empty (keys can be re-issued, and unlike
// sessions there is no external trust to unwind).
package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// APIKeyFileRepository is a file-backed ports.APIKeyRepository.
type APIKeyFileRepository struct {
	path string
	mu   sync.Mutex
}

var _ ports.APIKeyRepository = (*APIKeyFileRepository)(nil)

// NewAPIKeyFileRepository stores keys at ROOT_DIR/.file-share/api_keys.json.
func NewAPIKeyFileRepository(rootDir string) *APIKeyFileRepository {
	return &APIKeyFileRepository{path: filepath.Join(rootDir, ".file-share", "api_keys.json")}
}

type apiKeyDocument struct {
	Keys []models.APIKey `json:"keys"`
}

// Create persists a new key or fails when the id collides.
func (r *APIKeyFileRepository) Create(key models.APIKey) error {
	if key.ID == "" || key.Owner == "" {
		return fmt.Errorf("api key id and owner required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	doc, err := r.loadLocked()
	if err != nil {
		return err
	}
	for _, existing := range doc.Keys {
		if existing.ID == key.ID {
			return fmt.Errorf("api key %s already exists", key.ID)
		}
	}
	doc.Keys = append(doc.Keys, key)
	return r.saveLocked(doc)
}

// List returns an owner's keys newest first (empty owner = all).
func (r *APIKeyFileRepository) List(owner string) ([]models.APIKey, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	out := make([]models.APIKey, 0, len(doc.Keys))
	for _, k := range doc.Keys {
		if owner == "" || k.Owner == owner {
			out = append(out, k)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// Find returns the key with the given id.
func (r *APIKeyFileRepository) Find(id string) (*models.APIKey, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	for i := range doc.Keys {
		if doc.Keys[i].ID == id {
			key := doc.Keys[i]
			return &key, nil
		}
	}
	return nil, domainerrors.ErrNotFound
}

// Save replaces the record with the same id.
func (r *APIKeyFileRepository) Save(key models.APIKey) error {
	if key.ID == "" {
		return fmt.Errorf("api key id required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	doc, err := r.loadLocked()
	if err != nil {
		return err
	}
	for i := range doc.Keys {
		if doc.Keys[i].ID == key.ID {
			doc.Keys[i] = key
			return r.saveLocked(doc)
		}
	}
	return domainerrors.ErrNotFound
}

// Delete removes the key with the given id.
func (r *APIKeyFileRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc, err := r.loadLocked()
	if err != nil {
		return err
	}
	kept := doc.Keys[:0]
	found := false
	for _, k := range doc.Keys {
		if k.ID == id {
			found = true
			continue
		}
		kept = append(kept, k)
	}
	if !found {
		return domainerrors.ErrNotFound
	}
	return r.saveLocked(apiKeyDocument{Keys: kept})
}

func (r *APIKeyFileRepository) loadLocked() (apiKeyDocument, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return apiKeyDocument{Keys: []models.APIKey{}}, nil
		}
		return apiKeyDocument{}, fmt.Errorf("read api keys: %w", err)
	}
	var doc apiKeyDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		// A corrupt file must not brick auth: start empty (existing keys
		// stop working, which fails closed).
		return apiKeyDocument{Keys: []models.APIKey{}}, nil
	}
	if doc.Keys == nil {
		doc.Keys = []models.APIKey{}
	}
	return doc, nil
}

func (r *APIKeyFileRepository) saveLocked(doc apiKeyDocument) error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create api keys dir: %w", err)
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal api keys: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".api-keys-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp api keys: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write api keys: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("sync api keys: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close api keys: %w", err)
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		os.Remove(tmpName)
		return nil // best effort — the directory is already private
	}
	if err := os.Rename(tmpName, r.path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("replace api keys: %w", err)
	}
	return nil
}
