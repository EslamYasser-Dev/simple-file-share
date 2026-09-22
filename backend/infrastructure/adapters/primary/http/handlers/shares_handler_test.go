package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

type shareHandlerFixture struct {
	shares *SharesHandler
	share  *ShareDownloadHandler
	repo   *fs.ShareFileRepository
}

func newShareHandlerFixture(t *testing.T) *shareHandlerFixture {
	t.Helper()
	dir := t.TempDir()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memory.NewFileIndexRepository())
	shareRepo := fs.NewShareFileRepository(dir)
	scoper := policy.NewPathScoper()

	if _, err := fileRepo.WriteFile("hello.txt", io.NopCloser(strings.NewReader("sup"))); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}

	downloadService := services.NewDownloadService(
		services.NewDownloadFileService(fileRepo, scoper),
		services.NewDownloadZipService(fileRepo, scoper),
	)
	return &shareHandlerFixture{
		shares: NewSharesHandler(
			services.NewCreateShareService(fileRepo, shareRepo, scoper),
			services.NewListSharesService(shareRepo, scoper),
			services.NewRevokeShareService(shareRepo, scoper),
		),
		share: NewShareDownloadHandler(services.NewResolveShareService(shareRepo, scoper, downloadService)),
		repo:  shareRepo,
	}
}

func doJSON(t *testing.T, h http.Handler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// decodeShares unmarshals a share list response body.
func decodeShares(t *testing.T, rec *httptest.ResponseRecorder) []dto.ShareItem {
	t.Helper()
	var items []dto.ShareItem
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}
	return items
}

func TestSharesHandlerCreateListRevoke(t *testing.T) {
	f := newShareHandlerFixture(t)

	// Creating a link for a path that does not exist is a 404.
	rec := doJSON(t, f.shares, http.MethodPost, "/api/shares", `{"path":"/nope.txt","expiresInSeconds":3600}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("create missing path = %d, want 404", rec.Code)
	}

	// Negative or excessive validity is rejected as a validation error.
	rec = doJSON(t, f.shares, http.MethodPost, "/api/shares", `{"path":"/hello.txt","expiresInSeconds":-5}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create negative expiry = %d, want 400", rec.Code)
	}
	rec = doJSON(t, f.shares, http.MethodPost, "/api/shares", `{"path":"/hello.txt","expiresInSeconds":400000000}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create excessive expiry = %d, want 400", rec.Code)
	}

	// A valid link is created, listed, then revoked.
	rec = doJSON(t, f.shares, http.MethodPost, "/api/shares", `{"path":"/hello.txt","expiresInSeconds":3600}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create = %d, want 201 (body %s)", rec.Code, rec.Body.String())
	}
	var created dto.ShareItem
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.Token == "" || created.ExpiresAt == "" || created.Name != "hello.txt" {
		t.Errorf("created = %+v", created)
	}

	listed := decodeShares(t, doJSON(t, f.shares, http.MethodGet, "/api/shares", ""))
	if len(listed) != 1 || listed[0].Token != created.Token {
		t.Fatalf("listed = %+v, want the created share", listed)
	}

	// Deleting without (or with a malformed) token is the same as not found.
	rec = doJSON(t, f.shares, http.MethodDelete, "/api/shares", `{}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("revoke without token = %d, want 404", rec.Code)
	}

	rec = doJSON(t, f.shares, http.MethodDelete, "/api/shares", `{"token":"`+created.Token+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("revoke = %d, want 200", rec.Code)
	}
	listed = decodeShares(t, doJSON(t, f.shares, http.MethodGet, "/api/shares", ""))
	if len(listed) != 0 {
		t.Fatalf("listed after revoke = %+v, want empty", listed)
	}
}

func TestSharesHandlerRejectsOtherMethods(t *testing.T) {
	f := newShareHandlerFixture(t)
	rec := doJSON(t, f.shares, http.MethodPatch, "/api/shares", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("PATCH = %d, want 405", rec.Code)
	}
}

func TestShareDownloadHandler(t *testing.T) {
	f := newShareHandlerFixture(t)

	// Happy path: a public fetch of a valid token streams the file body.
	rec := doJSON(t, f.shares, http.MethodPost, "/api/shares", `{"path":"/hello.txt","expiresInSeconds":60}`)
	var created dto.ShareItem
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/share/"+created.Token, nil)
	w := httptest.NewRecorder()
	f.share.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("download = %d, want 200", w.Code)
	}
	if w.Body.String() != "sup" {
		t.Errorf("body = %q, want sup", w.Body.String())
	}

	// Unknown token -> 404.
	req = httptest.NewRequest(http.MethodGet, "/api/share/NotARealToken123", nil)
	w = httptest.NewRecorder()
	f.share.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown token = %d, want 404", w.Code)
	}

	// A crafted token containing traversal segments never reaches a repository.
	req = httptest.NewRequest(http.MethodGet, "/api/share/..%2F..%2Fetc%2Fpasswd", nil)
	w = httptest.NewRecorder()
	f.share.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("malformed token = %d, want 404", w.Code)
	}

	// Non-GET methods are refused outright.
	req = httptest.NewRequest(http.MethodPost, "/api/share/"+created.Token, nil)
	w = httptest.NewRecorder()
	f.share.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST = %d, want 405", w.Code)
	}

	// An already-expired share answers 410 Gone while the file still exists.
	expired, err := f.repo.ListAll()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(expired) == 0 {
		t.Fatal("expected at least one share to still exist")
	}
	if err := f.repo.Create(&models.Share{
		Token:     "expired-token",
		Path:      "hello.txt",
		Owner:     "",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour),
	}); err != nil {
		t.Fatalf("seed expired share: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/share/expired-token", nil)
	w = httptest.NewRecorder()
	f.share.ServeHTTP(w, req)
	if w.Code != http.StatusGone {
		t.Fatalf("expired share = %d, want 410", w.Code)
	}
}
