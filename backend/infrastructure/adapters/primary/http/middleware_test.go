package xhttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
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
	return AuthMiddleware(services.NewAuthenticateService(provider))
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
