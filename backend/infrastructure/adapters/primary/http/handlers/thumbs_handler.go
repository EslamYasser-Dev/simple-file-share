package handlers

import (
	"net/http"
	"strconv"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// ThumbsHandler serves JPEG previews on GET /api/thumbs?path=&w=. Only the
// caller's own namespace is reachable (same scoper as the view path), and
// re-encoding strips EXIF/GPS. Requires authentication.
type ThumbsHandler struct {
	svc *services.ThumbnailService
}

func NewThumbsHandler(svc *services.ThumbnailService) *ThumbsHandler {
	return &ThumbsHandler{svc: svc}
}

func (h *ThumbsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	width := 0
	if raw := q.Get("w"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			width = n
		}
	}
	thumb, err := h.svc.Execute(currentUser(r), q.Get("path"), width)
	if err != nil {
		respondWithError(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(thumb.Bytes)))
	w.Header().Set("Cache-Control", "private, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(thumb.Bytes)
}
