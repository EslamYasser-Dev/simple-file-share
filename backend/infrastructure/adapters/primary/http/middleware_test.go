package xhttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
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

func TestAuthMiddlewareAttachesUser(t *testing.T) {
	alice := &models.User{Username: "alice", IsAdmin: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := UserFromContext(r.Context())
		if got == nil || got.Username != "alice" {
			t.Errorf("context user = %+v, want alice", got)
		}
		w.WriteHeader(http.StatusOK)
	})
	handler := AuthMiddleware(fakeAuthProvider{user: alice})(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("alice", "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAuthMiddlewareRejectsMissingCredentials(t *testing.T) {
	handler := AuthMiddleware(fakeAuthProvider{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("next handler should not be called")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthMiddlewareRejectsBadCredentials(t *testing.T) {
	handler := AuthMiddleware(fakeAuthProvider{err: errors.New("nope")})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
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

func TestUserFromContextDefaultsNil(t *testing.T) {
	if u := UserFromContext(httptest.NewRequest(http.MethodGet, "/", nil).Context()); u != nil {
		t.Errorf("user = %+v, want nil", u)
	}
}
