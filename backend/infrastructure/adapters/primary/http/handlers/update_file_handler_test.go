package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

type updateFixture struct {
	dir     string
	handler *UpdateFileHandler
}

func newUpdateFixture(t *testing.T) *updateFixture {
	t.Helper()
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index, nil)
	handler := NewUpdateFileHandler(services.NewUpdateFileContentService(fileRepo, policy.NewPathScoper()))
	return &updateFixture{dir: dir, handler: handler}
}

func doUpdate(t *testing.T, h *UpdateFileHandler, path, content string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"path": path, "content": content})
	req := httptest.NewRequest(http.MethodPut, "/api/files/content", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestUpdateFileOverwritesContent(t *testing.T) {
	f := newUpdateFixture(t)
	if err := os.WriteFile(filepath.Join(f.dir, "doc.md"), []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	rec := doUpdate(t, f.handler, "doc.md", "# New content")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var result struct {
		Size int64 `json:"size"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &result)
	if result.Size != int64(len("# New content")) {
		t.Fatalf("size = %d, want %d", result.Size, len("# New content"))
	}

	data, err := os.ReadFile(filepath.Join(f.dir, "doc.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# New content" {
		t.Fatalf("content mismatch: %q", data)
	}
}

func TestUpdateFileNotFound(t *testing.T) {
	f := newUpdateFixture(t)
	rec := doUpdate(t, f.handler, "missing.md", "x")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestUpdateDirectoryRejected(t *testing.T) {
	f := newUpdateFixture(t)
	if err := os.MkdirAll(filepath.Join(f.dir, "folder"), 0o755); err != nil {
		t.Fatal(err)
	}

	rec := doUpdate(t, f.handler, "folder", "x")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestUpdatePathTraversalRejected(t *testing.T) {
	f := newUpdateFixture(t)
	rec := doUpdate(t, f.handler, "../outside.md", "x")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestUpdateRequiresPath(t *testing.T) {
	f := newUpdateFixture(t)
	rec := doUpdate(t, f.handler, "", "x")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
