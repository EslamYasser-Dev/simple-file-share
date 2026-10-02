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

type downloadFixture struct {
	dir     string
	handler *DownloadHandler
}

func newDownloadFixture(t *testing.T) *downloadFixture {
	t.Helper()
	dir := t.TempDir()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memory.NewFileIndexRepository(), nil)
	scoper := policy.NewPathScoper()
	svc := services.NewDownloadService(
		services.NewDownloadFileService(fileRepo, scoper),
		services.NewDownloadZipService(fileRepo, scoper),
	)
	return &downloadFixture{dir: dir, handler: NewDownloadHandler(svc)}
}

func (f *downloadFixture) write(t *testing.T, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(f.dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func downloadGet(t *testing.T, h http.Handler, path string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/files/download?path="+path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestDownloadRevalidatesWithETag(t *testing.T) {
	f := newDownloadFixture(t)
	f.write(t, "notes.txt", "hello world")

	rec := downloadGet(t, f.handler, "notes.txt", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET = %d body=%s", rec.Code, rec.Body)
	}
	if rec.Body.String() != "hello world" {
		t.Fatalf("body = %q", rec.Body.String())
	}
	etag := rec.Header().Get("ETag")
	if !strings.HasPrefix(etag, `"`) || !strings.HasSuffix(etag, `"`) {
		t.Fatalf("ETag = %q, want a quoted strong validator", etag)
	}
	if rec.Header().Get("Last-Modified") == "" {
		t.Fatal("missing Last-Modified")
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "attachment") {
		t.Fatalf("disposition = %q", cd)
	}

	// Matching validator -> 304 with no body and no disposition.
	cond := downloadGet(t, f.handler, "notes.txt", map[string]string{"If-None-Match": etag})
	if cond.Code != http.StatusNotModified {
		t.Fatalf("If-None-Match = %d, want 304", cond.Code)
	}
	if cond.Body.Len() != 0 {
		t.Fatalf("304 body = %q, want empty", cond.Body.String())
	}
	if cd := cond.Header().Get("Content-Disposition"); cd != "" {
		t.Fatalf("304 must not carry a disposition, got %q", cd)
	}

	// Star and weak forms of the same tag revalidate too.
	if r := downloadGet(t, f.handler, "notes.txt", map[string]string{"If-None-Match": "*"}); r.Code != http.StatusNotModified {
		t.Fatalf("star = %d, want 304", r.Code)
	}
	if r := downloadGet(t, f.handler, "notes.txt", map[string]string{"If-None-Match": "W/" + etag}); r.Code != http.StatusNotModified {
		t.Fatalf("weak = %d, want 304", r.Code)
	}
	// Stale validator falls through to a full body.
	stale := downloadGet(t, f.handler, "notes.txt", map[string]string{"If-None-Match": `"deadbeef"`})
	if stale.Code != http.StatusOK || stale.Body.String() != "hello world" {
		t.Fatalf("stale = %d %q", stale.Code, stale.Body.String())
	}
}

func TestDownloadSupportsRangeRequests(t *testing.T) {
	f := newDownloadFixture(t)
	f.write(t, "range.txt", "hello world")

	full := downloadGet(t, f.handler, "range.txt", nil)
	etag := full.Header().Get("ETag")

	rec := downloadGet(t, f.handler, "range.txt", map[string]string{"Range": "bytes=0-4"})
	if rec.Code != http.StatusPartialContent {
		t.Fatalf("Range = %d body=%s", rec.Code, rec.Body)
	}
	if rec.Body.String() != "hello" {
		t.Fatalf("partial body = %q, want hello", rec.Body.String())
	}
	if cr := rec.Header().Get("Content-Range"); cr != "bytes 0-4/11" {
		t.Fatalf("Content-Range = %q", cr)
	}
	if ar := rec.Header().Get("Accept-Ranges"); ar != "bytes" {
		t.Fatalf("Accept-Ranges = %q", ar)
	}

	// Unsatisfiable range -> 416.
	bad := downloadGet(t, f.handler, "range.txt", map[string]string{"Range": "bytes=999-1000"})
	if bad.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("unsatisfiable = %d, want 416", bad.Code)
	}

	// If-Range: matching validator honors the range, stale one falls back.
	matched := downloadGet(t, f.handler, "range.txt", map[string]string{
		"Range": "bytes=0-4", "If-Range": etag,
	})
	if matched.Code != http.StatusPartialContent || matched.Body.String() != "hello" {
		t.Fatalf("If-Range match = %d %q", matched.Code, matched.Body.String())
	}
	stale := downloadGet(t, f.handler, "range.txt", map[string]string{
		"Range": "bytes=0-4", "If-Range": `"stale"`,
	})
	if stale.Code != http.StatusOK || stale.Body.String() != "hello world" {
		t.Fatalf("If-Range stale = %d %q", stale.Code, stale.Body.String())
	}
}

func TestZipDownloadSkipsValidatorsAndServesFullBody(t *testing.T) {
	f := newDownloadFixture(t)
	if err := os.MkdirAll(filepath.Join(f.dir, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.dir, "docs", "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Zip archives are synthetic (no stable mtime/size identity): no ETag,
	// and a Range header must not corrupt the stream.
	rec := downloadGet(t, f.handler, "docs", map[string]string{"Range": "bytes=0-10"})
	if rec.Code != http.StatusOK {
		t.Fatalf("zip = %d, want 200", rec.Code)
	}
	if etag := rec.Header().Get("ETag"); etag != "" {
		t.Fatalf("zip ETag = %q, want none", etag)
	}
	if !strings.HasPrefix(rec.Body.String(), "PK") {
		t.Fatalf("body is not a zip: %q", rec.Body.String()[:min(4, rec.Body.Len())])
	}
}

func TestViewRevalidatesWithETag(t *testing.T) {
	dir := t.TempDir()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memory.NewFileIndexRepository(), nil)
	handler := NewViewHandler(services.NewDownloadFileService(fileRepo, policy.NewPathScoper()))
	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	get := func(headers map[string]string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/files/view?path=readme.md", nil)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	rec := get(nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET = %d", rec.Code)
	}
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("view response missing ETag")
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "inline") {
		t.Fatalf("disposition = %q", cd)
	}
	cond := get(map[string]string{"If-None-Match": etag})
	if cond.Code != http.StatusNotModified {
		t.Fatalf("If-None-Match = %d, want 304", cond.Code)
	}
}

func TestETagMatches(t *testing.T) {
	etag := `"abc"`
	cases := []struct {
		header string
		want   bool
	}{
		{"", false},
		{"*", true},
		{etag, true},
		{`W/` + etag, true},
		{`"other", ` + etag, true},
		{`"other"`, false},
		{etag[:len(etag)-1] + `x"`, false},
	}
	for _, c := range cases {
		if got := etagMatches(c.header, etag); got != c.want {
			t.Errorf("etagMatches(%q) = %v, want %v", c.header, got, c.want)
		}
	}
}
