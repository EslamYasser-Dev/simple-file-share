package handlers

import (
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

type SearchHandler struct {
	searchService *services.SearchFilesService
}

func NewSearchHandler(searchService *services.SearchFilesService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	page, err := parsePaging(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid cursor")
		return
	}

	query := r.URL.Query().Get("q")
	results, err := h.searchService.Execute(currentUser(r), query, page.fetchSize())
	if err != nil {
		respondWithError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, window(page, dto.FromFileInfos(results)))
}
