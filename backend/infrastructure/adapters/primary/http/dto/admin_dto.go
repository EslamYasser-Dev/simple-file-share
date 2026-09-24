package dto

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// CreateUserRequest is the body of POST /api/admin/users.
type CreateUserRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	Role            string `json:"role,omitempty"`
	QuotaBytes      int64  `json:"quotaBytes,omitempty"`
	Enabled         *bool  `json:"enabled,omitempty"`
	UseDefaultQuota bool   `json:"useDefaultQuota,omitempty"`
}

// UpdateUserRequest is the body of PATCH/PUT /api/admin/users/{username}.
// Pointers mean "leave unchanged".
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty"`
	Role     *string `json:"role,omitempty"`
	Enabled  *bool   `json:"enabled,omitempty"`
}

// ResetPasswordRequest is the body of POST /api/admin/users/{username}/password.
type ResetPasswordRequest struct {
	Password string `json:"password"`
}

// ChangePasswordRequest is the body of POST /api/auth/password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// UpsertRoleRequest is the body of POST/PUT /api/admin/roles.
type UpsertRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions"`
}

// RoleResponse is the wire shape of a role.
type RoleResponse struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	BuiltIn     bool     `json:"builtIn,omitempty"`
	Permissions []string `json:"permissions"`
}

func FromRole(r models.Role) RoleResponse {
	perms := r.Permissions
	if perms == nil {
		perms = []string{}
	}
	return RoleResponse{
		Name:        r.Name,
		Description: r.Description,
		BuiltIn:     r.BuiltIn,
		Permissions: perms,
	}
}

func FromRoles(roles []models.Role) []RoleResponse {
	out := make([]RoleResponse, 0, len(roles))
	for _, r := range roles {
		out = append(out, FromRole(r))
	}
	return out
}
