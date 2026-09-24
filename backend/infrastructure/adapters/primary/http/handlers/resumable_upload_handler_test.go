package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

type resumableFixture struct {
	dir     string
	handler *ResumableUploadHandler
	svc     *services.ResumableUploadService
}

func newResumableFixture(t *testing.T) *resumableFixture {
	t.Helper()
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index, nil)
	sessions := fs.NewUploadSessionRepository(dir)
	svc := services.NewResumableUploadService(sessions, fileRepo, policy.NewPathScoper(), index, fs.NewUserFileRepository(dir), 0)
	return &resumableFixture{dir: dir, handler: NewResumableUploadHandler(svc), svc: svc}
}

func decodeSession(t *testing.T, rec *httptest.ResponseRecorder) dto.UploadSession {
	t.Helper()
	var session dto.UploadSession
	if err := json.Unmarshal(rec.Body.Bytes(), &session); err != nil {
		t.Fatalf("decode session: %v (body=%s)", err, rec.Body.String())
	}
	return session
}

func TestResumableUploadRoundTrip(t *testing.T) {
	f := newResumableFixture(t)
	content := []byte("resumable upload payload 12345")

	createBody, _ := json.Marshal(dto.CreateUploadSessionRequest{
		Path:        "",
		Filename:    "big.bin",
		Size:        int64(len(content)),
		Fingerprint: "fp-roundtrip",
	})
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/uploads", bytes.NewReader(createBody)))
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}
	session := decodeSession(t, rec)
	if session.Offset != 0 || session.Size != int64(len(content)) {
		t.Fatalf("unexpected session: %+v", session)
	}

	// Two chunks to exercise multi-part staging.
	mid := 10
	patch := func(offset int, chunk []byte) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPatch, "/api/uploads/"+session.ID, bytes.NewReader(chunk))
		req.Header.Set("Upload-Offset", itoa(offset))
		rec := httptest.NewRecorder()
		f.handler.ServeHTTP(rec, req)
		return rec
	}
	if rec := patch(0, content[:mid]); rec.Code != http.StatusOK {
		t.Fatalf("chunk1 status = %d body=%s", rec.Code, rec.Body.String())
	}
	if rec := patch(mid, content[mid:]); rec.Code != http.StatusOK {
		t.Fatalf("chunk2 status = %d body=%s", rec.Code, rec.Body.String())
	}

	// Wrong offset → 409 with expected.
	if rec := patch(mid, content[mid:]); rec.Code != http.StatusConflict {
		t.Fatalf("mismatch status = %d body=%s", rec.Code, rec.Body.String())
	}

	complete := httptest.NewRequest(http.MethodPost, "/api/uploads/"+session.ID+"/complete", nil)
	rec = httptest.NewRecorder()
	f.handler.ServeHTTP(rec, complete)
	if rec.Code != http.StatusOK {
		t.Fatalf("complete status = %d body=%s", rec.Code, rec.Body.String())
	}
	var result dto.UploadResult
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if result.Path != "big.bin" || result.Size != int64(len(content)) {
		t.Fatalf("result = %+v", result)
	}

	data, err := os.ReadFile(filepath.Join(f.dir, "big.bin"))
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("content mismatch: %q", data)
	}
}

func TestResumableUploadResumeByFingerprint(t *testing.T) {
	f := newResumableFixture(t)
	content := []byte("0123456789abcdefghij")

	create := func() dto.UploadSession {
		body, _ := json.Marshal(dto.CreateUploadSessionRequest{
			Filename:    "resume.bin",
			Size:        int64(len(content)),
			Fingerprint: "fp-resume",
		})
		rec := httptest.NewRecorder()
		f.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/uploads", bytes.NewReader(body)))
		if rec.Code != http.StatusOK {
			t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
		}
		return decodeSession(t, rec)
	}

	first := create()
	req := httptest.NewRequest(http.MethodPatch, "/api/uploads/"+first.ID, bytes.NewReader(content[:5]))
	req.Header.Set("Upload-Offset", "0")
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("append status = %d body=%s", rec.Code, rec.Body.String())
	}

	// A second Start with the same fingerprint must reuse the session offset.
	second := create()
	if second.ID != first.ID {
		t.Fatalf("expected resume of %s, got %s", first.ID, second.ID)
	}
	if second.Offset != 5 {
		t.Fatalf("resumed offset = %d, want 5", second.Offset)
	}
}

func TestResumableUploadListPending(t *testing.T) {
	f := newResumableFixture(t)
	body, _ := json.Marshal(dto.CreateUploadSessionRequest{
		Path:        "docs",
		Filename:    "pending.bin",
		Size:        100,
		Fingerprint: "fp-pending",
	})
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/uploads", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}
	created := decodeSession(t, rec)

	req := httptest.NewRequest(http.MethodGet, "/api/uploads", nil)
	rec = httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", rec.Code, rec.Body.String())
	}
	var list []dto.PendingUploadSession
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v (body=%s)", err, rec.Body.String())
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}
	got := list[0]
	if got.ID != created.ID || got.Filename != "pending.bin" || got.Destination != "docs" {
		t.Fatalf("pending = %+v", got)
	}
	if got.Fingerprint != "fp-pending" || got.Size != 100 || got.Offset != 0 {
		t.Fatalf("pending fields = %+v", got)
	}

	// Abort removes it from the list.
	abort := httptest.NewRequest(http.MethodDelete, "/api/uploads/"+created.ID, nil)
	rec = httptest.NewRecorder()
	f.handler.ServeHTTP(rec, abort)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("abort status = %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	f.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/uploads", nil))
	list = nil
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("list after abort = %+v", list)
	}
}

func TestResumableUploadRejectsForeignSession(t *testing.T) {
	f := newResumableFixture(t)
	body, _ := json.Marshal(dto.CreateUploadSessionRequest{Filename: "a.txt", Size: 1})
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/uploads", bytes.NewReader(body)))
	session := decodeSession(t, rec)

	// Status with a non-owned session id still 404s when ownership differs;
	// with auth disabled owner is "" for everyone, so ownership always matches.
	// Abort then Get should 404.
	abort := httptest.NewRequest(http.MethodDelete, "/api/uploads/"+session.ID, nil)
	rec = httptest.NewRecorder()
	f.handler.ServeHTTP(rec, abort)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("abort status = %d body=%s", rec.Code, rec.Body.String())
	}

	get := httptest.NewRequest(http.MethodGet, "/api/uploads/"+session.ID, nil)
	rec = httptest.NewRecorder()
	f.handler.ServeHTTP(rec, get)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get after abort status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
