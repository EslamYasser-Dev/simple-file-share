package handlers

import (
	"net/http"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// ShareDownloadHandler serves a public share link at /api/share/{token}.
// Unlike every other file endpoint it needs no credentials: the token itself
// is the credential. Token structure is validated by the resolve use case, and
// the route is rate-limited to blunt brute-force retrieval.
type ShareDownloadHandler struct {
	service *services.ResolveShareService
}

func NewShareDownloadHandler(service *services.ResolveShareService) *ShareDownloadHandler {
	return &ShareDownloadHandler{service: service}
}

func (h *ShareDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/share/")
	download, err := h.service.Execute(token)
	if err != nil {
		respondWithError(w, err)
		return
	}

	if !download.Inline {
		serveDownload(w, download.Stream, download.Filename, download.ContentType)
		return
	}
	serveInline(w, download.Stream, download.Filename, download.ContentType)
}
