package analytics

import (
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

func TestStoreRecordAndAggregate(t *testing.T) {
	store := NewStore(t.TempDir())
	base := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

	events := []ports.AnalyticsEvent{
		{Type: "upload", Path: "a.txt", User: "alice", Bytes: 100, At: base},
		{Type: "upload", Path: "a.txt", User: "alice", Bytes: 50, At: base.Add(time.Hour)},
		{Type: "download", Path: "a.txt", User: "bob", At: base.Add(2 * time.Hour)},
		{Type: "download", Path: "b.txt", User: "bob", At: base.Add(3 * time.Hour)},
		{Type: "delete", Path: "c.txt", User: "alice", At: base.Add(25 * time.Hour)},
		{Type: "share", Path: "a.txt", User: "alice", At: base.Add(26 * time.Hour)},
		{Type: "login", User: "alice", At: base.Add(27 * time.Hour)},
		{Type: "upload", Path: "old.txt", User: "carol", Bytes: 999, At: base.AddDate(0, 0, -10)},
	}
	for _, e := range events {
		if err := store.Record(e); err != nil {
			t.Fatalf("record: %v", err)
		}
	}

	since := base.Add(-time.Minute)
	until := base.AddDate(0, 0, 2)

	ov, err := store.Overview(since, until)
	if err != nil {
		t.Fatal(err)
	}
	if ov.Uploads != 2 || ov.Downloads != 2 || ov.Deletes != 1 || ov.Shares != 1 || ov.Logins != 1 {
		t.Fatalf("overview counts = %+v", ov)
	}
	if ov.BytesUploaded != 150 {
		t.Fatalf("bytesUploaded = %d", ov.BytesUploaded)
	}
	if ov.ActiveUsers != 2 {
		t.Fatalf("activeUsers = %d, want 2", ov.ActiveUsers)
	}

	timeline, err := store.Timeline(since, until)
	if err != nil {
		t.Fatal(err)
	}
	if len(timeline) != 2 {
		t.Fatalf("timeline buckets = %d, want 2 (%+v)", len(timeline), timeline)
	}
	if timeline[0].Count != 4 || timeline[1].Count != 3 {
		t.Fatalf("bucket counts = %d,%d", timeline[0].Count, timeline[1].Count)
	}

	top, err := store.TopFiles(since, until, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(top) == 0 || top[0].Path != "a.txt" || top[0].Count != 3 {
		t.Fatalf("top files = %+v", top)
	}
	// Old event outside window is excluded from top-files bytes for a.txt only.
	if top[0].Bytes != 150 {
		t.Fatalf("top[0].bytes = %d", top[0].Bytes)
	}
}

func TestStoreEmptyFile(t *testing.T) {
	store := NewStore(t.TempDir())
	ov, err := store.Overview(time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if ov.TotalEvents != 0 || ov.ByType == nil {
		t.Fatalf("empty overview = %+v", ov)
	}
}
