package models

import "time"

// Timeline event kinds recorded for the upload feed.
const (
	TimelineUpload     = "upload"
	TimelineShare      = "share"
	TimelineVisibility = "visibility"
)

// TimelineEvent is one feed entry: something an owner uploaded, shared, or
// re-scoped. Visibility is snapshotted at record time and refreshed on
// visibility changes, so feed filtering never leaks downgraded items.
// Entries carry metadata only (owner, name, size) — never raw paths outside
// the viewer's scope; serving still re-scopes through PathScoper.
type TimelineEvent struct {
	ID         string
	Owner      string
	Kind       string
	Path       string
	Name       string
	Size       int64
	Visibility VisibilityLevel
	CreatedAt  time.Time
}
