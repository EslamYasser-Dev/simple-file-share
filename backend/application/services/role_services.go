package services

import (
	"sort"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

var rolePattern = mustCompileRole()

// ListRolesService returns every role (built-in + custom) for the admin console.
type ListRolesService struct {
	roles *RoleCatalog
}

func NewListRolesService(roles *RoleCatalog) *ListRolesService {
	return &ListRolesService{roles: roles}
}

func (s *ListRolesService) Execute(actor *models.User) ([]models.Role, error) {
	if !canManageRoles(actor, s.roles) {
		return nil, forbiddenErr("manage roles")
	}
	return s.roles.List(), nil
}

// CreateOrUpdateRoleService upserts a custom role (built-ins rejected).
type CreateOrUpdateRoleService struct {
	users ports.UserRepository
	repo  ports.RoleRepository
	roles *RoleCatalog
}

func NewCreateOrUpdateRoleService(users ports.UserRepository, repo ports.RoleRepository, roles *RoleCatalog) *CreateOrUpdateRoleService {
	return &CreateOrUpdateRoleService{users: users, repo: repo, roles: roles}
}

func (s *CreateOrUpdateRoleService) Execute(actor *models.User, name, description string, permissions []string) (models.Role, error) {
	if !canManageRoles(actor, s.roles) {
		return models.Role{}, forbiddenErr("manage roles")
	}
	name = strings.TrimSpace(strings.ToLower(name))
	if !rolePattern.MatchString(name) {
		return models.Role{}, validation("role", name, "role must be 2-32 chars: lowercase letters, digits, dashes, underscores")
	}
	if models.BuiltInRole(name) {
		return models.Role{}, validation("role", name, "built-in roles cannot be modified")
	}
	perms, err := sanitizePermissions(permissions)
	if err != nil {
		return models.Role{}, err
	}
	role := models.Role{Name: name, Description: strings.TrimSpace(description), Permissions: perms}
	if err := s.repo.Upsert(role); err != nil {
		return models.Role{}, err
	}
	s.roles.Reload()
	return role, nil
}

// DeleteRoleService removes a custom role after ensuring no user still uses it.
type DeleteRoleService struct {
	users ports.UserRepository
	repo  ports.RoleRepository
	roles *RoleCatalog
}

func NewDeleteRoleService(users ports.UserRepository, repo ports.RoleRepository, roles *RoleCatalog) *DeleteRoleService {
	return &DeleteRoleService{users: users, repo: repo, roles: roles}
}

func (s *DeleteRoleService) Execute(actor *models.User, name string) error {
	if !canManageRoles(actor, s.roles) {
		return forbiddenErr("manage roles")
	}
	name = strings.TrimSpace(name)
	if models.BuiltInRole(name) {
		return validation("role", name, "built-in roles cannot be deleted")
	}
	list, err := s.users.ListUsers()
	if err != nil {
		return err
	}
	for _, u := range list {
		if u.Role == name {
			return validation("role", name, "role is still assigned to users")
		}
	}
	if err := s.repo.Delete(name); err != nil {
		return err
	}
	s.roles.Reload()
	return nil
}

func canManageRoles(actor *models.User, roles *RoleCatalog) bool {
	if actor == nil {
		return true
	}
	return actor.IsSystemView() || actor.HasPermission(models.PermRolesManage, roles)
}

func sanitizePermissions(in []string) ([]string, error) {
	allowed := map[string]bool{}
	for _, p := range models.AllPermissions() {
		allowed[p] = true
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, p := range in {
		p = strings.TrimSpace(p)
		if p == "" || !allowed[p] {
			return nil, validation("permissions", p, "unknown permission")
		}
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}
