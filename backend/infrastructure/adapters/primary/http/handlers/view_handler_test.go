package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

type viewFixture struct {
	dir     string
	handler *ViewHandler
}

func newViewFixture(t *testing.T) *viewFixture {
	t.Helper()
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index, nil)
	handler := NewViewHandler(services.NewDownloadFileService(fileRepo, policy.NewPathScoper()))
	return &viewFixture{dir: dir, handler: handler}
}

func (f *viewFixture) write(t *testing.T, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(f.dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func viewRequest(t *testing.T, h *ViewHandler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/files/view?path="+path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestViewInlineMarkdown(t *testing.T) {
	f := newViewFixture(t)
	content := "# Hello\n\nSome **bold** text.\n"
	f.write(t, "readme.md", content)

	rec := viewRequest(t, f.handler, "readme.md")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/markdown" {
		t.Fatalf("Content-Type = %q, want text/markdown", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "inline") {
		t.Fatalf("expected inline disposition, got %q", cd)
	}
	if rec.Body.String() != content {
		t.Fatalf("content mismatch: got %q", rec.Body.String())
	}
}

func TestViewForcesDownloadForUnsafeContentType(t *testing.T) {
	f := newViewFixture(t)
	f.write(t, "page.html", "<script>alert(1)</script>")

	rec := viewRequest(t, f.handler, "page.html")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html" {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "attachment") {
		t.Fatalf("expected attachment disposition, got %q", cd)
	}
}

func TestViewUnknownExtensionForcesDownload(t *testing.T) {
	f := newViewFixture(t)
	f.write(t, "data.bin", "\x00\x01\x02")

	rec := viewRequest(t, f.handler, "data.bin")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Fatalf("Content-Type = %q", ct)
	}
}

func TestViewDirectoryReturnsConflict(t *testing.T) {
	f := newViewFixture(t)
	if err := os.MkdirAll(filepath.Join(f.dir, "folder"), 0o755); err != nil {
		t.Fatal(err)
	}

	rec := viewRequest(t, f.handler, "folder")
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
}

func TestViewMissingFileReturnsNotFound(t *testing.T) {
	f := newViewFixture(t)
	rec := viewRequest(t, f.handler, "does-not-exist.txt")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestViewRejectsWrongMethod(t *testing.T) {
	f := newViewFixture(t)
	req := httptest.NewRequest(http.MethodPost, "/api/files/view?path=readme.md", nil)
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
