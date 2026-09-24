package ports

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// RoleRepository stores custom roles. Built-in roles are not persisted here.
type RoleRepository interface {
	// List returns every stored custom role ordered by name.
	List() ([]models.Role, error)
	// Get returns a custom role by name, or domainerrors.ErrNotFound.
	Get(name string) (models.Role, error)
	// Upsert creates or replaces a custom role.
	Upsert(role models.Role) error
	// Delete removes a custom role. Built-ins must be rejected by the caller.
	Delete(name string) error
}
