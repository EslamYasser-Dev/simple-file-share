package services

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/indexstore"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

// bootEnv is an in-memory filesystem stand-in: entries describe the stat
// walk, contents feed per-file extraction, and counters prove which paths
// were actually read.
type bootEnv struct {
	entries   []*models.FileInfo
	contents  map[string]string
	opens     []string
	textWalks int
}

func (e *bootEnv) walk(string) ([]*models.FileInfo, error) {
	return append([]*models.FileInfo(nil), e.entries...), nil
}

func (e *bootEnv) textWalk(string) ([]ports.TextDocument, error) {
	e.textWalks++
	var docs []ports.TextDocument
	for _, entry := range e.entries {
		if entry.IsDir {
			continue
		}
		if content, ok := e.contents[entry.Path]; ok {
			docs = append(docs, ports.TextDocument{Path: entry.Path, Content: content})
		}
	}
	return docs, nil
}

func (e *bootEnv) open(rel string) (io.ReadCloser, string, error) {
	e.opens = append(e.opens, rel)
	content, ok := e.contents[rel]
	if !ok {
		return nil, "", os.ErrNotExist
	}
	return io.NopCloser(strings.NewReader(content)), "text/plain", nil
}

func file(path string, size int64, mod time.Time) *models.FileInfo {
	return &models.FileInfo{
		Name:     filepath.Base(path),
		Path:     path,
		Size:     size,
		Modified: mod,
	}
}

func runBuild(t *testing.T, svc *RebuildIndexService, env *bootEnv) IndexBuild {
	t.Helper()
	build, err := svc.Execute("root", env.walk, env.textWalk)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if build.SnapshotErr != nil {
		t.Fatalf("SnapshotErr: %v", build.SnapshotErr)
	}
	return build
}

func TestRebuildWithoutSnapshotsIsFullEveryBoot(t *testing.T) {
	env := &bootEnv{
		entries:  []*models.FileInfo{file("a.txt", 10, time.Unix(100, 0))},
		contents: map[string]string{"a.txt": "alpha content"},
	}
	svc := NewRebuildIndexService(memory.NewFileIndexRepository(), memory.NewTextIndex())

	first := runBuild(t, svc, env)
	if first.Source != "full" || first.TextDocs != 1 {
		t.Fatalf("first build = %+v", first)
	}
	second := runBuild(t, svc, env)
	if second.Source != "full" || env.textWalks != 2 {
		t.Fatalf("snapshots off must always full-walk: %+v textWalk=%d", second, env.textWalks)
	}
}

func TestRebuildSnapshotSecondBootSkipsExtraction(t *testing.T) {
	env := &bootEnv{
		entries: []*models.FileInfo{
			file("a.txt", 10, time.Unix(100, 0)),
			file("dir/b.md", 20, time.Unix(200, 0)),
		},
		contents: map[string]string{"a.txt": "alpha unique", "dir/b.md": "bravo unique"},
	}
	index := memory.NewFileIndexRepository()
	text := memory.NewTextIndex()
	svc := NewRebuildIndexService(index, text)
	svc.EnableSnapshotPersist(indexstore.NewStore(t.TempDir()), env.open)

	first := runBuild(t, svc, env)
	if first.Source != "full" {
		t.Fatalf("first boot source = %q, want full (no snapshot yet)", first.Source)
	}
	if len(env.opens) != 0 {
		t.Fatalf("full path must use textWalk, not open; opens=%v", env.opens)
	}

	// Second boot: identical fingerprints → everything from the snapshot.
	env.opens = nil
	env.textWalks = 0
	second := runBuild(t, svc, env)
	if second.Source != "snapshot" {
		t.Fatalf("second boot source = %q, want snapshot", second.Source)
	}
	if env.textWalks != 0 {
		t.Fatalf("textWalk called %d times on incremental boot", env.textWalks)
	}
	if len(env.opens) != 0 {
		t.Fatalf("unchanged files must not be re-read, opens=%v", env.opens)
	}
	if second.TextDocs != 2 {
		t.Fatalf("TextDocs = %d, want 2", second.TextDocs)
	}

	// The restored content must actually be searchable.
	hits, err := text.Search("bravo", 10)
	if err != nil || len(hits) == 0 {
		t.Fatalf("snapshot-restored content not searchable: hits=%v err=%v", hits, err)
	}
}

