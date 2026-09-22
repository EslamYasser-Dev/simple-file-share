package policy

import (
	"mime"
	"path/filepath"
	"strings"
)

// ContentDispositionPolicy decides whether a downloaded file may be streamed
// inline to the browser. Only safe types (no active content) pass the
// allowlist; everything else must be forced to download so uploaded HTML/SVG
// can never execute on the server's origin.
type ContentDispositionPolicy struct{}

// ContentTypeFor maps a filename to a media type, falling back to a generic
// binary type when the extension is unknown.
func (ContentDispositionPolicy) ContentTypeFor(name string) string {
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

// IsInlineContentType reports whether a content type is safe to render inline.
func (ContentDispositionPolicy) IsInlineContentType(contentType string) bool {
	return inlineContentTypes[contentType]
}

// inlineContentTypes is the allowlist of media types the browser may render on
// our origin. Everything else is served as an attachment download.
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
