package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

func newRegisterHandler(t *testing.T, signupEnabled bool) (*RegisterHandler, *fs.UserFileRepository) {
	t.Helper()
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index, nil)
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	service := services.NewRegisterUserService(userRepo, hasher, fileRepo, policy.NewPathScoper(), signupEnabled, 0)
	return NewRegisterHandler(service), userRepo
}

func doRegister(t *testing.T, h *RegisterHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRegisterHandlerCreatesAccount(t *testing.T) {
	h, _ := newRegisterHandler(t, true)

	rec := doRegister(t, h, `{"username":"alice","password":"secret"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body %s)", rec.Code, rec.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["username"] != "alice" {
		t.Errorf("username = %v", got["username"])
	}
}

func TestRegisterHandlerConflict(t *testing.T) {
	h, _ := newRegisterHandler(t, true)
	if rec := doRegister(t, h, `{"username":"alice","password":"secret"}`); rec.Code != http.StatusCreated {
		t.Fatalf("seed registration status = %d", rec.Code)
	}
	if rec := doRegister(t, h, `{"username":"alice","password":"secret"}`); rec.Code != http.StatusConflict {
		t.Errorf("duplicate status = %d, want 409", rec.Code)
	}
}

func TestRegisterHandlerDisabled(t *testing.T) {
	h, _ := newRegisterHandler(t, false)
	if rec := doRegister(t, h, `{"username":"alice","password":"secret"}`); rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestMeHandlerWithoutUserIsSystemAdmin(t *testing.T) {
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	scoper := policy.NewPathScoper()
	userRepo := fs.NewUserFileRepository(dir)
	handler := NewMeHandler(services.NewUserInfoService(userRepo, index, scoper))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/auth/me", nil))

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["isAdmin"] != true {
		t.Errorf("isAdmin = %v, want true for system view", got["isAdmin"])
	}
}

func TestMeHandlerWithUser(t *testing.T) {
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	scoper := policy.NewPathScoper()
	userRepo := fs.NewUserFileRepository(dir)
	handler := NewMeHandler(services.NewUserInfoService(userRepo, index, scoper))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req = req.WithContext(authctx.WithUser(req.Context(), &models.User{Username: "bob"}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["username"] != "bob" || got["isAdmin"] != false {
		t.Errorf("response = %+v", got)
	}
}

func TestAdminUsersHandlerRequiresAdmin(t *testing.T) {
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	handler := NewAdminUsersHandler(services.NewListUsersService(userRepo, memory.NewFileIndexRepository(), policy.NewPathScoper()))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	req = req.WithContext(authctx.WithUser(req.Context(), &models.User{Username: "bob"}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestAuthInfoHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	NewAuthInfoHandler(true, []string{"github"}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/auth/info", nil))

	if rec.Code != http.StatusOK || rec.Body.String() == "" {
		t.Errorf("status = %d body = %q", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"oauth":["github"]`) {
		t.Errorf("body missing oauth list: %s", rec.Body.String())
	}
}
