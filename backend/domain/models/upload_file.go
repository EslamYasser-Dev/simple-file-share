package models

import "io"

type FileUpload struct {
	Filename string
	Size     int64
}

type UploadPart interface {
	Filename() string
	Content() io.ReadCloser
}
