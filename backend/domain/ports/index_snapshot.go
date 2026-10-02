package ports

import (
	"errors"
	"time"
)

// ErrNoSnapshot reports that no persisted index snapshot exists yet (fresh
// install, or INDEX_SNAPSHOT disabled on a previous boot).
var ErrNoSnapshot = errors.New("no index snapshot")

// IndexSnapshotDoc is one extracted text document with the filesystem
// fingerprint it was extracted from. Embedded=false marks documents whose
// content was omitted to respect the snapshot size budget: the fingerprint is
// kept, but the next boot must re-extract the content.
type IndexSnapshotDoc struct {
	Path     string
	Size     int64
	Modified time.Time
	Content  string
	Embedded bool
}

// IndexSnapshot is the persisted state of the full-text index between boots.
// The metadata index is intentionally absent: rebuilding it from a stat walk
// is cheap, while re-extracting file contents is not.
type IndexSnapshot struct {
	TakenAt time.Time
	Docs    []IndexSnapshotDoc
}

// IndexSnapshotStore persists and restores index snapshots atomically.
// Implementations must be crash-safe: a torn write must never be loadable as
// a valid snapshot (corruption is detected and reported as an error so the
// caller falls back to a full rebuild).
type IndexSnapshotStore interface {
	// Load returns the last snapshot, ErrNoSnapshot when none exists, or a
	// decode/checksum error when the file is corrupt.
	Load() (*IndexSnapshot, error)
	// Save atomically replaces the stored snapshot.
	Save(snapshot *IndexSnapshot) error
}
