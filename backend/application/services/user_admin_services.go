package services

import (
	"strings"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// CreateUserInput is the admin-provisioned account payload.
type CreateUserInput struct {
	Username   string
	Password   string
	Role       string
	QuotaBytes int64
	// Enabled nil means create enabled (default); false creates a disabled account.
	Enabled         *bool
	UseDefaultQuota bool
}

// CreateUserService provisions an account from the admin console (bypasses signup).
type CreateUserService struct {
	users             ports.UserRepository
	hasher            ports.PasswordHasher
	fileRepo          ports.FileRepository
	scoper            ports.PathScoper
	roles             *RoleCatalog
	defaultQuotaBytes int64
	tokens            ports.TokenManager
}

func NewCreateUserService(
	users ports.UserRepository,
	hasher ports.PasswordHasher,
	fileRepo ports.FileRepository,
	scoper ports.PathScoper,
	roles *RoleCatalog,
	defaultQuotaBytes int64,
) *CreateUserService {
	return &CreateUserService{
		users:             users,
		hasher:            hasher,
		fileRepo:          fileRepo,
		scoper:            scoper,
		roles:             roles,
		defaultQuotaBytes: defaultQuotaBytes,
	}
}

// SetTokenManager enables session revocation on destructive account ops.
func (s *CreateUserService) SetTokenManager(t ports.TokenManager) { s.tokens = t }

func (s *CreateUserService) Execute(actor *models.User, in CreateUserInput) (*models.User, error) {
	if !actor.HasPermission(models.PermUsersCreate, s.roles) {
		return nil, forbiddenErr("create users")
	}
	username := strings.TrimSpace(in.Username)
	if !usernamePattern.MatchString(username) {
		return nil, validation("username", username, "username must be 3-32 characters (letters, digits, dots, dashes, underscores)")
	}
	if len(in.Password) < minPasswordLength {
		return nil, validation("password", "", "password must be at least 4 characters")
	}
	role := strings.TrimSpace(in.Role)
	if role == "" {
		role = models.RoleMember
	}
	if !s.roles.RoleExists(role) {
		return nil, validation("role", role, "unknown role")
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	quota := s.defaultQuotaBytes
	if in.UseDefaultQuota {
		quota = s.defaultQuotaBytes
	} else if in.QuotaBytes >= 0 && !in.UseDefaultQuota && in.QuotaBytes != 0 {
		quota = in.QuotaBytes
	}

	user := &models.User{
		Username:     username,
		PasswordHash: hash,
		Role:         role,
		Enabled:      true,
		QuotaBytes:   quota,
		CreatedAt:    time.Now().UTC(),
	}
	user.Normalize(s.roles)
	if in.Enabled != nil && !*in.Enabled {
		user.Enabled = false
	}

	if err := s.users.CreateUser(user); err != nil {
		return nil, err
	}
	_ = s.fileRepo.CreateDirectory(s.scoper.PrivatePrefix(user.Username))
	user.PasswordHash = ""
	return user, nil
}

// UpdateUserInput carries optional profile fields; nil means "leave unchanged".
type UpdateUserInput struct {
	Username *string
	Role     *string
	Enabled  *bool
}

// UpdateUserService edits profile, role, enablement, and optional rename.
type UpdateUserService struct {
	users  ports.UserRepository
	roles  *RoleCatalog
	scoper ports.PathScoper
	mover  ports.StorageMover
	shares ports.ShareRepository
	index  ports.FileIndexRepository
	tokens ports.TokenManager
}

func NewUpdateUserService(
	users ports.UserRepository,
	roles *RoleCatalog,
	scoper ports.PathScoper,
	mover ports.StorageMover,
	shares ports.ShareRepository,
	index ports.FileIndexRepository,
) *UpdateUserService {
	return &UpdateUserService{users: users, roles: roles, scoper: scoper, mover: mover, shares: shares, index: index}
}

// SetTokenManager enables revoking sessions on rename/disable/role change.
func (s *UpdateUserService) SetTokenManager(t ports.TokenManager) { s.tokens = t }

func (s *UpdateUserService) Execute(actor *models.User, targetUsername string, in UpdateUserInput) (*models.User, error) {
	if !actor.HasPermission(models.PermUsersUpdate, s.roles) {
		return nil, forbiddenErr("update users")
	}

	target, err := s.users.FindByUsername(targetUsername)
	if err != nil {
		return nil, err
	}

	oldName := target.Username
	renamed := false
	if in.Username != nil {
		newName := strings.TrimSpace(*in.Username)
		if newName != "" && newName != target.Username {
			if !usernamePattern.MatchString(newName) {
				return nil, validation("username", newName, "username must be 3-32 characters (letters, digits, dots, dashes, underscores)")
			}
			if _, err := s.users.FindByUsername(newName); err == nil {
				return nil, domainerrors.ErrUserAlreadyExists
			} else if !isNotFound(err) {
				return nil, err
			}
			// Only the account owner (when self-service is later added) or
			// users.update may rename; admins rename freely.
			target.Username = newName
			renamed = true
		}
	}

	if in.Role != nil {
		role := strings.TrimSpace(*in.Role)
		if !s.roles.RoleExists(role) {
			return nil, validation("role", role, "unknown role")
		}
		if models.BuiltInRole(target.Role) && target.Role == models.RoleAdmin && role != models.RoleAdmin {
			if err := s.ensureNotLastAdmin(target.Username); err != nil {
				return nil, err
			}
		}
		if target.Role == models.RoleAdmin || target.IsAdmin {
			if role != models.RoleAdmin && !hasPerm(s.roles.PermissionsFor(role), models.PermFilesSystem) {
				if err := s.ensureNotLastAdmin(target.Username); err != nil {
					return nil, err
				}
			}
		}
		target.Role = role
		target.Normalize(s.roles)
	}

	if in.Enabled != nil {
		if !*in.Enabled && (target.IsAdmin || hasPerm(s.roles.PermissionsFor(target.Role), models.PermFilesSystem)) {
			if err := s.ensureNotLastAdmin(target.Username); err != nil {
				return nil, err
			}
		}
		target.Enabled = *in.Enabled
	}

	if renamed {
		if err := s.renameAccountStorage(oldName, target.Username); err != nil {
			return nil, err
		}
		if err := s.reassignShares(oldName, target.Username); err != nil {
			return nil, err
		}
	}

	if err := s.users.UpdateUser(target, oldName); err != nil {
		return nil, err
	}

	// Kill existing sessions on identity/privilege changes.
	if renamed || (in.Enabled != nil && !*in.Enabled) || in.Role != nil {
		if s.tokens != nil {
			s.tokens.RevokeSubject(oldName, time.Now().Add(24*time.Hour))
			if renamed {
				s.tokens.RevokeSubject(target.Username, time.Now().Add(24*time.Hour))
			}
		}
	}

	target.PasswordHash = ""
	return target, nil
}

func (s *UpdateUserService) ensureNotLastAdmin(target string) error {
	list, err := s.users.ListUsers()
	if err != nil {
		return err
	}
	remaining := 0
	for _, u := range list {
		if u.Username == target {
			continue
		}
		if u.Enabled && (u.IsAdmin || hasPerm(s.roles.PermissionsFor(u.Role), models.PermFilesSystem)) {
			remaining++
		}
	}
	if remaining == 0 {
		return validation("role", target, "cannot remove or disable the last enabled admin")
	}
	return nil
}

func (s *UpdateUserService) renameAccountStorage(oldName, newName string) error {
	if s.mover == nil {
		return nil
	}
	oldPath := s.scoper.PrivatePrefix(oldName)
	newPath := s.scoper.PrivatePrefix(newName)
	return s.mover.MovePath(oldPath, newPath)
}

func (s *UpdateUserService) reassignShares(oldName, newName string) error {
	if s.shares == nil {
		return nil
	}
	shares, err := s.shares.ListByOwner(oldName)
	if err != nil {
		return err
	}
	for _, sh := range shares {
		// Share model owner rewrite requires Create with same token.
		cp := *sh
		cp.Owner = newName
		if err := s.shares.Create(&cp); err != nil {
			return err
		}
	}
	return nil
}

// DeleteUserService hard-deletes an account and cascades files + shares.
type DeleteUserService struct {
	users  ports.UserRepository
	roles  *RoleCatalog
	scoper ports.PathScoper
	files  ports.FileRepository
	shares ports.ShareRepository
	tokens ports.TokenManager
}

func NewDeleteUserService(
	users ports.UserRepository,
	roles *RoleCatalog,
	scoper ports.PathScoper,
	files ports.FileRepository,
	shares ports.ShareRepository,
) *DeleteUserService {
	return &DeleteUserService{users: users, roles: roles, scoper: scoper, files: files, shares: shares}
}

func (s *DeleteUserService) SetTokenManager(t ports.TokenManager) { s.tokens = t }

func (s *DeleteUserService) Execute(actor *models.User, username string) error {
	if !actor.HasPermission(models.PermUsersDelete, s.roles) {
		return forbiddenErr("delete users")
	}
	if actor.Username == username {
		return validation("username", username, "cannot delete your own account")
	}
	target, err := s.users.FindByUsername(username)
	if err != nil {
		return err
	}
	if target.IsAdmin || hasPerm(s.roles.PermissionsFor(target.Role), models.PermFilesSystem) {
		list, err := s.users.ListUsers()
		if err != nil {
			return err
		}
		remaining := 0
		for _, u := range list {
			if u.Username == username {
				continue
			}
			if u.Enabled && (u.IsAdmin || hasPerm(s.roles.PermissionsFor(u.Role), models.PermFilesSystem)) {
				remaining++
			}
		}
		if remaining == 0 {
			return validation("username", username, "cannot delete the last enabled admin")
		}
	}

	if owned, err := s.shares.ListByOwner(username); err == nil {
		for _, sh := range owned {
			_ = s.shares.Delete(sh.Token)
		}
	}

	if s.files != nil {
		_ = s.files.DeletePath(s.scoper.PrivatePrefix(username))
	}
	if s.tokens != nil {
		s.tokens.RevokeSubject(username, time.Now().Add(24*time.Hour))
	}
	return s.users.DeleteUser(username)
}

// ResetPasswordService lets an admin set a new password without the old one.
type ResetPasswordService struct {
	users  ports.UserRepository
	hasher ports.PasswordHasher
	roles  *RoleCatalog
	tokens ports.TokenManager
}

func NewResetPasswordService(users ports.UserRepository, hasher ports.PasswordHasher, roles *RoleCatalog) *ResetPasswordService {
	return &ResetPasswordService{users: users, hasher: hasher, roles: roles}
}

func (s *ResetPasswordService) SetTokenManager(t ports.TokenManager) { s.tokens = t }

func (s *ResetPasswordService) Execute(actor *models.User, username, newPassword string) error {
	if !actor.HasPermission(models.PermUsersUpdate, s.roles) {
		return forbiddenErr("reset passwords")
	}
	if len(newPassword) < minPasswordLength {
		return validation("password", "", "password must be at least 4 characters")
	}
	user, err := s.users.FindByUsername(username)
	if err != nil {
		return err
	}
	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	if err := s.users.UpdateUser(user, user.Username); err != nil {
		return err
	}
	if s.tokens != nil {
		s.tokens.RevokeSubject(username, time.Now().Add(24*time.Hour))
	}
	return nil
}

// ChangePasswordService lets a user change their own password.
type ChangePasswordService struct {
	users  ports.UserRepository
	hasher ports.PasswordHasher
	tokens ports.TokenManager
}

func NewChangePasswordService(users ports.UserRepository, hasher ports.PasswordHasher, tokens ports.TokenManager) *ChangePasswordService {
	return &ChangePasswordService{users: users, hasher: hasher, tokens: tokens}
}

func (s *ChangePasswordService) Execute(actor *models.User, currentPassword, newPassword string) error {
	if actor == nil {
		return forbiddenErr("change password")
	}
	if len(newPassword) < minPasswordLength {
		return validation("password", "", "password must be at least 4 characters")
	}
	user, err := s.users.FindByUsername(actor.Username)
	if err != nil {
		return err
	}
	if user.PasswordHash == "" {
		return validation("password", "", "account has no password (OAuth-only); use admin reset")
	}
	if !s.hasher.Verify(currentPassword, user.PasswordHash) {
		return domainerrors.ErrInvalidCredentials
	}
	hash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	if err := s.users.UpdateUser(user, user.Username); err != nil {
		return err
	}
	// Keep the current session: self password change does not revoke tokens.
	// Admin resets and disable/rename still revoke via RevokeSubject.
	return nil
}

func forbiddenErr(action string) error {
	return &domainerrors.ForbiddenError{Action: action}
}

func isNotFound(err error) bool {
	return err != nil && (err == domainerrors.ErrUserNotFound || err == domainerrors.ErrNotFound)
}

func hasPerm(perms []string, perm string) bool {
	return models.HasPermission(perms, perm)
}
