package models

type FileUpload struct {
	Filename string
	Size     int64
}

type UploadPart interface {
	Filename() string
	Content() ReadCloser
}

type ReadCloser interface {
	Read(p []byte) (n int, err error)
	Close() error
}
