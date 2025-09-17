package handlers

import (
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// ListHandler handles requests to list directory contents.
type ListHandler struct {
	listService *services.ListFilesService
	port        string
}

// NewListHandler creates a new ListHandler.
func NewListHandler(listService *services.ListFilesService, port string) *ListHandler {
	return &ListHandler{
		listService: listService,
		port:        port,
	}
}

// ServeHTTP serves directory listing or delegates if not a directory.
func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := cleanPath(r.URL.Path)
	if containsPathTraversal(path) {
		http.Error(w, "Path traversal detected", http.StatusForbidden)
		return
	}

	pageData, err := h.listService.Execute(path)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if pageData == nil {
		// Not a directory — should be handled by download handler in router
		http.Error(w, "Not a directory", http.StatusNotFound)
		return
	}

	pageData.Port = h.port
	// if err := frontend.Tpl.Execute(w, pageData); err != nil {
	// 	http.Error(w, "Template render failed", http.StatusInternalServerError)
	// }
}
