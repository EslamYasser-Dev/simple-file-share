package models

// Role is a named permission set. Built-in roles (admin, member) are not
// persisted; custom roles live in the role repository.
type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	BuiltIn     bool     `json:"builtIn,omitempty"`
	Permissions []string `json:"permissions"`
}

// RoleResolver looks up effective permissions for a role name. Implementations
// merge built-in roles with custom roles from storage.
type RoleResolver interface {
	// PermissionsFor returns the permission list for role, or nil when the
	// role is unknown.
	PermissionsFor(role string) []string
	// RoleExists reports whether role is built-in or a stored custom role.
	RoleExists(role string) bool
}
