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
}

type UploadResult struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
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
	}
}

func FromFileInfos(files []*models.FileInfo) []FileItem {
	items := make([]FileItem, 0, len(files))
	for _, f := range files {
		items = append(items, FromFileInfo(f))
	}
	return items
}
