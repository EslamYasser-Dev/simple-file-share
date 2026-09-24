package handlers

import (
	"bytes"
	"encoding/json"
	"io"
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

type versionFixture struct {
	dir      string
	list     *VersionsHandler
	download *VersionDownloadHandler
	restore  *VersionRestoreHandler
}

func newVersionHandlerFixture(t *testing.T) *versionFixture {
	t.Helper()
	dir := t.TempDir()
	local := fs.NewLocalFileRepository(dir)
	index := memory.NewFileIndexRepository()
	fileRepo := fs.NewIndexedFileRepository(local, index, nil)
	scoper := policy.NewPathScoper()

	for _, content := range []string{"v1", "v2"} {
		if _, err := local.WriteFile("doc.txt", io.NopCloser(strings.NewReader(content))); err != nil {
			t.Fatal(err)
		}
	}

	return &versionFixture{
		dir:      dir,
		list:     NewVersionsHandler(services.NewListVersionsService(fileRepo, local, scoper)),
		download: NewVersionDownloadHandler(services.NewDownloadVersionService(fileRepo, local, scoper)),
		restore:  NewVersionRestoreHandler(services.NewRestoreVersionService(fileRepo, local, scoper)),
	}
}

func TestVersionsListHandler(t *testing.T) {
	f := newVersionHandlerFixture(t)
	req := httptest.NewRequest(http.MethodGet, "/api/files/versions?path=doc.txt", nil)
	rec := httptest.NewRecorder()
	f.list.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var items []struct {
		Name    string `json:"name"`
		Version int    `json:"version"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(items) != 1 || items[0].Version != 1 || items[0].Name != "doc.txt" {
		t.Fatalf("unexpected versions: %+v", items)
	}
}

func TestVersionListNotFound(t *testing.T) {
	f := newVersionHandlerFixture(t)
	req := httptest.NewRequest(http.MethodGet, "/api/files/versions?path=missing.txt", nil)
	rec := httptest.NewRecorder()
	f.list.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestVersionDownloadHandler(t *testing.T) {
	f := newVersionHandlerFixture(t)
	req := httptest.NewRequest(http.MethodGet, "/api/files/version?path=doc.txt&n=1", nil)
	rec := httptest.NewRecorder()
	f.download.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "v1" {
		t.Fatalf("body = %q, want v1", rec.Body.String())
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "doc.txt") {
		t.Fatalf("disposition = %q, want doc.txt", cd)
	}
}

func TestVersionDownloadInvalidNumber(t *testing.T) {
	f := newVersionHandlerFixture(t)
	for _, n := range []string{"", "abc", "0", "-1"} {
		req := httptest.NewRequest(http.MethodGet, "/api/files/version?path=doc.txt&n="+n, nil)
		rec := httptest.NewRecorder()
		f.download.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("n=%q status = %d, want 400", n, rec.Code)
		}
	}
}

func TestVersionRestoreHandler(t *testing.T) {
	f := newVersionHandlerFixture(t)
	body, _ := json.Marshal(map[string]any{"path": "doc.txt", "n": 1})
	req := httptest.NewRequest(http.MethodPost, "/api/files/version/restore", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	f.restore.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	got, err := os.ReadFile(filepath.Join(f.dir, "doc.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "v1" {
		t.Fatalf("content = %q, want v1", got)
	}
}

func TestVersionRestoreInvalidBody(t *testing.T) {
	f := newVersionHandlerFixture(t)
	req := httptest.NewRequest(http.MethodPost, "/api/files/version/restore", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	f.restore.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
