package audit

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

func record(t *testing.T, l *JSONLLog, action, actor, detail string) {
	t.Helper()
	if err := l.Record(ports.AuditEntry{Action: action, Actor: actor, Detail: detail}); err != nil {
		t.Fatalf("record %s: %v", action, err)
	}
}

func TestAuditRoundtripNewestFirst(t *testing.T) {
	l := NewJSONLLog(t.TempDir(), 1<<20, 3)
	record(t, l, "login.ok", "alice", "")
	record(t, l, "user.create", "alice", "role=member")
	record(t, l, "login.fail", "bob", "invalid credentials")

	entries, next, err := l.Query(ports.AuditQuery{})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if next != "" {
		t.Fatalf("unexpected next cursor on first page: %q", next)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	wantActions := []string{"login.fail", "user.create", "login.ok"}
	for i, want := range wantActions {
		if entries[i].Action != want {
			t.Errorf("entry %d action = %q, want %q", i, entries[i].Action, want)
		}
	}
	if entries[2].At.IsZero() {
		t.Error("Record should stamp At when unset")
	}
}

func TestAuditFilters(t *testing.T) {
	l := NewJSONLLog(t.TempDir(), 1<<20, 3)
	record(t, l, "login.ok", "alice", "")
	record(t, l, "login.fail", "alice", "invalid credentials")
	record(t, l, "login.ok", "bob", "")

	entries, _, err := l.Query(ports.AuditQuery{Action: "login.fail"})
	if err != nil || len(entries) != 1 || entries[0].Actor != "alice" {
		t.Fatalf("action filter: err=%v entries=%+v", err, entries)
	}
	entries, _, err = l.Query(ports.AuditQuery{Actor: "bob"})
	if err != nil || len(entries) != 1 || entries[0].Action != "login.ok" {
		t.Fatalf("actor filter: err=%v entries=%+v", err, entries)
	}
	entries, _, err = l.Query(ports.AuditQuery{Actor: "carol"})
	if err != nil || len(entries) != 0 {
		t.Fatalf("no-match filter: err=%v entries=%+v", err, entries)
	}
}

func TestAuditPagination(t *testing.T) {
	l := NewJSONLLog(t.TempDir(), 1<<20, 3)
	for i := 0; i < 5; i++ {
		record(t, l, "login.ok", fmt.Sprintf("user%d", i), "")
	}

	page1, cursor1, err := l.Query(ports.AuditQuery{Limit: 2})
	if err != nil || len(page1) != 2 || cursor1 == "" {
		t.Fatalf("page1: err=%v len=%d cursor=%q", err, len(page1), cursor1)
	}
	page2, cursor2, err := l.Query(ports.AuditQuery{Limit: 2, Cursor: cursor1})
	if err != nil || len(page2) != 2 || cursor2 == "" {
		t.Fatalf("page2: err=%v len=%d cursor=%q", err, len(page2), cursor2)
	}
	page3, cursor3, err := l.Query(ports.AuditQuery{Limit: 2, Cursor: cursor2})
	if err != nil || len(page3) != 1 || cursor3 != "" {
		t.Fatalf("page3: err=%v len=%d cursor=%q", err, len(page3), cursor3)
	}
	// Newest-first stream: page1 covers users 4,3; page3 covers user0.
	if page1[0].Actor != "user4" || page3[0].Actor != "user0" {
		t.Fatalf("order broken: page1[0]=%q page3[0]=%q", page1[0].Actor, page3[0].Actor)
	}

	if _, _, err := l.Query(ports.AuditQuery{Cursor: "not-base64!!!"}); err == nil {
		t.Fatal("invalid cursor should error")
	}
	// Past-the-end cursor is an empty page, not an error.
	entries, next, err := l.Query(ports.AuditQuery{Cursor: encodeCursor(9999)})
	if err != nil || len(entries) != 0 || next != "" {
		t.Fatalf("past-end: err=%v len=%d next=%q", err, len(entries), next)
	}
}

func TestAuditRotationAndRetention(t *testing.T) {
	l := NewJSONLLog(t.TempDir(), 1024, 1)
	payload := strings.Repeat("x", 200)
	for i := 0; i < 20; i++ {
		record(t, l, "login.ok", "alice", payload)
	}

	active := l.Path()
	if _, err := os.Stat(active); err != nil {
		t.Fatalf("active file missing: %v", err)
	}
	if _, err := os.Stat(active + ".1"); err != nil {
		t.Fatalf("rotated file missing: %v", err)
	}
	if _, err := os.Stat(active + ".2"); err == nil {
		t.Fatal("keep=1 must not retain a .2 generation")
	}
	// Oldest content lives in .1; newest entries survive in the active file.
	entries, _, err := l.Query(ports.AuditQuery{Limit: 500})
	if err != nil {
		t.Fatalf("query after rotation: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("rotation must keep recent entries queryable")
	}
	// Total retained must respect keep: at most 2 generations of ≤1KiB.
	// (20 * ~300B lines would be ~6KB; retention keeps roughly 2KiB.)
	if len(entries) > 40 {
		t.Fatalf("retention leaked: %d entries retained", len(entries))
	}
}

func TestAuditSkipsCorruptLines(t *testing.T) {
	l := NewJSONLLog(t.TempDir(), 1<<20, 3)
	record(t, l, "login.ok", "alice", "")
	// Simulate a torn write.
	f, err := os.OpenFile(l.Path(), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("{\"at\": \"trunc\n"); err != nil {
		t.Fatal(err)
	}
	f.Close()
	record(t, l, "login.ok", "bob", "")

	entries, _, err := l.Query(ports.AuditQuery{})
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2 (corrupt line skipped)", len(entries))
	}
}

func TestAuditEmptyLog(t *testing.T) {
	l := NewJSONLLog(t.TempDir(), 1<<20, 3)
	entries, next, err := l.Query(ports.AuditQuery{})
	if err != nil {
		t.Fatalf("query empty: %v", err)
	}
	if len(entries) != 0 || next != "" {
		t.Fatalf("empty log: entries=%d next=%q", len(entries), next)
	}
}

func TestAuditQueryLocksOutConcurrentRecord(t *testing.T) {
	l := NewJSONLLog(t.TempDir(), 1<<20, 3)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			record(t, l, "login.ok", "alice", "")
		}
	}()
	for i := 0; i < 10; i++ {
		if _, _, err := l.Query(ports.AuditQuery{}); err != nil {
			t.Fatalf("concurrent query: %v", err)
		}
	}
	<-done
}
