package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

type DeleteHandler struct {
	deleteService *services.DeletePathService
}

func NewDeleteHandler(deleteService *services.DeletePathService) *DeleteHandler {
	return &DeleteHandler{deleteService: deleteService}
}

type deleteRequest struct {
	Path string `json:"path"`
}

func (h *DeleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req deleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.deleteService.Execute(currentUser(r), req.Path); err != nil {
		respondWithError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, dto.MessageResponse{Message: "deleted"})
}
