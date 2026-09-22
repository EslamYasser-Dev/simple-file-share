package ports

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// PathScoper is the domain port that translates between the virtual paths
// clients use and the physical paths inside storage. Implementations encode the
// per-user namespace and shared-folder authorization policy.
type PathScoper interface {
	// PrivatePrefix returns the physical directory holding a user's private files.
	PrivatePrefix(username string) string

	// IsSystemView reports whether the user has unrestricted (admin/system)
	// access where virtual and physical paths coincide.
	IsSystemView(user *models.User) bool

	// ReadPath resolves a virtual path for a read operation.
	ReadPath(user *models.User, rel string) (string, error)

	// WritePath resolves a virtual path for a write operation, enforcing the
	// shared folder's read-only policy for regular users.
	WritePath(user *models.User, rel string) (string, error)

	// PhysicalToVirtual converts a physical path from storage back into the
	// virtual path a regular user should see.
	PhysicalToVirtual(user *models.User, physical string) string

	// ReadPrefixes returns the physical subtrees a user is allowed to read.
	// A nil slice means the user may read everything (system view).
	ReadPrefixes(user *models.User) []string

	// IsPrivateRoot reports whether rel points at the user's own home directory.
	IsPrivateRoot(user *models.User, rel string) bool
}
