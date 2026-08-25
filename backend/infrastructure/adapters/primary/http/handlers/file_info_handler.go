package handlers

import (
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

type FileInfoHandler struct {
	infoService *services.GetFileInfoService
}

func NewFileInfoHandler(infoService *services.GetFileInfoService) *FileInfoHandler {
	return &FileInfoHandler{infoService: infoService}
}

func (h *FileInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info, err := h.infoService.Execute(pathFromQuery(r))
	if err != nil {
		respondWithError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, dto.FromFileInfo(info))
}
