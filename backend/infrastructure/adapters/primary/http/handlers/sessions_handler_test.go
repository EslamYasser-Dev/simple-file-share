package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/sessions"
)

type sessionsFixture struct {
	handler *SessionsHandler
	svc     *services.SessionService
	jwt     ports.TokenManager
	store   ports.SessionRegistry
}

func newSessionsFixture(t *testing.T) sessionsFixture {
	t.Helper()
	jwt := auth.NewJWTManager("test-secret", time.Hour)
	store := sessions.NewRegistry(t.TempDir())
	svc := services.NewSessionService(store, jwt)
	return sessionsFixture{handler: NewSessionsHandler(svc), svc: svc, jwt: jwt, store: store}
}

func aliceCtx(r *http.Request) *http.Request {
	return r.WithContext(authctx.WithUser(r.Context(), &models.User{
		Username: "alice", Role: models.RoleMember, Enabled: true,
	}))
}

func TestSessionsListMarksCurrent(t *testing.T) {
	f := newSessionsFixture(t)
	token, claims, err := f.jwt.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}
	f.svc.Observe(&services.TokenPair{ID: claims.ID, Subject: "alice", ExpiresAt: claims.ExpiresAt}, "198.51.100.4", "cli")
	// A second, foreign session must not appear for alice.
	f.svc.Observe(&services.TokenPair{ID: "other", Subject: "bob", ExpiresAt: time.Now().Add(time.Hour)}, "", "")

	req := aliceCtx(httptest.NewRequest(http.MethodGet, "/api/auth/sessions", nil))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body)
	}

	var page struct {
		Items []struct {
			JTI     string `json:"jti"`
			Current bool   `json:"current"`
			Remote  string `json:"remote"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("response must be an object with items: %v (%s)", err, rec.Body)
	}
	if len(page.Items) != 1 {
		t.Fatalf("alice must see exactly her session, got %d", len(page.Items))
	}
	if page.Items[0].JTI != claims.ID || !page.Items[0].Current {
		t.Fatalf("current session wrong: %+v", page.Items[0])
	}
	if page.Items[0].Remote != "198.51.100.4" {
		t.Fatalf("remote lost: %+v", page.Items[0])
	}
}

func TestSessionsRevokeOwn(t *testing.T) {
	f := newSessionsFixture(t)
	_, claims, err := f.jwt.Issue("alice")
	if err != nil {
		t.Fatal(err)
	}
	f.svc.Observe(&services.TokenPair{ID: claims.ID, Subject: "alice", ExpiresAt: claims.ExpiresAt}, "", "")

	req := aliceCtx(httptest.NewRequest(http.MethodDelete, "/api/auth/sessions/"+claims.ID, nil))
	req.SetPathValue("jti", claims.ID)
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body)
	}
	if list, _ := f.store.List("alice"); len(list) != 0 {
		t.Fatalf("session survived revoke: %+v", list)
	}
}

func TestSessionsRevokeForeignIsForbidden(t *testing.T) {
	f := newSessionsFixture(t)
	f.svc.Observe(&services.TokenPair{ID: "bob-jti", Subject: "bob", ExpiresAt: time.Now().Add(time.Hour)}, "", "")

	req := aliceCtx(httptest.NewRequest(http.MethodDelete, "/api/auth/sessions/bob-jti", nil))
	req.SetPathValue("jti", "bob-jti")
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", rec.Code, rec.Body)
	}
	if list, _ := f.store.List("bob"); len(list) != 1 {
		t.Fatalf("foreign session must survive: %+v", list)
	}
}

func TestSessionsRevokeUnknownIsNotFound(t *testing.T) {
	f := newSessionsFixture(t)
	req := aliceCtx(httptest.NewRequest(http.MethodDelete, "/api/auth/sessions/nope", nil))
	req.SetPathValue("jti", "nope")
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", rec.Code, rec.Body)
	}
}

func TestSessionsMethodNotAllowedAndMissingJTI(t *testing.T) {
	f := newSessionsFixture(t)
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, aliceCtx(httptest.NewRequest(http.MethodPost, "/api/auth/sessions", nil)))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status = %d, want 405", rec.Code)
	}

	rec = httptest.NewRecorder()
	f.handler.ServeHTTP(rec, aliceCtx(httptest.NewRequest(http.MethodDelete, "/api/auth/sessions", nil)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing jti status = %d, want 400", rec.Code)
	}
}
