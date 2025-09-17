package xhttp

import (
	"mime/multipart"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

// UploadPartAdapter adapts *multipart.Part to domain.UploadPart.
// This is an Anti-Corruption Layer between HTTP multipart and domain.
type UploadPartAdapter struct {
	Part *multipart.Part
}

// Filename returns the original filename from the HTTP form.
func (upa *UploadPartAdapter) Filename() string {
	return upa.Part.FileName()
}

// Content returns the file content stream as domain.ReadCloser.
func (upa *UploadPartAdapter) Content() models.ReadCloser {
	return upa.Part
}

// Ensure *multipart.Part implements domain.ReadCloser (it does via embedded io.Reader + Close())
var _ models.ReadCloser = (*multipart.Part)(nil)
