package ports

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// VisibilityRepository stores per-file privacy settings. Implementations must
// be safe for concurrent use. A missing record means private.
type VisibilityRepository interface {
	// Set stores (or replaces) the visibility for owner+path.
	Set(vis *models.FileVisibility) error
	// Get returns the visibility for owner+path. When no record exists it
	// returns a private default with AllowStream false, not an error.
	Get(owner, path string) (*models.FileVisibility, error)
	// Delete removes the record for owner+path (file deleted); missing is a
	// no-op.
	Delete(owner, path string) error
}
