package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// VersionsHandler lists a file's historical snapshots: GET /api/files/versions.
type VersionsHandler struct {
	listService *services.ListVersionsService
}

func NewVersionsHandler(listService *services.ListVersionsService) *VersionsHandler {
	return &VersionsHandler{listService: listService}
}

func (h *VersionsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	versions, err := h.listService.Execute(currentUser(r), r.URL.Query().Get("path"))
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dto.FromFileInfos(versions))
}

// VersionDownloadHandler streams a single snapshot: GET /api/files/version.
type VersionDownloadHandler struct {
	service *services.DownloadVersionService
}

func NewVersionDownloadHandler(service *services.DownloadVersionService) *VersionDownloadHandler {
	return &VersionDownloadHandler{service: service}
}

func (h *VersionDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	n, err := strconv.Atoi(r.URL.Query().Get("n"))
	if err != nil || n < 1 {
		respondError(w, http.StatusBadRequest, "invalid version number")
		return
	}

	download, err := h.service.Execute(currentUser(r), r.URL.Query().Get("path"), n)
	if err != nil {
		respondWithError(w, err)
		return
	}

	if !download.Inline {
		serveDownload(w, download.Stream, download.Filename, download.ContentType)
		return
	}
	serveInline(w, download.Stream, download.Filename, download.ContentType)
}

// VersionRestoreHandler rolls the file back to a snapshot:
// POST /api/files/version/restore.
type VersionRestoreHandler struct {
	service *services.RestoreVersionService
}

func NewVersionRestoreHandler(service *services.RestoreVersionService) *VersionRestoreHandler {
	return &VersionRestoreHandler{service: service}
}

func (h *VersionRestoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.RestoreVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Execute(currentUser(r), req.Path, req.N); err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dto.MessageResponse{Message: "restored"})
}
