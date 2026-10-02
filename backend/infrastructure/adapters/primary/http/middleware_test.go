package xhttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
	return AuthMiddleware(services.NewAuthenticateService(provider), nil, nil)
}

func TestAuthMiddlewareAcceptsBearer(t *testing.T) {
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	hash, _ := hasher.Hash("secret")
	_ = userRepo.CreateUser(&models.User{Username: "alice", PasswordHash: hash, Enabled: true, CreatedAt: time.Now().UTC()})
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
	handler := AuthMiddleware(services.NewAuthenticateService(provider), tokens, nil)(next)

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
	_ = userRepo.CreateUser(&models.User{Username: "alice", PasswordHash: hash, Enabled: true, CreatedAt: time.Now().UTC()})
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
	handler := AuthMiddleware(services.NewAuthenticateService(provider), tokens, nil)(next)

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

// Basic auth has no OTP channel: an account with TOTP enabled must get the
// totp_required challenge instead of a plain 401, and must succeed only via
// the token endpoint.
func TestAuthMiddlewareBasicRejectsTwoFactorAccount(t *testing.T) {
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	hash, _ := hasher.Hash("secret")
	if err := userRepo.CreateUser(&models.User{Username: "alice", PasswordHash: hash, Role: models.RoleMember, Enabled: true, CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if err := userRepo.CreateUser(&models.User{Username: "bob", PasswordHash: hash, Role: models.RoleMember, Enabled: true, CreatedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	provider := auth.NewUserAuthProvider(userRepo, hasher)
	authSvc := services.NewAuthenticateService(provider)
	jwt := auth.NewJWTManager("mw-secret", time.Hour)
	tokens := services.NewTokenService(authSvc, jwt, userRepo)
	totp := services.NewTOTPService(userRepo, hasher, services.NewRoleCatalog(nil))
	tokens.SetTwoFactor(totp)

	alice, err := userRepo.FindByUsername("alice")
	if err != nil {
		t.Fatal(err)
	}
	secret, _, err := totp.Enroll(alice)
	if err != nil {
		t.Fatal(err)
	}
	code, _ := auth.TOTPCode(secret, time.Now())
	if _, err := totp.Confirm(alice, code); err != nil {
		t.Fatal(err)
	}

	mw := AuthMiddleware(authSvc, tokens, nil)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u := authctx.UserFromContext(r.Context()); u == nil {
			t.Error("handler must see an authenticated user")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Basic + enrolled → JSON totp_required challenge.
	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	req.SetBasicAuth("alice", "secret")
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("basic enrolled status = %d, want 401; body=%s", rec.Code, rec.Body)
	}
	if body := rec.Body.String(); !strings.Contains(body, "totp_required") {
		t.Fatalf("body = %s, want totp_required", body)
	}

	// Basic + plain account still works.
	req = httptest.NewRequest(http.MethodGet, "/api/files", nil)
	req.SetBasicAuth("bob", "secret")
	rec = httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("basic plain status = %d, want 200; body=%s", rec.Code, rec.Body)
	}

	// Token issued after OTP login passes as Bearer.
	fresh, _ := auth.TOTPCode(secret, time.Now())
	pair, err := tokens.LoginWithOTP("alice", "secret", fresh)
	if err != nil {
		t.Fatalf("otp login: %v", err)
	}
	req = httptest.NewRequest(http.MethodGet, "/api/files", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec = httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("bearer after otp status = %d, want 200; body=%s", rec.Code, rec.Body)
	}
}

// API keys authenticate as their owner but only within their scope:
// read = safe methods, write = mutations, admin = /api/admin/*, and the
// credential surface stays off limits entirely (except /api/auth/me).
func TestAuthMiddlewareAPIKeyScopes(t *testing.T) {
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	hash, _ := hasher.Hash("secret")
	alice := &models.User{Username: "alice", PasswordHash: hash, Role: models.RoleMember, Enabled: true, CreatedAt: time.Now().UTC()}
	if err := userRepo.CreateUser(alice); err != nil {
		t.Fatal(err)
	}
	provider := auth.NewUserAuthProvider(userRepo, hasher)
	authSvc := services.NewAuthenticateService(provider)
	jwt := auth.NewJWTManager("mw-secret", time.Hour)
	tokens := services.NewTokenService(authSvc, jwt, userRepo)
	apiKeys := services.NewAPIKeyService(fs.NewAPIKeyFileRepository(dir), userRepo, hasher)

	mint := func(scope string) string {
		t.Helper()
		plaintext, _, err := apiKeys.Issue(alice, "k", scope, 0)
		if err != nil {
			t.Fatal(err)
		}
		return plaintext
	}
	readKey := mint(models.APIKeyScopeRead)
	writeKey := mint(models.APIKeyScopeWrite)
	adminKey := mint(models.APIKeyScopeAdmin)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u := authctx.UserFromContext(r.Context()); u == nil {
			t.Error("key-authenticated request must carry the owner")
		}
		w.WriteHeader(http.StatusOK)
	})
	mw := AuthMiddleware(authSvc, tokens, apiKeys)
	status := func(key, method, path string) int {
		t.Helper()
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Authorization", "Bearer "+key)
		rec := httptest.NewRecorder()
		mw(next).ServeHTTP(rec, req)
		return rec.Code
	}

	for _, tc := range []struct {
		name, key, method, path string
		want                    int
	}{
		{"read GET file", readKey, http.MethodGet, "/api/files", http.StatusOK},
		{"read GET share list", readKey, http.MethodGet, "/api/shares", http.StatusOK},
		{"read cannot POST", readKey, http.MethodPost, "/api/files", http.StatusForbidden},
		{"read cannot DELETE", readKey, http.MethodDelete, "/api/shares", http.StatusForbidden},
		{"read cannot read admin", readKey, http.MethodGet, "/api/admin/users", http.StatusForbidden},
		{"write can POST", writeKey, http.MethodPost, "/api/shares", http.StatusOK},
		{"write can PUT", writeKey, http.MethodPut, "/api/files/content", http.StatusOK},
		{"write cannot admin GET", writeKey, http.MethodGet, "/api/admin/users", http.StatusForbidden},
		{"write cannot admin POST", writeKey, http.MethodPost, "/api/admin/users", http.StatusForbidden},
		{"admin can admin GET", adminKey, http.MethodGet, "/api/admin/users", http.StatusOK},
		{"admin can admin DELETE", adminKey, http.MethodDelete, "/api/admin/users/bob", http.StatusOK},
		{"key blocked from key management", adminKey, http.MethodDelete, "/api/auth/api-keys/x", http.StatusForbidden},
		{"key blocked from password change", writeKey, http.MethodPost, "/api/auth/password", http.StatusForbidden},
		{"key blocked from sessions", writeKey, http.MethodGet, "/api/auth/sessions", http.StatusForbidden},
		{"key allowed to verify identity", readKey, http.MethodGet, "/api/auth/me", http.StatusOK},
		{"unknown key rejected", services.APIKeyPrefix + "ffffffffffffffff_" + strings.Repeat("z", 32), http.MethodGet, "/api/files", http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := status(tc.key, tc.method, tc.path); got != tc.want {
				t.Fatalf("%s %s = %d, want %d", tc.method, tc.path, got, tc.want)
			}
		})
	}

	// Disabled feature: keys stop authenticating entirely.
	disabled := AuthMiddleware(authSvc, tokens, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	req.Header.Set("Authorization", "Bearer "+readKey)
	rec := httptest.NewRecorder()
	disabled(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("keys disabled: %d, want 401", rec.Code)
	}
}
