package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

type DirectoryHandler struct {
	createService *services.CreateDirectoryService
}

func NewDirectoryHandler(createService *services.CreateDirectoryService) *DirectoryHandler {
	return &DirectoryHandler{createService: createService}
}

type directoryRequest struct {
	Path string `json:"path"`
}

func (h *DirectoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req directoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.createService.Execute(req.Path); err != nil {
		respondWithError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"message": "directory created"})
}
