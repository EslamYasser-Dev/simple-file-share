package models

// PageData holds a directory listing result.
type PageData struct {
	Root  string
	Files []*FileInfo
}
