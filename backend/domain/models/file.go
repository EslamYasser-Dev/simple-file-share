package models

import "time"

// FileInfo represents a file or directory entry in the storage root.
type FileInfo struct {
	Name     string
	Path     string
	Size     int64
	IsDir    bool
	Modified time.Time
}
