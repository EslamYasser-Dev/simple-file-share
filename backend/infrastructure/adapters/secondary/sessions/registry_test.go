package sessions

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

func future(d time.Duration) time.Time { return time.Now().Add(d) }

func TestRecordListNewestFirstAndFilter(t *testing.T) {
	r := NewRegistry(t.TempDir())
	base := time.Now().UTC().Truncate(time.Second)
	mustRecord(t, r, ports.Session{JTI: "j1", Subject: "alice", IssuedAt: base, ExpiresAt: future(time.Hour)})
	mustRecord(t, r, ports.Session{JTI: "j2", Subject: "alice", IssuedAt: base.Add(time.Minute), ExpiresAt: future(time.Hour)})
	mustRecord(t, r, ports.Session{JTI: "j3", Subject: "bob", IssuedAt: base.Add(2 * time.Minute), ExpiresAt: future(time.Hour)})

	alice, err := r.List("alice")
	if err != nil || len(alice) != 2 {
		t.Fatalf("alice: err=%v len=%d", err, len(alice))
	}
	if alice[0].JTI != "j2" {
		t.Fatalf("newest first: got %q, want j2", alice[0].JTI)
	}
	all, err := r.List("")
	if err != nil || len(all) != 3 {
		t.Fatalf("all: err=%v len=%d", err, len(all))
	}
}

func TestRecordUpsertKeepsOneAndEnriches(t *testing.T) {
	r := NewRegistry(t.TempDir())
	issued := time.Now().UTC().Truncate(time.Second)
	mustRecord(t, r, ports.Session{JTI: "j1", Subject: "alice", IssuedAt: issued, ExpiresAt: future(time.Hour)})
	mustRecord(t, r, ports.Session{JTI: "j1", Subject: "alice", IssuedAt: issued, ExpiresAt: future(time.Hour), Remote: "10.0.0.9", UserAgent: "curl"})

	list, err := r.List("alice")
	if err != nil || len(list) != 1 {
		t.Fatalf("upsert must keep one entry: err=%v len=%d", err, len(list))
	}
	if list[0].Remote != "10.0.0.9" || list[0].UserAgent != "curl" {
		t.Fatalf("enrichment lost: %+v", list[0])
	}
}

func TestRecordPrunesExpired(t *testing.T) {
	r := NewRegistry(t.TempDir())
	mustRecord(t, r, ports.Session{JTI: "old", Subject: "alice", IssuedAt: time.Now().Add(-2 * time.Hour), ExpiresAt: time.Now().Add(-time.Hour)})
	mustRecord(t, r, ports.Session{JTI: "live", Subject: "alice", IssuedAt: time.Now(), ExpiresAt: future(time.Hour)})

	list, err := r.List("alice")
	if err != nil || len(list) != 1 || list[0].JTI != "live" {
		t.Fatalf("expired session survived: err=%v list=%+v", err, list)
	}
}

func TestRecordCapsPerSubject(t *testing.T) {
	r := NewRegistry(t.TempDir())
	base := time.Now().UTC()
	for i := 0; i < 60; i++ {
		mustRecord(t, r, ports.Session{
			JTI:       "jti-" + string(rune('a'+i%26)) + time.Duration(i).String(),
			Subject:   "alice",
			IssuedAt:  base.Add(time.Duration(i) * time.Second),
			ExpiresAt: future(24 * time.Hour),
		})
	}
	list, err := r.List("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 50 {
		t.Fatalf("kept %d sessions, want cap of 50", len(list))
	}
	// Newest survive: the last recorded has the latest IssuedAt.
	if !list[0].IssuedAt.After(list[len(list)-1].IssuedAt) {
		t.Fatal("list must be newest first")
	}
}

func TestFindAndRemove(t *testing.T) {
	r := NewRegistry(t.TempDir())
	mustRecord(t, r, ports.Session{JTI: "j1", Subject: "alice", ExpiresAt: future(time.Hour)})

	sess, ok, err := r.Find("j1")
	if err != nil || !ok || sess.Subject != "alice" {
		t.Fatalf("find: err=%v ok=%v sess=%+v", err, ok, sess)
	}
	if _, ok, _ := r.Find("nope"); ok {
		t.Fatal("unknown jti must not be found")
	}
	if err := r.Remove("j1"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := r.Remove("j1"); err != nil {
		t.Fatalf("remove must be idempotent: %v", err)
	}
	if list, _ := r.List("alice"); len(list) != 0 {
		t.Fatalf("session survived remove: %+v", list)
	}
}

func TestMissingAndCorruptFileYieldEmptyList(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry(dir)
	if list, err := r.List(""); err != nil || len(list) != 0 {
		t.Fatalf("missing file: err=%v list=%d", err, len(list))
	}
	path := filepath.Join(dir, ".file-share", "sessions.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if list, err := r.List(""); err != nil || len(list) != 0 {
		t.Fatalf("corrupt file must reset, not fail: err=%v list=%d", err, len(list))
	}
	// The registry is usable again after a corrupt read.
	mustRecord(t, r, ports.Session{JTI: "j1", Subject: "alice", ExpiresAt: future(time.Hour)})
	if list, _ := r.List(""); len(list) != 1 {
		t.Fatalf("recovery record failed: %d", len(list))
	}
}

func TestRecordRejectsMissingJTI(t *testing.T) {
	r := NewRegistry(t.TempDir())
	if err := r.Record(ports.Session{Subject: "alice"}); err == nil {
		t.Fatal("missing jti must error")
	}
}

func mustRecord(t *testing.T, r *Registry, s ports.Session) {
	t.Helper()
	if err := r.Record(s); err != nil {
		t.Fatalf("record %q: %v", s.JTI, err)
	}
}
