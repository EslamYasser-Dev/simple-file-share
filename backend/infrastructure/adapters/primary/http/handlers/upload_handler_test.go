package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

type uploadFixture struct {
	dir     string
	handler *UploadHandler
}

func newUploadFixture(t *testing.T) *uploadFixture {
	t.Helper()
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index)
	handler := NewUploadHandler(services.NewUploadService(fileRepo, policy.NewPathScoper(), index, fs.NewUserFileRepository(dir), 0))
	return &uploadFixture{dir: dir, handler: handler}
}

// multipartBody builds a multipart request body from name/content pairs.
// Non-file fields are written with CreateFormField.
func multipartBody(fields map[string]string, files map[string]string) (io.Reader, string, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	for name, value := range fields {
		if err := mw.WriteField(name, value); err != nil {
			return nil, "", err
		}
	}
	for name, content := range files {
		fw, err := mw.CreateFormFile("file", name)
		if err != nil {
			return nil, "", err
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			return nil, "", err
		}
	}
	if err := mw.Close(); err != nil {
		return nil, "", err
	}
	return &body, mw.FormDataContentType(), nil
}

func doUpload(t *testing.T, h *UploadHandler, body io.Reader, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/upload", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// Regression test: mime/multipart shares a single buffered reader across
// parts, so a later NextPart() call silently discards any unread bytes from
// the current part. The handler must drain each part before parsing the next,
// otherwise every upload lands as a 0-byte file.
func TestUploadSingleFilePersistsContent(t *testing.T) {
	f := newUploadFixture(t)
	body, ct, err := multipartBody(nil, map[string]string{"hello.txt": "hello world 12345\n"})
	if err != nil {
		t.Fatal(err)
	}

	rec := doUpload(t, f.handler, body, ct)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	result := firstUpload(t, rec)
	if result.Path != "hello.txt" || result.Size != int64(len("hello world 12345\n")) {
		t.Fatalf("unexpected result: %+v", result)
	}

	data, err := os.ReadFile(filepath.Join(f.dir, "hello.txt"))
	if err != nil {
		t.Fatalf("read back file: %v", err)
	}
	if string(data) != "hello world 12345\n" {
		t.Fatalf("content mismatch: got %q", string(data))
	}
}

func TestUploadMultipleFiles(t *testing.T) {
	f := newUploadFixture(t)
	files := map[string]string{
		"a.txt": "aaa",
		"b.txt": "bbbb",
		"c.txt": "cccccc",
	}
	body, ct, err := multipartBody(nil, files)
	if err != nil {
		t.Fatal(err)
	}

	rec := doUpload(t, f.handler, body, ct)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var results []dto.UploadResult
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(results) != len(files) {
		t.Fatalf("expected %d uploads, got %d (%+v)", len(files), len(results), results)
	}

	for name, content := range files {
		data, err := os.ReadFile(filepath.Join(f.dir, filepath.FromSlash(name)))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if string(data) != content {
			t.Fatalf("%s: got %q want %q", name, data, content)
		}
	}
}

func TestUploadWithPathField(t *testing.T) {
	f := newUploadFixture(t)
	body, ct, err := multipartBody(map[string]string{"path": "sub"}, map[string]string{"note.md": "# Note"})
	if err != nil {
		t.Fatal(err)
	}

	rec := doUpload(t, f.handler, body, ct)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	result := firstUpload(t, rec)
	want := filepath.ToSlash(filepath.Join("sub", "note.md"))
	if result.Path != want {
		t.Fatalf("path = %q, want %q", result.Path, want)
	}

	data, err := os.ReadFile(filepath.Join(f.dir, "sub", "note.md"))
	if err != nil {
		t.Fatalf("read nested file: %v", err)
	}
	if string(data) != "# Note" {
		t.Fatalf("content mismatch: %q", data)
	}
}

// RFC 7578 §4.2: mime/multipart strips directory information from a part's
// filename, so a hostile "../evil.txt" arrives here as the safe basename
// "evil.txt" — never escaping the root. This test pins that guarantee.
func TestUploadSanitizesDirectoryInFilename(t *testing.T) {
	f := newUploadFixture(t)
	body, ct, err := multipartBody(nil, map[string]string{"../../evil.txt": "pwned"})
	if err != nil {
		t.Fatal(err)
	}

	rec := doUpload(t, f.handler, body, ct)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	result := firstUpload(t, rec)
	if result.Path != "evil.txt" {
		t.Fatalf("path = %q, want sanitized basename evil.txt", result.Path)
	}

	// The file must exist inside the root and must not have escaped it.
	if _, err := os.Stat(filepath.Join(f.dir, "evil.txt")); err != nil {
		t.Fatalf("uploaded file missing inside root: %v", err)
	}
	escaped := filepath.Join(filepath.Dir(f.dir), "evil.txt")
	if _, err := os.Stat(escaped); !os.IsNotExist(err) {
		t.Fatalf("file escaped the root directory: %v", err)
	}
}

func TestUploadEmptyBody(t *testing.T) {
	f := newUploadFixture(t)
	rec := doUpload(t, f.handler, bytes.NewReader(nil), "multipart/form-data; boundary=xyz")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

// firstUpload decodes the (always-array) upload response and returns its first
// element, failing the test when the response is empty.
func firstUpload(t *testing.T, rec *httptest.ResponseRecorder) dto.UploadResult {
	t.Helper()
	var results []dto.UploadResult
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected at least one upload result, body = %s", rec.Body.String())
	}
	return results[0]
}
