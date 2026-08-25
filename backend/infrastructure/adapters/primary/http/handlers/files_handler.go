package handlers

import (
	"net/http"
)

// FilesHandler dispatches GET and DELETE on /api/files.
type FilesHandler struct {
	list   *ListHandler
	delete *DeleteHandler
}

func NewFilesHandler(list *ListHandler, delete *DeleteHandler) *FilesHandler {
	return &FilesHandler{list: list, delete: delete}
}

func (h *FilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list.ServeHTTP(w, r)
	case http.MethodDelete:
		h.delete.ServeHTTP(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
