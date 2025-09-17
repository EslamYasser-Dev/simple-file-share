package handlers

import (
    "net/http"
    "strings"

    "github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// RootHandler decides between directory listing and file/zip download for GET "/".
type RootHandler struct {
	listService     *services.ListFilesService
	fileService     *services.DownloadFileService
	zipService      *services.DownloadZipService
	port            string
}

func NewRootHandler(list *services.ListFilesService, file *services.DownloadFileService, zip *services.DownloadZipService, port string) *RootHandler {
	return &RootHandler{listService: list, fileService: file, zipService: zip, port: port}
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := cleanPath(r.URL.Path)
	if containsPathTraversal(path) {
		http.Error(w, "Path traversal detected", http.StatusForbidden)
		return
	}

	// ZIP request
	if strings.HasSuffix(path, ".zip") {
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

	// Try as directory first
	pageData, err := h.listService.Execute(path)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if pageData != nil {
		pageData.Port = h.port
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Directory listing available via API. Integrate template to render UI."))
		return
	}

	// Fallback: serve as file
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
