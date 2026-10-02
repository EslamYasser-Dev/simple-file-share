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
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

type apiKeysFixture struct {
	handler *APIKeysHandler
	svc     *services.APIKeyService
	alice   *models.User
}

func newAPIKeysFixture(t *testing.T) apiKeysFixture {
	t.Helper()
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	svc := services.NewAPIKeyService(fs.NewAPIKeyFileRepository(dir), userRepo, hasher)
	h := NewAPIKeysHandler(svc)
	h.SetAudit(nil)
	alice := &models.User{Username: "alice", Role: models.RoleMember, Enabled: true, CreatedAt: time.Now().UTC()}
	if err := userRepo.CreateUser(alice); err != nil {
		t.Fatal(err)
	}
	return apiKeysFixture{handler: h, svc: svc, alice: alice}
}

func apiKeyDo(h http.Handler, u *models.User, method, path, body string) *httptest.ResponseRecorder {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
	}
	if u != nil {
		req = req.WithContext(authctx.WithUser(req.Context(), u))
	}
	segs := strings.Split(strings.Trim(path, "/"), "/")
	if len(segs) == 4 && segs[1] == "auth" && segs[2] == "api-keys" {
		req.SetPathValue("id", segs[3])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAPIKeysCreateListRevokeFlow(t *testing.T) {
	f := newAPIKeysFixture(t)

	// Empty list.
	rec := apiKeyDo(f.handler, f.alice, http.MethodGet, "/api/auth/api-keys", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", rec.Code, rec.Body)
	}
	var list struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list.Items) != 0 {
		t.Fatalf("empty list body = %s err=%v", rec.Body, err)
	}

	// Create.
	rec = apiKeyDo(f.handler, f.alice, http.MethodPost, "/api/auth/api-keys", `{"name":"ci bot","scope":"write"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body)
	}
	var created struct {
		Key   string `json:"key"`
		ID    string `json:"id"`
		Scope string `json:"scope"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("create body = %s err=%v", rec.Body, err)
	}
	if !strings.HasPrefix(created.Key, services.APIKeyPrefix+created.ID+"_") {
		t.Fatalf("key %q does not embed id %q", created.Key, created.ID)
	}
	if created.Scope != models.APIKeyScopeWrite {
		t.Fatalf("scope = %q", created.Scope)
	}

	// List shows metadata but never the secret.
	rec = apiKeyDo(f.handler, f.alice, http.MethodGet, "/api/auth/api-keys", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list.Items) != 1 {
		t.Fatalf("list body = %s err=%v", rec.Body, err)
	}
	if strings.Contains(rec.Body.String(), strings.Split(created.Key, "_")[2]) {
		t.Fatal("list leaked the secret")
	}
	if strings.Contains(rec.Body.String(), "secretHash") {
		t.Fatal("list leaked the hash field")
	}

	// Revoke.
	rec = apiKeyDo(f.handler, f.alice, http.MethodDelete, "/api/auth/api-keys/"+created.ID, "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "revoked") {
		t.Fatalf("revoke = %d body=%s", rec.Code, rec.Body)
	}
	if rec := apiKeyDo(f.handler, f.alice, http.MethodDelete, "/api/auth/api-keys/"+created.ID, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("double revoke = %d, want 404", rec.Code)
	}
	rec = apiKeyDo(f.handler, f.alice, http.MethodGet, "/api/auth/api-keys", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list.Items) != 0 {
		t.Fatalf("after revoke list = %s", rec.Body)
	}
}

func TestAPIKeysCreateValidation(t *testing.T) {
	f := newAPIKeysFixture(t)

	for name, body := range map[string]string{
		"missing name": `{}`,
		"bad scope":    `{"name":"x","scope":"root"}`,
		"broken json":  `{"name":`,
		"empty name":   `{"name":"  "}`,
		"long name":    `{"name":"` + strings.Repeat("n", 81) + `"}`,
		"negative ttl": `{"name":"x","expiresIn":-5}`,
	} {
		rec := apiKeyDo(f.handler, f.alice, http.MethodPost, "/api/auth/api-keys", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d body=%s, want 400", name, rec.Code, rec.Body)
		}
	}

	// No body at all still reaches name validation.
	rec := apiKeyDo(f.handler, f.alice, http.MethodPost, "/api/auth/api-keys", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty body status = %d, want 400", rec.Code)
	}
}

func TestAPIKeysCrossOwnerAndAnonymous(t *testing.T) {
	f := newAPIKeysFixture(t)
	plaintext, key, err := f.svc.Issue(f.alice, "mine", "", 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = plaintext

	// Anonymous callers get 403 on every method.
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/auth/api-keys"},
		{http.MethodPost, "/api/auth/api-keys"},
		{http.MethodDelete, "/api/auth/api-keys/" + key.ID},
	} {
		if rec := apiKeyDo(f.handler, nil, tc.method, tc.path, `{"name":"x"}`); rec.Code != http.StatusForbidden {
			t.Errorf("%s %s = %d, want 403", tc.method, tc.path, rec.Code)
		}
	}

	// Wrong method → 405.
	if rec := apiKeyDo(f.handler, f.alice, http.MethodPut, "/api/auth/api-keys", ""); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("PUT = %d, want 405", rec.Code)
	}
}

func TestAPIKeysCreateExpiryTimestamp(t *testing.T) {
	f := newAPIKeysFixture(t)
	rec := apiKeyDo(f.handler, f.alice, http.MethodPost, "/api/auth/api-keys",
		`{"name":"temp","expiresIn":3600}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create = %d body=%s", rec.Code, rec.Body)
	}
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	keys, err := f.svc.List(f.alice)
	if err != nil || len(keys) != 1 {
		t.Fatalf("list: %v %d", err, len(keys))
	}
	if keys[0].ExpiresAt.Before(time.Now().Add(30 * time.Minute)) {
		t.Fatalf("expiresAt = %v, want ~1h out", keys[0].ExpiresAt)
	}
}
