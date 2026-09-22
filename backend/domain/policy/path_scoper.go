// Package policy contains pure domain services that encode business rules
// without any I/O or framework dependencies.
package policy

import (
	"path/filepath"
	"strings"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// Private sub-tree that holds each user's home directory.
const UsersRoot = "users"

// Public sub-tree readable by every user (writable only by admins).
const SharedRoot = "shared"

// PathScoper maps virtual paths (what a client sees and sends) to physical
// paths relative to the storage root, and back.
//
// Virtual namespace:
//   - Admin (or system views where auth is disabled) address the real root,
//     so virtual and physical paths coincide.
//   - Regular users address their private home as `/` (physical
//     `users/<name>/...`) plus the global shared folder under `/shared`
//     (physical `shared/...`). Anything under `users/` belonging to another
//     account is unreachable.
type PathScoper struct{}

var _ ports.PathScoper = (*PathScoper)(nil)

func NewPathScoper() *PathScoper {
	return &PathScoper{}
}

// PrivatePrefix returns the physical directory holding a user's private files.
func (s *PathScoper) PrivatePrefix(username string) string {
	return filepath.ToSlash(filepath.Join(UsersRoot, username))
}

// IsSystemView reports whether the user has unrestricted (admin/system) access.
func (s *PathScoper) IsSystemView(user *models.User) bool {
	return user.IsSystemView()
}

// ReadPath resolves a virtual path for reading. Any regular user can read
// their own home and the global shared folder.
func (s *PathScoper) ReadPath(user *models.User, rel string) (string, error) {
	if s.IsSystemView(user) {
		return rel, nil
	}

	rel = strings.Trim(rel, "/")
	first, _, _ := strings.Cut(rel, "/")
	switch first {
	case UsersRoot:
		return "", &domainerrors.ForbiddenError{Action: "access", Path: rel}
	case SharedRoot:
		return rel, nil
	default:
		if rel == "" {
			return s.PrivatePrefix(user.Username), nil
		}
		return filepath.ToSlash(filepath.Join(s.PrivatePrefix(user.Username), rel)), nil
	}
}

// WritePath resolves a virtual path for a write operation and enforces the
// shared folder's read-only policy for regular users.
func (s *PathScoper) WritePath(user *models.User, rel string) (string, error) {
	physical, err := s.ReadPath(user, rel)
	if err != nil {
		return "", err
	}

	if !s.IsSystemView(user) && isSharedPath(rel) {
		return "", &domainerrors.ForbiddenError{Action: "write", Path: rel}
	}
	return physical, nil
}

// PhysicalToVirtual converts a physical path from storage back to the virtual
// path a regular user should see. System views see physical paths unchanged.
func (s *PathScoper) PhysicalToVirtual(user *models.User, physical string) string {
	if s.IsSystemView(user) {
		return physical
	}

	physical = strings.Trim(physical, "/")
	prefix := s.PrivatePrefix(user.Username)
	if physical == prefix {
		return ""
	}
	if strings.HasPrefix(physical, prefix+"/") {
		return strings.TrimPrefix(physical, prefix+"/")
	}
	// shared/** and any unprefixed path map to themselves.
	return physical
}

// ReadPrefixes returns the physical subtrees a user may read. A nil slice means
// unrestricted access (system view).
func (s *PathScoper) ReadPrefixes(user *models.User) []string {
	if s.IsSystemView(user) {
		return nil
	}
	return []string{s.PrivatePrefix(user.Username), SharedRoot}
}

// IsPrivateRoot reports whether rel points at the user's own home directory.
func (s *PathScoper) IsPrivateRoot(user *models.User, rel string) bool {
	if s.IsSystemView(user) {
		return false
	}
	return strings.Trim(rel, "/") == s.PrivatePrefix(user.Username)
}

func isSharedPath(rel string) bool {
	rel = strings.Trim(rel, "/")
	return rel == SharedRoot || strings.HasPrefix(rel, SharedRoot+"/")
}
