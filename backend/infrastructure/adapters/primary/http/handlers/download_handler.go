package handlers

import (
	"net/http"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

type DownloadHandler struct {
	fileService *services.DownloadFileService
	zipService  *services.DownloadZipService
}

func NewDownloadHandler(fileService *services.DownloadFileService, zipService *services.DownloadZipService) *DownloadHandler {
	return &DownloadHandler{
		fileService: fileService,
		zipService:  zipService,
	}
}

func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := pathFromQuery(r)
	if strings.HasSuffix(path, ".zip") {
		path = strings.TrimSuffix(path, ".zip")
		stream, filename, err := h.zipService.Execute(path)
		if err != nil {
			respondWithError(w, err)
			return
		}
		if stream == nil {
			respondJSON(w, http.StatusBadRequest, map[string]string{"error": "not a directory"})
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
		respondJSON(w, http.StatusConflict, map[string]string{"error": "path is a directory"})
		return
	}
	serveDownload(w, stream, filename, "application/octet-stream")
}
