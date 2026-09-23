package xhttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

type fakeAuthProvider struct {
	user *models.User
	err  error
}

func (f fakeAuthProvider) Authenticate(_, _ string) (*models.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.user, nil
}

var _ ports.AuthProvider = fakeAuthProvider{}

func testAuthMiddleware(provider ports.AuthProvider) func(http.Handler) http.Handler {
	return AuthMiddleware(services.NewAuthenticateService(provider), nil)
}

func TestAuthMiddlewareAcceptsBearer(t *testing.T) {
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	hash, _ := hasher.Hash("secret")
	_ = userRepo.CreateUser(&models.User{Username: "alice", PasswordHash: hash, CreatedAt: time.Now().UTC()})
	provider := auth.NewUserAuthProvider(userRepo, hasher)
	jwt := auth.NewJWTManager("mw-secret", time.Hour)
	tokens := services.NewTokenService(services.NewAuthenticateService(provider), jwt, userRepo)
	pair, err := tokens.Login("alice", "secret")
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := authctx.UserFromContext(r.Context())
		if got == nil || got.Username != "alice" {
			t.Errorf("context user = %+v, want alice", got)
		}
		w.WriteHeader(http.StatusOK)
	})
	handler := AuthMiddleware(services.NewAuthenticateService(provider), tokens)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	// Revoked token must be rejected.
	if err := tokens.Revoke(pair.AccessToken); err != nil {
		t.Fatal(err)
	}
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req3)
	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("revoked status = %d, want 401", rec2.Code)
	}
}

func TestAuthMiddlewareAttachesUser(t *testing.T) {
	alice := &models.User{Username: "alice", IsAdmin: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := authctx.UserFromContext(r.Context())
		if got == nil || got.Username != "alice" {
			t.Errorf("context user = %+v, want alice", got)
		}
		w.WriteHeader(http.StatusOK)
	})
	handler := testAuthMiddleware(fakeAuthProvider{user: alice})(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("alice", "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAuthMiddlewareAcceptsSessionCookie(t *testing.T) {
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	hash, _ := hasher.Hash("secret")
	_ = userRepo.CreateUser(&models.User{Username: "alice", PasswordHash: hash, CreatedAt: time.Now().UTC()})
	provider := auth.NewUserAuthProvider(userRepo, hasher)
	jwt := auth.NewJWTManager("mw-secret", time.Hour)
	tokens := services.NewTokenService(services.NewAuthenticateService(provider), jwt, userRepo)
	pair, err := tokens.Login("alice", "secret")
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := authctx.UserFromContext(r.Context())
		if got == nil || got.Username != "alice" {
			t.Errorf("context user = %+v, want alice", got)
		}
		w.WriteHeader(http.StatusOK)
	})
	handler := AuthMiddleware(services.NewAuthenticateService(provider), tokens)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: pair.AccessToken})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("cookie session status = %d, want 200", rec.Code)
	}

	// Revoked cookie token must be rejected.
	if err := tokens.Revoke(pair.AccessToken); err != nil {
		t.Fatal(err)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: SessionCookieName, Value: pair.AccessToken})
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Errorf("revoked cookie status = %d, want 401", rec2.Code)
	}
}

func TestAuthMiddlewareRejectsMissingCredentials(t *testing.T) {
	handler := testAuthMiddleware(fakeAuthProvider{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("next handler should not be called")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthMiddlewareRejectsBadCredentials(t *testing.T) {
	handler := testAuthMiddleware(fakeAuthProvider{err: errors.New("nope")})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("alice", "wrong")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthctxUserFromContextDefaultsNil(t *testing.T) {
	if u := authctx.UserFromContext(httptest.NewRequest(http.MethodGet, "/", nil).Context()); u != nil {
		t.Errorf("user = %+v, want nil", u)
	}
}

func TestCORSReflectsSameOriginWithCredentials(t *testing.T) {
	h := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	req.Host = "files.example"
	req.Header.Set("Origin", "https://files.example")
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://files.example" {
		t.Errorf("ACAO = %q, want reflected origin", got)
	}
	if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("expected Allow-Credentials true")
	}
	if rec.Header().Get("Access-Control-Allow-Origin") == "*" {
		t.Fatal("must never send ACAO:* with credentials")
	}
}

func TestCORSRejectsForeignOrigin(t *testing.T) {
	h := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	req.Host = "files.example"
	req.Header.Set("Origin", "https://evil.example")
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("ACAO = %q, want empty for foreign origin", got)
	}
}
