package models

import "time"

// User is an authenticated account. PasswordHash holds the encoded password
// hash and is never serialized to API responses. QuotaBytes is the account's
// storage allowance in bytes; 0 means unlimited. OAuthProvider/OAuthSubject
// bind the account to a stable IdP user id so display-name collisions cannot
// take over another person's account.
//
// Role is the RBAC role name (built-in "admin"/"member" or a custom role).
// IsAdmin remains a denormalized flag meaning "has files.system" so path
// scoping and legacy checks keep working; it is kept in sync when Role is set
// through role-aware helpers. Enabled=false blocks login and token use.
type User struct {
	Username      string
	PasswordHash  string
	Role          string
	IsAdmin       bool
	Enabled       bool
	QuotaBytes    int64
	CreatedAt     time.Time
	OAuthProvider string
	OAuthSubject  string
}

// Normalize fills defaults for accounts created before RBAC fields existed.
// Empty role becomes member (or admin when IsAdmin was set); missing enabled
// is treated as true by callers that pass enabledDefault when loading.
func (u *User) Normalize(role RoleResolver) {
	if u == nil {
		return
	}
	if u.Role == "" {
		if u.IsAdmin {
			u.Role = RoleAdmin
		} else {
			u.Role = RoleMember
		}
	}
	if role != nil {
		u.IsAdmin = HasPermission(role.PermissionsFor(u.Role), PermFilesSystem)
	} else if u.Role == RoleAdmin {
		u.IsAdmin = true
	} else if u.Role == RoleMember {
		u.IsAdmin = false
	}
}

// HasPermission reports whether the account grants perm. A nil user (auth
// disabled) is the system view and holds every permission. Unknown roles
// grant nothing except the nil-user case. Legacy accounts that only set
// IsAdmin (no Role field) are treated as full admins.
func (u *User) HasPermission(perm string, role RoleResolver) bool {
	if u == nil {
		return true
	}
	if u.Role == "" {
		return u.IsAdmin
	}
	var perms []string
	if role != nil {
		perms = role.PermissionsFor(u.Role)
	} else {
		perms = BuiltInPermissions(u.Role)
	}
	return HasPermission(perms, perm)
}

// IsSystemView reports whether the user has unrestricted (admin/system)
// access. A nil user is the system view used when auth is disabled.
func (u *User) IsSystemView() bool {
	return u == nil || u.IsAdmin
}

// UserStats is a read model that combines an account with its storage usage,
// used by the admin console.
type UserStats struct {
	Username   string
	Role       string
	IsAdmin    bool
	Enabled    bool
	QuotaBytes int64
	CreatedAt  time.Time
	Files      int
	Size       int64
}
