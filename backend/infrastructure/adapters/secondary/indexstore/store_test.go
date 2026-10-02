package indexstore

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

func sampleSnapshot() *ports.IndexSnapshot {
	return &ports.IndexSnapshot{
		TakenAt: time.Unix(1700000000, 0).UTC(),
		Docs: []ports.IndexSnapshotDoc{
			{Path: "a.txt", Size: 10, Modified: time.Unix(100, 0).UTC(), Content: "alpha", Embedded: true},
			{Path: "b/c.md", Size: 20, Modified: time.Unix(200, 0).UTC(), Content: "bravo", Embedded: true},
			{Path: "big.txt", Size: 99, Modified: time.Unix(300, 0).UTC(), Content: "", Embedded: false},
		},
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	store := NewStore(t.TempDir())
	want := sampleSnapshot()
	if err := store.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !got.TakenAt.Equal(want.TakenAt) || len(got.Docs) != len(want.Docs) {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
	for i, d := range got.Docs {
		w := want.Docs[i]
		if d.Path != w.Path || d.Size != w.Size || d.Content != w.Content ||
			d.Embedded != w.Embedded || !d.Modified.Equal(w.Modified) {
			t.Fatalf("doc %d = %+v, want %+v", i, d, w)
		}
	}
}

func TestLoadMissingIsErrNoSnapshot(t *testing.T) {
	store := NewStore(t.TempDir())
	if _, err := store.Load(); err != ports.ErrNoSnapshot {
		t.Fatalf("Load error = %v, want ErrNoSnapshot", err)
	}
}

func TestLoadDetectsCorruption(t *testing.T) {
	cases := map[string]func(t *testing.T, path string){
		"truncated": func(t *testing.T, path string) {
			raw, _ := os.ReadFile(path)
			if err := os.WriteFile(path, raw[:8], 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"bad magic": func(t *testing.T, path string) {
			raw, _ := os.ReadFile(path)
			raw[0] = 'X'
			if err := os.WriteFile(path, raw, 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"bit flip in payload": func(t *testing.T, path string) {
			raw, _ := os.ReadFile(path)
			raw[len(raw)-1] ^= 0x01
			if err := os.WriteFile(path, raw, 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"unsupported version": func(t *testing.T, path string) {
			raw, _ := os.ReadFile(path)
			raw[7] = 99 // version big-endian last byte
			if err := os.WriteFile(path, raw, 0o600); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			store := NewStore(t.TempDir())
			if err := store.Save(sampleSnapshot()); err != nil {
				t.Fatal(err)
			}
			corrupt(t, store.Path())
			if _, err := store.Load(); err == nil {
				t.Fatal("corrupt snapshot loaded without error")
			}
		})
	}
}

func TestSaveOverwritesAtomically(t *testing.T) {
	store := NewStore(t.TempDir())
	if err := store.Save(sampleSnapshot()); err != nil {
		t.Fatal(err)
	}
	replacement := &ports.IndexSnapshot{
		TakenAt: time.Unix(1700001111, 0).UTC(),
		Docs:    []ports.IndexSnapshotDoc{{Path: "only.txt", Size: 1, Modified: time.Unix(1, 0).UTC(), Content: "x", Embedded: true}},
	}
	if err := store.Save(replacement); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Docs) != 1 || got.Docs[0].Path != "only.txt" {
		t.Fatalf("replacement not loaded: %+v", got)
	}
	// No stray temp files left behind.
	entries, err := os.ReadDir(filepath.Dir(store.Path()))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != filepath.Base(store.Path()) && e.Name() != ".file-share" {
			// .file-share itself is the dir; only the snapshot should exist here.
			t.Fatalf("unexpected file left in snapshot dir: %s", e.Name())
		}
	}
}
