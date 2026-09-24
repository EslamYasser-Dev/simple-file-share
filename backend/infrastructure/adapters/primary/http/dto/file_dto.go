package dto

import (
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

type FileItem struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	IsDir    bool   `json:"isDir"`
	Modified string `json:"modified"`
	Version  int    `json:"version,omitempty"`
}

type UploadResult struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// UploadSession is the wire form of a resumable upload session.
type UploadSession struct {
	ID        string `json:"id"`
	Offset    int64  `json:"offset"`
	Size      int64  `json:"size"`
	ChunkSize int64  `json:"chunkSize"`
	ExpiresAt string `json:"expiresAt"`
}

// PendingUploadSession is one resumable session listed for resume UI.
type PendingUploadSession struct {
	ID          string `json:"id"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Destination string `json:"destination,omitempty"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	Offset      int64  `json:"offset"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
	ExpiresAt   string `json:"expiresAt"`
}

// OffsetMismatchError is returned with 409 when a chunk arrives at the wrong
// offset; Expected is the position the client should resume from.
type OffsetMismatchError struct {
	Error    string `json:"error"`
	Expected int64  `json:"expected"`
}

type HealthResponse struct {
	Status string `json:"status"`
	Uptime string `json:"uptime,omitempty"`
}

func FromFileInfo(f *models.FileInfo) FileItem {
	if f == nil {
		return FileItem{}
	}
	modified := ""
	if !f.Modified.IsZero() {
		modified = f.Modified.UTC().Format(time.RFC3339)
	}
	return FileItem{
		Name:     f.Name,
		Path:     f.Path,
		Size:     f.Size,
		IsDir:    f.IsDir,
		Modified: modified,
		Version:  int(f.Version),
	}
}

func FromFileInfos(files []*models.FileInfo) []FileItem {
	items := make([]FileItem, 0, len(files))
	for _, f := range files {
		items = append(items, FromFileInfo(f))
	}
	return items
}
