package ports

import (
	"io"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

// VersionRepository exposes the historical snapshots kept for overwritten files.
type VersionRepository interface {
	// ListVersions returns the numbered snapshots of a file, oldest first.
	ListVersions(path string) ([]*models.FileInfo, error)
	// ServeVersion streams the nth snapshot (1-indexed) with the canonical file name.
	ServeVersion(path string, n int) (io.ReadCloser, string, error)
	// RestoreVersion replaces the file's content with the nth snapshot, first
	// snapshotting the current content so nothing is lost.
	RestoreVersion(path string, n int) error
}
