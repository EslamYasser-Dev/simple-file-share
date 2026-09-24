package fs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// RoleFileRepository persists custom roles under the storage root metadata
// directory (`.file-share/roles.json`). Built-in roles are never stored here.
type RoleFileRepository struct {
	path string
	mu   sync.RWMutex
}

var _ ports.RoleRepository = (*RoleFileRepository)(nil)

func NewRoleFileRepository(rootDir string) *RoleFileRepository {
	return &RoleFileRepository{
		path: filepath.Join(rootDir, ".file-share", "roles.json"),
	}
}

type roleDocument struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions"`
}

func (r *RoleFileRepository) List() ([]models.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	docs, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	roles := make([]models.Role, 0, len(docs))
	for _, d := range docs {
		roles = append(roles, models.Role{
			Name:        d.Name,
			Description: d.Description,
			Permissions: d.Permissions,
		})
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i].Name < roles[j].Name })
	return roles, nil
}

func (r *RoleFileRepository) Get(name string) (models.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	docs, err := r.loadLocked()
	if err != nil {
		return models.Role{}, err
	}
	d, ok := docs[name]
	if !ok {
		return models.Role{}, domainerrors.ErrNotFound
	}
	return models.Role{Name: d.Name, Description: d.Description, Permissions: d.Permissions}, nil
}

func (r *RoleFileRepository) Upsert(role models.Role) error {
	if role.Name == "" {
		return domainerrors.NewValidationError("role", "", "role name is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	docs, err := r.loadLocked()
	if err != nil {
		return err
	}
	docs[role.Name] = roleDocument{
		Name:        role.Name,
		Description: role.Description,
		Permissions: role.Permissions,
	}
	return r.saveLocked(docs)
}

func (r *RoleFileRepository) Delete(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	docs, err := r.loadLocked()
	if err != nil {
		return err
	}
	if _, ok := docs[name]; !ok {
		return domainerrors.ErrNotFound
	}
	delete(docs, name)
	return r.saveLocked(docs)
}

func (r *RoleFileRepository) loadLocked() (map[string]roleDocument, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]roleDocument{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read roles file: %w", err)
	}
	var docs map[string]roleDocument
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("parse roles file: %w", err)
	}
	if docs == nil {
		docs = map[string]roleDocument{}
	}
	return docs, nil
}

func (r *RoleFileRepository) saveLocked(docs map[string]roleDocument) error {
	if err := os.MkdirAll(filepath.Dir(r.path), storageDirPerm); err != nil {
		return fmt.Errorf("create roles dir: %w", err)
	}
	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode roles: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(r.path), ".roles-*.tmp")
	if err != nil {
		return fmt.Errorf("create roles temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write roles temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync roles temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close roles temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), r.path); err != nil {
		return fmt.Errorf("replace roles file: %w", err)
	}
	return nil
}
