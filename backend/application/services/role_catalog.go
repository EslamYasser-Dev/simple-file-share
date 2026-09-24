package services

import (
	"sort"
	"sync"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// RoleCatalog merges built-in roles with custom roles from storage and
// implements models.RoleResolver.
type RoleCatalog struct {
	repo ports.RoleRepository
	mu   sync.RWMutex
	// cache is rebuilt from the repository on each mutation; reads hit cache.
	custom map[string]models.Role
}

var _ models.RoleResolver = (*RoleCatalog)(nil)

func NewRoleCatalog(repo ports.RoleRepository) *RoleCatalog {
	c := &RoleCatalog{repo: repo, custom: map[string]models.Role{}}
	c.Reload()
	return c
}

// Reload re-reads custom roles from storage. Safe to call after Upsert/Delete.
func (c *RoleCatalog) Reload() {
	if c == nil || c.repo == nil {
		return
	}
	roles, err := c.repo.List()
	if err != nil {
		return
	}
	next := make(map[string]models.Role, len(roles))
	for _, r := range roles {
		if models.BuiltInRole(r.Name) {
			continue
		}
		next[r.Name] = r
	}
	c.mu.Lock()
	c.custom = next
	c.mu.Unlock()
}

func (c *RoleCatalog) PermissionsFor(role string) []string {
	if role == "" {
		return nil
	}
	if models.BuiltInRole(role) {
		return models.BuiltInPermissions(role)
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if r, ok := c.custom[role]; ok {
		return r.Permissions
	}
	return nil
}

func (c *RoleCatalog) RoleExists(role string) bool {
	if models.BuiltInRole(role) {
		return true
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.custom[role]
	return ok
}

// List returns built-in roles first, then custom roles, all sorted by name
// within each group.
func (c *RoleCatalog) List() []models.Role {
	out := []models.Role{
		{
			Name:        models.RoleAdmin,
			Description: "Full administrative access",
			BuiltIn:     true,
			Permissions: models.AllPermissions(),
		},
		{
			Name:        models.RoleMember,
			Description: "Standard file user",
			BuiltIn:     true,
			Permissions: nil,
		},
	}
	c.mu.RLock()
	customs := make([]models.Role, 0, len(c.custom))
	for _, r := range c.custom {
		customs = append(customs, models.Role{Name: r.Name, Description: r.Description, Permissions: r.Permissions})
	}
	c.mu.RUnlock()
	sort.Slice(customs, func(i, j int) bool { return customs[i].Name < customs[j].Name })
	return append(out, customs...)
}
