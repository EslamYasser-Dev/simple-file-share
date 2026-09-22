package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// UpdateFileHandler overwrites a text file's contents (used by the markdown
// editor in the web UI). Only plain-text content is accepted.
type UpdateFileHandler struct {
	updateService *services.UpdateFileContentService
}

func NewUpdateFileHandler(updateService *services.UpdateFileContentService) *UpdateFileHandler {
	return &UpdateFileHandler{updateService: updateService}
}

func (h *UpdateFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.UpdateContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	size, err := h.updateService.Execute(currentUser(r), req.Path, req.Content)
	if err != nil {
		respondWithError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, dto.UpdateFileResponse{Message: "file updated", Size: size})
}
