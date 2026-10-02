package services

import (
	"errors"
	"strings"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// memAudit is an in-memory ports.AuditLog for service tests.
type memAudit struct {
	entries []ports.AuditEntry
	queries []ports.AuditQuery
}

func (m *memAudit) Record(entry ports.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *memAudit) Query(q ports.AuditQuery) ([]ports.AuditEntry, string, error) {
	m.queries = append(m.queries, q)
	return m.entries, "", nil
}

type failAudit struct{}

func (failAudit) Record(ports.AuditEntry) error { return errors.New("disk full") }
func (failAudit) Query(ports.AuditQuery) ([]ports.AuditEntry, string, error) {
	return nil, "", errors.New("disk full")
}

func TestAuditServiceNilStoreIsSafe(t *testing.T) {
	var nilSvc *AuditService
	nilSvc.Record("login.ok", "alice", "1.2.3.4", "", "") // must not panic
	if nilSvc.Enabled() {
		t.Fatal("nil service must be disabled")
	}

	svc := NewAuditService(nil, NewRoleCatalog(nil))
	if svc.Enabled() {
		t.Fatal("nil store must be disabled")
	}
	svc.Record("login.ok", "alice", "1.2.3.4", "", "") // must not panic
	entries, next, err := svc.Query(nil, ports.AuditQuery{})
	if err != nil || next != "" || len(entries) != 0 {
		t.Fatalf("disabled query: err=%v next=%q entries=%v", err, next, entries)
	}
	if entries == nil {
		t.Fatal("disabled query must return a non-nil empty slice")
	}
}

func TestAuditServiceRecordClipsAndDefaultsActor(t *testing.T) {
	mem := &memAudit{}
	svc := NewAuditService(mem, NewRoleCatalog(nil))
	svc.Record("login.fail", strings.Repeat("a", 5000), strings.Repeat("i", 5000), strings.Repeat("t", 5000), strings.Repeat("d", 5000))
	svc.Record("login.ok", "", "", "", "")

	if got := len(mem.entries[0].Actor); got != 64 {
		t.Errorf("actor len = %d, want 64", got)
	}
	if got := len(mem.entries[0].Target); got != 128 {
		t.Errorf("target len = %d, want 128", got)
	}
	if got := len(mem.entries[0].Remote); got != 64 {
		t.Errorf("remote len = %d, want 64", got)
	}
	if got := len(mem.entries[0].Detail); got != 256 {
		t.Errorf("detail len = %d, want 256", got)
	}
	if mem.entries[0].At.IsZero() {
		t.Error("entry must be stamped")
	}
	if mem.entries[1].Actor != "-" {
		t.Errorf("empty actor = %q, want \"-\"", mem.entries[1].Actor)
	}
}

func TestAuditServiceRecordSwallowsStoreError(t *testing.T) {
	// A failing store must never propagate to the handler.
	svc := NewAuditService(failAudit{}, NewRoleCatalog(nil))
	svc.Record("login.ok", "alice", "", "", "") // must not panic or be observable
}

func TestAuditServiceQueryPermissions(t *testing.T) {
	mem := &memAudit{}
	svc := NewAuditService(mem, NewRoleCatalog(nil))

	if _, _, err := svc.Query(nil, ports.AuditQuery{}); err != nil {
		t.Fatalf("system view (auth disabled) must read audit: %v", err)
	}
	admin := &models.User{Username: "root", Role: models.RoleAdmin, IsAdmin: true, Enabled: true}
	if _, _, err := svc.Query(admin, ports.AuditQuery{}); err != nil {
		t.Fatalf("admin must read audit: %v", err)
	}
	member := &models.User{Username: "bob", Role: models.RoleMember, Enabled: true}
	if _, _, err := svc.Query(member, ports.AuditQuery{}); err == nil {
		t.Fatal("member must not read audit")
	}
}

func TestAuditServiceQueryMapsStoreError(t *testing.T) {
	svc := NewAuditService(failAudit{}, NewRoleCatalog(nil))
	admin := &models.User{Username: "root", Role: models.RoleAdmin, IsAdmin: true, Enabled: true}
	if _, _, err := svc.Query(admin, ports.AuditQuery{}); err == nil {
		t.Fatal("store error must propagate")
	}
}
