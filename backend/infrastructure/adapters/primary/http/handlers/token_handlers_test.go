package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

func newTokenHandlers(t *testing.T) (*TokenHandler, *RefreshHandler, *RevokeHandler, *fs.UserFileRepository) {
	t.Helper()
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	provider := auth.NewUserAuthProvider(userRepo, hasher)
	jwt := auth.NewJWTManager("test-secret", time.Hour)
	tokens := services.NewTokenService(services.NewAuthenticateService(provider), jwt, userRepo)
	return NewTokenHandler(tokens), NewRefreshHandler(tokens), NewRevokeHandler(tokens), userRepo
}

func TestTokenHandlerIssuesAndRefreshes(t *testing.T) {
	tokenH, refreshH, revokeH, users := newTokenHandlers(t)
	hash, _ := auth.NewPBKDF2Hasher().Hash("secret")
	_ = users.CreateUser(&models.User{Username: "alice", PasswordHash: hash, Enabled: true, CreatedAt: time.Now().UTC()})

	rec := httptest.NewRecorder()
	tokenH.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/auth/token", strings.NewReader(`{"username":"alice","password":"secret"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("token status = %d body = %s", rec.Code, rec.Body.String())
	}
	var issued struct {
		AccessToken string `json:"accessToken"`
		TokenType   string `json:"tokenType"`
		ExpiresIn   int64  `json:"expiresIn"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &issued); err != nil {
		t.Fatal(err)
	}
	if issued.AccessToken == "" || issued.TokenType != "Bearer" || issued.ExpiresIn <= 0 {
		t.Fatalf("issued = %+v", issued)
	}
	if c := sessionCookie(rec); c == nil || c.Value == "" || !c.HttpOnly {
		t.Fatalf("token response must set HttpOnly fs_session cookie, got %#v", c)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", strings.NewReader(`{"accessToken":"`+issued.AccessToken+`"}`))
	rec2 := httptest.NewRecorder()
	refreshH.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Fatalf("refresh status = %d body = %s", rec2.Code, rec2.Body.String())
	}
	var refreshed struct {
		AccessToken string `json:"accessToken"`
	}
	_ = json.Unmarshal(rec2.Body.Bytes(), &refreshed)
	if refreshed.AccessToken == "" || refreshed.AccessToken == issued.AccessToken {
		t.Fatalf("refreshed = %+v", refreshed)
	}

	rec3 := httptest.NewRecorder()
	revokeH.ServeHTTP(rec3, httptest.NewRequest(http.MethodPost, "/api/auth/revoke", strings.NewReader(`{"accessToken":"`+refreshed.AccessToken+`"}`)))
	if rec3.Code != http.StatusOK {
		t.Fatalf("revoke status = %d body = %s", rec3.Code, rec3.Body.String())
	}
}

func TestTokenHandlerRejectsBadPassword(t *testing.T) {
	tokenH, _, _, users := newTokenHandlers(t)
	hash, _ := auth.NewPBKDF2Hasher().Hash("secret")
	_ = users.CreateUser(&models.User{Username: "alice", PasswordHash: hash, Enabled: true, CreatedAt: time.Now().UTC()})

	rec := httptest.NewRecorder()
	tokenH.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/auth/token", strings.NewReader(`{"username":"alice","password":"nope"}`)))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRevokeClearsSessionCookie(t *testing.T) {
	_, _, revokeH, _ := newTokenHandlers(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/revoke", nil)
	revokeH.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if c := sessionCookie(rec); c == nil || c.Value != "" || c.MaxAge >= 0 {
		t.Fatalf("revoke must clear fs_session, got %#v", c)
	}
}

func sessionCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, raw := range rec.Header().Values("Set-Cookie") {
		if !strings.HasPrefix(raw, "fs_session=") {
			continue
		}
		header := http.Header{}
		header.Add("Set-Cookie", raw)
		resp := http.Response{Header: header}
		for _, c := range resp.Cookies() {
			if c.Name == "fs_session" {
				return c
			}
		}
	}
	return nil
}
