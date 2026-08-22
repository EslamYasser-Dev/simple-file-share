package handlers

import (
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

type ListHandler struct {
	listService *services.ListFilesService
}

func NewListHandler(listService *services.ListFilesService) *ListHandler {
	return &ListHandler{listService: listService}
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pageData, err := h.listService.Execute(pathFromQuery(r))
	if err != nil {
		respondWithError(w, err)
		return
	}
	if pageData == nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "not a directory"})
		return
	}

	respondJSON(w, http.StatusOK, dto.FromFileInfos(pageData.Files))
}
