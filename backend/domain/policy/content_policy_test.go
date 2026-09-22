package policy

import "testing"

func TestContentTypeForKnownAndUnknownExtensions(t *testing.T) {
	p := ContentDispositionPolicy{}
	if got := p.ContentTypeFor("report.pdf"); got != "application/pdf" {
		t.Errorf("pdf → %q, want application/pdf", got)
	}
	if got := p.ContentTypeFor("photo.PNG"); got != "image/png" {
		t.Errorf("uppercase png → %q, want image/png", got)
	}
	if got := p.ContentTypeFor("noextension"); got != "application/octet-stream" {
		t.Errorf("no extension → %q, want application/octet-stream", got)
	}
}

func TestInlineContentTypes(t *testing.T) {
	p := ContentDispositionPolicy{}
	for _, safe := range []string{
		"application/pdf",
		"image/png",
		"text/plain",
		"video/mp4",
		"audio/mpeg",
	} {
		if !p.IsInlineContentType(safe) {
			t.Errorf("%q should be inline", safe)
		}
	}
	for _, unsafe := range []string{
		"text/html",
		"image/svg+xml",
		"application/zip",
		"application/javascript",
		"text/xml",
	} {
		if p.IsInlineContentType(unsafe) {
			t.Errorf("%q must not be inline", unsafe)
		}
	}
}