func TestRebuildExtractsOnlyChangedAndNewFiles(t *testing.T) {
	older := time.Unix(100, 0)
	env := &bootEnv{
		entries: []*models.FileInfo{
			file("keep.txt", 10, older),
			file("changed.txt", 11, older),
		},
		contents: map[string]string{"keep.txt": "keep original", "changed.txt": "changed old"},
	}
	svc := NewRebuildIndexService(memory.NewFileIndexRepository(), memory.NewTextIndex())
	svc.EnableSnapshotPersist(indexstore.NewStore(t.TempDir()), env.open)
	runBuild(t, svc, env)

	// Mutate one file (new size+mtime) and add one; leave the other alone.
	newer := time.Unix(500, 0)
	env.entries = []*models.FileInfo{
		file("keep.txt", 10, older),
		file("changed.txt", 99, newer),
		file("added.txt", 7, newer),
	}
	env.contents["changed.txt"] = "changed new"
	env.contents["added.txt"] = "brand new"
	env.opens = nil

	build := runBuild(t, svc, env)
	if build.Source != "snapshot" {
		t.Fatalf("source = %q", build.Source)
	}
	wantOpened := map[string]bool{"changed.txt": true, "added.txt": true}
	if len(env.opens) != len(wantOpened) {
		t.Fatalf("opens = %v, want exactly changed+added", env.opens)
	}
	for _, p := range env.opens {
		if !wantOpened[p] {
			t.Fatalf("unexpected open of %q", p)
		}
	}
	if build.TextDocs != 3 {
		t.Fatalf("TextDocs = %d, want 3", build.TextDocs)
	}
}

func TestRebuildDropsRemovedFiles(t *testing.T) {
	older := time.Unix(100, 0)
	env := &bootEnv{
		entries:  []*models.FileInfo{file("gone.txt", 10, older), file("stay.txt", 10, older)},
		contents: map[string]string{"gone.txt": "vanish", "stay.txt": "remain"},
	}
	text := memory.NewTextIndex()
	svc := NewRebuildIndexService(memory.NewFileIndexRepository(), text)
	svc.EnableSnapshotPersist(indexstore.NewStore(t.TempDir()), env.open)
	runBuild(t, svc, env)

	env.entries = []*models.FileInfo{file("stay.txt", 10, older)}
	env.opens = nil
	build := runBuild(t, svc, env)
	if build.TextDocs != 1 {
		t.Fatalf("TextDocs = %d, want 1 (removed file dropped)", build.TextDocs)
	}
	hits, err := text.Search("vanish", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("removed file still searchable: %v", hits)
	}
}

func TestRebuildSkipsNonIndexableNewFiles(t *testing.T) {
	older := time.Unix(100, 0)
	env := &bootEnv{
		entries:  []*models.FileInfo{file("photo.png", 10, older)},
		contents: map[string]string{"photo.png": "not really a png"},
	}
	svc := NewRebuildIndexService(memory.NewFileIndexRepository(), memory.NewTextIndex())
	svc.EnableSnapshotPersist(indexstore.NewStore(t.TempDir()), env.open)

	build := runBuild(t, svc, env)
	if build.Source != "full" { // no snapshot yet → textWalk (fake, includes png)
		t.Fatalf("source = %q", build.Source)
	}

	// Next boot: new binary appears; must not be opened via extractOne.
	env.entries = append(env.entries, file("photo2.bin", 5, older))
	env.contents["photo2.bin"] = "binary"
	env.opens = nil
	runBuild(t, svc, env)
	for _, p := range env.opens {
		if p == "photo2.bin" || p == "photo.png" {
			t.Fatalf("non-indexable path extracted: %v", env.opens)
		}
	}
}

func TestRebuildCorruptSnapshotFallsBackToFull(t *testing.T) {
	dir := t.TempDir()
	env := &bootEnv{
		entries:  []*models.FileInfo{file("a.txt", 10, time.Unix(100, 0))},
		contents: map[string]string{"a.txt": "alpha"},
	}
	svc := NewRebuildIndexService(memory.NewFileIndexRepository(), memory.NewTextIndex())
	store := indexstore.NewStore(dir)
	svc.EnableSnapshotPersist(store, env.open)
	runBuild(t, svc, env)

	// Corrupt the snapshot payload (keep header so Load must catch CRC).
	raw, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 0xFF
	if err := os.WriteFile(store.Path(), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	env.opens = nil
	env.textWalks = 0
	build := runBuild(t, svc, env)
	if build.Source != "full" || env.textWalks != 1 {
		t.Fatalf("corrupt snapshot must force full rebuild: %+v textWalk=%d", build, env.textWalks)
	}
	// A fresh valid snapshot must have been written for the next boot.
	if _, err := store.Load(); err != nil {
		t.Fatalf("fresh snapshot unreadable: %v", err)
	}
}

func TestRebuildSnapshotSaveFailureIsNonFatal(t *testing.T) {
	env := &bootEnv{
		entries:  []*models.FileInfo{file("a.txt", 10, time.Unix(100, 0))},
		contents: map[string]string{"a.txt": "alpha"},
	}
	svc := NewRebuildIndexService(memory.NewFileIndexRepository(), memory.NewTextIndex())
	svc.EnableSnapshotPersist(failingStore{}, env.open)

	build, err := svc.Execute("root", env.walk, env.textWalk)
	if err != nil {
		t.Fatalf("Execute must survive save failure: %v", err)
	}
	if build.SnapshotErr == nil {
		t.Fatal("SnapshotErr must surface the save failure")
	}
}

type failingStore struct{}

func (failingStore) Load() (*ports.IndexSnapshot, error) { return nil, ports.ErrNoSnapshot }
func (failingStore) Save(*ports.IndexSnapshot) error {
	return errors.New("disk full")
}

var _ ports.IndexSnapshotStore = failingStore{}
