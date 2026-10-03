package handlers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

func newGoogleExchangeFixture(t *testing.T, clientID string) (*GoogleExchangeHandler, *rsa.PrivateKey) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	source := func(_ context.Context) (map[string]*rsa.PublicKey, error) {
		return map[string]*rsa.PublicKey{"test-kid": &priv.PublicKey}, nil
	}

	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memory.NewFileIndexRepository(), nil)
	scoper := policy.NewPathScoper()
	hasher := auth.NewPBKDF2Hasher()
	tokens := services.NewTokenService(
		services.NewAuthenticateService(auth.NewUserAuthProvider(userRepo, hasher)),
		auth.NewJWTManager("test-secret", time.Hour),
		userRepo,
	)
	oauthSvc := services.NewOAuthLoginService(
		map[string]*auth.OAuthProvider{},
		userRepo, fileRepo, scoper, tokens, true, 0,
	)
	loginSvc := services.NewGoogleLoginService(
		oauthSvc,
		services.NewGoogleIDTokenVerifier(clientID, source),
	)
	return NewGoogleExchangeHandler(loginSvc), priv
}

func mintExchangeToken(t *testing.T, priv *rsa.PrivateKey, aud string) string {
	t.Helper()
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   "accounts.google.com",
		"aud":   aud,
		"sub":   "google-sub-xyz",
		"email": "carol@gmail.com",
		"name":  "Carol",
		"exp":   now.Add(time.Hour).Unix(),
		"iat":   now.Unix(),
	})
	token.Header["kid"] = "test-kid"
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func postExchange(t *testing.T, h *GoogleExchangeHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/google/exchange", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// TestGoogleExchangeHandler drives the public endpoint: first call creates
// the account, the second returns to it, and bad input is rejected without
// leaking which half failed.
func TestGoogleExchangeHandler(t *testing.T) {
	h, priv := newGoogleExchangeFixture(t, "test-client-id")

	rec := postExchange(t, h, `{"idToken":"`+mintExchangeToken(t, priv, "test-client-id")+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("exchange: got %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"accessToken":"`) {
		t.Fatalf("missing access token: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"isNewAccount":true`) {
		t.Fatalf("first exchange must report new account: %s", rec.Body.String())
	}

	rec = postExchange(t, h, `{"idToken":"`+mintExchangeToken(t, priv, "test-client-id")+`"}`)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"isNewAccount":false`) {
		t.Fatalf("second exchange: got %d %s", rec.Code, rec.Body.String())
	}

	rec = postExchange(t, h, `{"idToken":"garbage"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad token: got %d", rec.Code)
	}
	rec = postExchange(t, h, `{}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing token: got %d", rec.Code)
	}

	unconfigured, _ := newGoogleExchangeFixture(t, "")
	rec = postExchange(t, unconfigured, `{"idToken":"`+mintExchangeToken(t, priv, "test-client-id")+`"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unconfigured: got %d", rec.Code)
	}
}
