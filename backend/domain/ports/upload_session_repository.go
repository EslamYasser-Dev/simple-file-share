package ports

import (
	"io"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

// UploadSessionRepository stores resumable upload metadata and the staged
// bytes for each session. Implementations must be safe for concurrent use.
type UploadSessionRepository interface {
	// Create persists a new session and creates an empty staging file.
	Create(session *models.UploadSession) error
	// Get looks up a session by id.
	Get(id string) (*models.UploadSession, error)
	// FindByFingerprint returns the newest active session for owner+fp, or
	// domainerrors.ErrNotFound when none matches.
	FindByFingerprint(owner, fingerprint string) (*models.UploadSession, error)
	// List returns active (non-expired) sessions for owner, newest first.
	List(owner string) ([]*models.UploadSession, error)
	// UpdateOffset advances the staged offset (and refreshes UpdatedAt).
	// The implementation must persist both metadata and any prior staging data.
	UpdateOffset(id string, offset int64, updatedAt time.Time) error
	// OpenStaging opens the staged bytes for reading (complete) or writing
	// (append). Callers own the returned closer.
	OpenStaging(id string, write bool) (io.ReadWriteCloser, error)
	// Append writes body at offset into the staging file and returns the new
	// size. offset must match the current staging size.
	Append(id string, offset int64, body io.Reader) (int64, error)
	// Delete removes the session and its staging file.
	Delete(id string) error
	// PurgeExpired removes sessions past expiresAt and returns how many were
	// removed.
	PurgeExpired(now time.Time) (int, error)
}
