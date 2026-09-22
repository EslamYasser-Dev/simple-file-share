package handlers

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// ViewHandler streams a file inline so the browser can render it (PDF, images,
// video, audio, plain text). Content types outside the allowlist are forced to
// download, preventing uploaded HTML/SVG from executing scripts on our origin.
type ViewHandler struct {
	fileService *services.DownloadFileService
}

func NewViewHandler(fileService *services.DownloadFileService) *ViewHandler {
	return &ViewHandler{fileService: fileService}
}

var inlineContentTypes = map[string]bool{
	"application/pdf":  true,
	"image/png":        true,
	"image/jpeg":       true,
	"image/gif":        true,
	"image/webp":       true,
	"image/bmp":        true,
	"text/plain":       true,
	"text/markdown":    true,
	"text/x-markdown":  true,
	"text/csv":         true,
	"application/json": true,
	"video/mp4":        true,
	"video/webm":       true,
	"video/ogg":        true,
	"video/quicktime":  true,
	"video/x-matroska": true,
	"video/x-msvideo":  true,
	"audio/mpeg":       true,
	"audio/ogg":        true,
	"audio/wav":        true,
	"audio/mp4":        true,
	"audio/webm":       true,
	"audio/x-m4a":      true,
}

func (h *ViewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := pathFromQuery(r)
	download, err := h.fileService.Execute(currentUser(r), path)
	if err != nil {
		respondWithError(w, err)
		return
	}

	contentType := contentTypeFor(download.Filename)
	if !inlineContentTypes[contentType] {
		serveDownload(w, download.Stream, download.Filename, "application/octet-stream")
		return
	}

	serveInline(w, download.Stream, download.Filename, contentType)
}

// contentTypeFor maps a filename to a media type, falling back to a generic
// binary type when the extension is unknown.
func contentTypeFor(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".pdf" {
		return "application/pdf"
	}
	if ct := mime.TypeByExtension(ext); ct != "" {
		if i := strings.IndexByte(ct, ';'); i >= 0 {
			ct = ct[:i]
		}
		return ct
	}
	return "application/octet-stream"
}
