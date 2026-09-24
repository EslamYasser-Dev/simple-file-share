package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	secondaryauth "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
)

func TestOAuthCallbackMismatchesStateCookie(t *testing.T) {
	svc := newOAuthStartSvc()
	h := NewOAuthCallbackHandler(svc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/github/callback?code=x&state=aaa", nil)
	req.AddCookie(&http.Cookie{Name: oauthStateCookie, Value: "bbb"})
	h.ServeHTTP(rec, req)
	// Complete fails → redirect to SPA error fragment (no token leak).
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc == "" || !contains(loc, "#oauthError=1") {
		t.Fatalf("location = %q, want oauthError fragment", loc)
	}
	if contains(loc, "access_token=") {
		t.Fatal("error redirect must not include a token")
	}
}

func TestOAuthCallbackRequiresStateCookie(t *testing.T) {
	svc := newOAuthStartSvc()
	h := NewOAuthCallbackHandler(svc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/github/callback?code=x&state=aaa", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if contains(rec.Header().Get("Location"), "access_token=") {
		t.Fatal("missing cookie must not issue a token")
	}
}

func TestValidateRedirectBase(t *testing.T) {
	if err := secondaryauth.ValidateRedirectBase("https://files.example"); err != nil {
		t.Errorf("valid base rejected: %v", err)
	}
	if err := secondaryauth.ValidateRedirectBase("https://evil.com/steal"); err == nil {
		t.Error("path on base should be rejected")
	}
	if err := secondaryauth.ValidateRedirectBase("javascript:alert(1)"); err == nil {
		t.Error("non-http scheme should be rejected")
	}
	if err := secondaryauth.ValidateRedirectBase(""); err == nil {
		t.Error("empty base should be rejected")
	}
}

func newOAuthStartSvc() *services.OAuthLoginService {
	return services.NewOAuthLoginServiceForTest()
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
