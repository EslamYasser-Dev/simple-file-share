package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

type stubAuditStore struct {
	entries []ports.AuditEntry
	lastQ   ports.AuditQuery
}

func (s *stubAuditStore) Record(e ports.AuditEntry) error {
	s.entries = append(s.entries, e)
	return nil
}

func (s *stubAuditStore) Query(q ports.AuditQuery) ([]ports.AuditEntry, string, error) {
	s.lastQ = q
	next := ""
	if len(q.Cursor) == 0 && len(s.entries) > q.Limit && q.Limit > 0 {
		next = "next-page"
	}
	return s.entries, next, nil
}

func auditFixture(t *testing.T) (*AdminAuditHandler, *stubAuditStore) {
	t.Helper()
	store := &stubAuditStore{}
	store.entries = []ports.AuditEntry{
		{Action: "login.fail", Actor: "mallory"},
		{Action: "login.ok", Actor: "alice"},
	}
	return NewAdminAuditHandler(services.NewAuditService(store, services.NewRoleCatalog(nil))), store
}

func TestAdminAuditHandlerReturnsObjectPage(t *testing.T) {
	h, _ := auditFixture(t)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit?limit=1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body)
	}

	// The pagination contract is an OBJECT: {"items":[...],"nextCursor":"..."}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("response must be a JSON object: %v", err)
	}
	if _, ok := raw["items"]; !ok {
		t.Fatalf("missing items key: %s", rec.Body)
	}
	if _, ok := raw["nextCursor"]; !ok {
		t.Fatalf("missing nextCursor key: %s", rec.Body)
	}

	var page struct {
		Items      []ports.AuditEntry `json:"items"`
		NextCursor string             `json:"nextCursor"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(page.Items) != 2 || page.NextCursor == "" {
		t.Fatalf("page = %+v, want 2 items and a cursor", page)
	}
}

func TestAdminAuditHandlerPassesFilters(t *testing.T) {
	h, store := auditFixture(t)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit?action=login.fail&actor=mallory&limit=10&cursor=abc", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if store.lastQ.Action != "login.fail" || store.lastQ.Actor != "mallory" || store.lastQ.Limit != 10 || store.lastQ.Cursor != "abc" {
		t.Fatalf("query not passed through: %+v", store.lastQ)
	}
}

func TestAdminAuditHandlerMethodNotAllowed(t *testing.T) {
	h, _ := auditFixture(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/admin/audit", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestAdminAuditHandlerForbiddenForMember(t *testing.T) {
	h, _ := auditFixture(t)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit", nil)
	req = req.WithContext(authctx.WithUser(req.Context(), &models.User{
		Username: "bob", Role: models.RoleMember, Enabled: true,
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", rec.Code, rec.Body)
	}
}

func TestAdminAuditHandlerSystemViewAllowed(t *testing.T) {
	h, _ := auditFixture(t)
	// nil user in context == auth disabled system view.
	req := httptest.NewRequest(http.MethodGet, "/api/admin/audit", nil)
	req = req.WithContext(authctx.WithUser(req.Context(), nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body)
	}
}
