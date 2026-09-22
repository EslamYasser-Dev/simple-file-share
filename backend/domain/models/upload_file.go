package models

import "io"

type FileUpload struct {
	Filename string
	Size     int64
}

// UploadPart describes a single file being uploaded. Name is the raw file
// name supplied by the client (mime/multipart already strips directory parts),
// Destination is the virtual directory it should land in ("" = the caller's
// root), and Content is drained synchronously by the upload use case.
type UploadPart struct {
	Name        string
	Destination string
	Content     io.ReadCloser
}
