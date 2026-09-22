package handlers

import (
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// ViewHandler streams a file inline so the browser can render it (PDF, images,
// video, audio, plain text). The DownloadFileService decides whether the
// content type is safe to render; anything else is forced to download, so
// uploaded HTML/SVG can never execute on our origin.
type ViewHandler struct {
	fileService *services.DownloadFileService
}

func NewViewHandler(fileService *services.DownloadFileService) *ViewHandler {
	return &ViewHandler{fileService: fileService}
}

func (h *ViewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	download, err := h.fileService.Execute(currentUser(r), r.URL.Query().Get("path"))
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
