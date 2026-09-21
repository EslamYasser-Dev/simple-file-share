package handlers

import (
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

type DownloadHandler struct {
	service *services.DownloadService
}

func NewDownloadHandler(service *services.DownloadService) *DownloadHandler {
	return &DownloadHandler{service: service}
}

func (h *DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	download, err := h.service.Execute(currentUser(r), pathFromQuery(r))
	if err != nil {
		respondWithError(w, err)
		return
	}
	serveDownload(w, download.Stream, download.Filename, download.ContentType)
}
