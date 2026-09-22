package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

type DirectoryHandler struct {
	createService *services.CreateDirectoryService
}

func NewDirectoryHandler(createService *services.CreateDirectoryService) *DirectoryHandler {
	return &DirectoryHandler{createService: createService}
}

func (h *DirectoryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreateDirectoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.createService.Execute(currentUser(r), req.Path); err != nil {
		respondWithError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, dto.MessageResponse{Message: "directory created"})
}
