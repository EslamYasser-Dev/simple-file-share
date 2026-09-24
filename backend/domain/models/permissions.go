package models

// Permission strings used by the RBAC layer. Services authorize with
// User.HasPermission so HTTP and gRPC share one gate.
const (
	PermUsersRead    = "users.read"
	PermUsersCreate  = "users.create"
	PermUsersUpdate  = "users.update"
	PermUsersDelete  = "users.delete"
	PermQuotaManage  = "quota.manage"
	PermRolesManage  = "roles.manage"
	PermSharesManage = "shares.manage"
	PermFilesSystem  = "files.system"
)

// AllPermissions is the full set granted to the built-in admin role.
func AllPermissions() []string {
	return []string{
		PermUsersRead,
		PermUsersCreate,
		PermUsersUpdate,
		PermUsersDelete,
		PermQuotaManage,
		PermRolesManage,
		PermSharesManage,
		PermFilesSystem,
	}
}

// AllPermissionsIfAdmin returns the full permission set when isAdmin, else nil.
// Used by HTTP DTOs when the caller only has the denormalized IsAdmin flag.
func AllPermissionsIfAdmin(isAdmin bool) []string {
	if isAdmin {
		return AllPermissions()
	}
	return nil
}

// RoleNames for the built-in roles (never deletable).
const (
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// BuiltInRole reports whether name is a reserved built-in role.
func BuiltInRole(name string) bool {
	return name == RoleAdmin || name == RoleMember
}

// BuiltInPermissions returns the permission set for a built-in role.
// Unknown names yield nil (no elevated permissions).
func BuiltInPermissions(name string) []string {
	switch name {
	case RoleAdmin:
		return AllPermissions()
	default:
		return nil
	}
}

// HasPermission reports whether perms contains perm.
func HasPermission(perms []string, perm string) bool {
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}
