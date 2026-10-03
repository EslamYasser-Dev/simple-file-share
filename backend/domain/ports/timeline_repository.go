package ports

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// TimelineRepository stores feed events newest-first. Implementations must be
// safe for concurrent use.
type TimelineRepository interface {
	// Append persists one event, pruning the oldest entries beyond the
	// retention cap.
	Append(event *models.TimelineEvent) error
	// List returns events newest-first starting after the cursor
	// (createdAt+id of the last seen entry; empty cursor starts at newest),
	// up to limit entries. Limit is clamped by the caller-facing service.
	List(after string, limit int) ([]*models.TimelineEvent, error)
	// UpdateVisibility refreshes the visibility snapshot on every event of
	// owner+path (visibility downgrade must hide old entries immediately).
	UpdateVisibility(owner, path string, level models.VisibilityLevel) error
}
