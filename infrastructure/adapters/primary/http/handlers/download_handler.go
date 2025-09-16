package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/domain/errors"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// DownloadHandler handles file and directory ZIP downloads.
type DownloadHandler struct {
	fileService *services.DownloadFileService
	zipService  *services.DownloadZipService
}

// NewDownloadHandler creates a new DownloadHandler.
func NewDownloadHandler(fileService *services.DownloadFileService, zipService *services.DownloadZipService) *DownloadHandler {
	return &DownloadHandler{
		fileService: fileService,
		zipService:  zipService,
	}
}

// ServeHTTP determines whether to serve a file or a ZIP archive.
func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := cleanPath(r.URL.Path)
	if containsPathTraversal(path) {
		http.Error(w, "Path traversal detected", http.StatusForbidden)
		return
	}

	isZipRequest := strings.HasSuffix(path, ".zip")
	if isZipRequest {
		path = strings.TrimSuffix(path, ".zip")
		stream, filename, err := h.zipService.Execute(path)
		if err != nil {
			respondWithError(w, err)
			return
		}
		if stream == nil {
			http.Error(w, "Not a directory", http.StatusBadRequest)
			return
		}
		serveDownload(w, stream, filename, "application/zip")
		return
	}

	stream, filename, err := h.fileService.Execute(path)
	if err != nil {
		respondWithError(w, err)
		return
	}
	if stream == nil {
		http.Error(w, "Is a directory", http.StatusConflict)
		return
	}
	serveDownload(w, stream, filename, "application/octet-stream")
}

// serveDownload writes stream to HTTP response with headers.
func serveDownload(w http.ResponseWriter, stream io.ReadCloser, filename, contentType string) {
	defer stream.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if _, err := io.Copy(w, stream); err != nil {
		http.Error(w, "Stream copy failed", http.StatusInternalServerError)
	}
}

// respondWithError maps domain errors to HTTP status codes.
func respondWithError(w http.ResponseWriter, err error) {
	if _, ok := err.(*errors.NotFoundError); ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// cleanPath normalizes path for security and consistency.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	return filepath.ToSlash(p)
}

// containsPathTraversal checks for dangerous path patterns.
func containsPathTraversal(p string) bool {
	return strings.Contains(p, "..")
}
