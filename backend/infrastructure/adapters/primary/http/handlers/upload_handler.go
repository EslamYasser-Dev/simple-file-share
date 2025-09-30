package handlers

import (
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// UploadHandler handles multipart file uploads.
type UploadHandler struct {
	uploadService *services.UploadService
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(uploadService *services.UploadService) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
	}
}

// ServeHTTP processes uploaded files and returns HTML response.
func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optional destination path can be provided via query param or multipart field 'path'
	destPrefix := strings.TrimPrefix(r.URL.Query().Get("path"), "/")

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Invalid multipart request", http.StatusBadRequest)
		return
	}

	var parts []models.UploadPart
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Skip malformed parts
		}
		if part.FileName() == "" {
			// Possibly a form field such as 'path'
			if part.FormName() == "path" {
				if b, readErr := ioutil.ReadAll(part); readErr == nil {
					destPrefix = strings.TrimPrefix(string(b), "/")
				}
			}
			// Non-file part; continue
			continue
		}

		filename := part.FileName()
		if destPrefix != "" {
			// Ensure we join safely and normalize to forward slashes for repository
			filename = filepath.ToSlash(filepath.Join(destPrefix, filename))
		}

		// Wrap the part to override the filename reported to the service layer
		// multipart.Part implements models.ReadCloser via embedded interfaces
		parts = append(parts, &uploadPartWithName{name: filename, rc: part})
	}

	uploads, err := h.uploadService.Execute(parts)
	if err != nil {
		http.Error(w, "Upload processing failed", http.StatusInternalServerError)
		return
	}

	renderUploadResponse(w, uploads)
}

// uploadPartWithName allows overriding the filename while passing through content
type uploadPartWithName struct {
	name string
	rc   models.ReadCloser
}

func (u *uploadPartWithName) Filename() string           { return u.name }
func (u *uploadPartWithName) Content() models.ReadCloser { return u.rc }

// renderUploadResponse generates HTML feedback for uploaded files.
func renderUploadResponse(w http.ResponseWriter, uploads []models.FileUpload) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if len(uploads) == 0 {
		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Upload Result</title></head>
			<body>
				<h3>⚠️ No files were uploaded!</h3>
				<p>Please go back and select a file or folder.</p>
				<a href="/">📁 Back to files</a>
			</body>
			</html>`)
		return
	}

	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head><title>Upload Result</title></head>
		<body>
			<h3>✅ Successfully uploaded %d file(s)!</h3>
			<ul>`, len(uploads))

	for _, upload := range uploads {
		safeName := html.EscapeString(upload.Filename)
		fmt.Fprintf(w, `<li>%s (%d bytes)</li>`, safeName, upload.Size)
	}

	fmt.Fprintf(w, `
			</ul>
			<a href="/">📁 Back to files</a>
		</body>
		</html>`)
}
