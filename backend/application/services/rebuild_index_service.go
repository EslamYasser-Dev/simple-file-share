package services

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// maxSnapshotBytes budgets the total embedded text content per snapshot.
// Beyond it, documents keep their fingerprints but omit content and are
// re-extracted on the next boot — a soft cap so pathological trees cannot
// produce multi-gigabyte snapshots.
const maxSnapshotBytes = 64 << 20 // 64 MiB

// IndexBuild describes one startup index build for logging and metrics.
type IndexBuild struct {
	// Source is "snapshot" when unchanged text documents came from the
	// persisted snapshot, "full" when every document was extracted.
	Source   string
	Files    int
	TextDocs int
	Duration time.Duration
	// SnapshotErr is non-nil when persisting the refreshed snapshot failed.
	// It never fails the boot: it only means the next boot re-extracts more.
	SnapshotErr error
}

// RebuildIndexService rebuilds the metadata search index from the current
// storage state, and optionally the full-text content index.
//
// When snapshot persistence is enabled (EnableSnapshotPersist), the text
// index is rebuilt incrementally: every file from the stat walk is compared
// against the snapshot's size/mtime fingerprint, unchanged documents are
// reused, and only new/changed files are re-extracted. The metadata index is
// always rebuilt from the walk — stat is cheap, extraction is not.
type RebuildIndexService struct {
	index     ports.FileIndexRepository
	text      ports.TextIndex
	snapshots ports.IndexSnapshotStore
	// open reads a storage-relative path for extraction; nil disables
	// snapshots (full extraction every boot).
	open func(rel string) (io.ReadCloser, string, error)
}

func NewRebuildIndexService(index ports.FileIndexRepository, text ports.TextIndex) *RebuildIndexService {
	return &RebuildIndexService{index: index, text: text}
}

// EnableSnapshotPersist turns on incremental text-index rebuilds. open must
// read from the active content backend (local or S3), so one implementation
// serves both storage modes.
func (s *RebuildIndexService) EnableSnapshotPersist(
	store ports.IndexSnapshotStore,
	open func(rel string) (io.ReadCloser, string, error),
) {
	s.snapshots = store
	s.open = open
}

// Execute rebuilds metadata from walk. When textWalk is non-nil it also
// rebuilds the full-text index (startup path), reusing the persisted
// snapshot for unchanged documents when enabled.
func (s *RebuildIndexService) Execute(
	rootDir string,
	walk func(string) ([]*models.FileInfo, error),
	textWalk func(string) ([]ports.TextDocument, error),
) (IndexBuild, error) {
	start := time.Now()
	entries, err := walk(rootDir)
	if err != nil {
		return IndexBuild{}, err
	}
	if err := s.index.Rebuild(entries); err != nil {
		return IndexBuild{}, err
	}
	build := IndexBuild{Files: len(entries)}

	if s.text != nil && textWalk != nil {
		docs, source, persistErr, err := s.buildTextDocs(rootDir, entries, textWalk)
		if err != nil {
			return IndexBuild{}, err
		}
		if err := s.text.Rebuild(docs); err != nil {
			return IndexBuild{}, err
		}
		build.Source = source
		build.TextDocs = len(docs)
		build.SnapshotErr = persistErr
	}
	build.Duration = time.Since(start)
	return build, nil
}

// buildTextDocs resolves the text documents for this boot: either the full
// walk (snapshots off or unusable) or the fingerprint-merged set.
func (s *RebuildIndexService) buildTextDocs(
	rootDir string,
	entries []*models.FileInfo,
	textWalk func(string) ([]ports.TextDocument, error),
) (docs []ports.TextDocument, source string, persistErr error, err error) {
	current := make(map[string]*models.FileInfo, len(entries))
	for _, e := range entries {
		if e != nil && !e.IsDir {
			current[e.Path] = e
		}
	}

	if s.snapshots == nil || s.open == nil {
		full, walkErr := textWalk(rootDir)
		if walkErr != nil {
			return nil, "full", nil, walkErr
		}
		return full, "full", s.persistSnapshot(full, current), nil
	}

	snap, loadErr := s.snapshots.Load()
	if loadErr != nil || snap == nil {
		// Missing or corrupt snapshot: fall back to full extraction, then
		// persist a fresh one so the next boot is incremental.
		full, walkErr := textWalk(rootDir)
		if walkErr != nil {
			return nil, "full", nil, walkErr
		}
		return full, "full", s.persistSnapshot(full, current), nil
	}

	byPath := make(map[string]*ports.IndexSnapshotDoc, len(snap.Docs))
	for i := range snap.Docs {
		byPath[snap.Docs[i].Path] = &snap.Docs[i]
	}

	docs = make([]ports.TextDocument, 0, len(current))
	var reextract []string
	for path, fi := range current {
		old := byPath[path]
		if old != nil && old.Embedded &&
			old.Size == fi.Size && old.Modified.Equal(fi.Modified) {
			docs = append(docs, ports.TextDocument{Path: path, Content: old.Content})
			continue
		}
		// New, changed, or fingerprint-only entry: extract below.
		reextract = append(reextract, path)
	}
	sort.Strings(reextract)
	for _, path := range reextract {
		if doc, ok := s.extractOne(path); ok {
			docs = append(docs, doc)
		}
	}
	// Deterministic order keeps snapshots byte-stable across identical boots.
	sort.Slice(docs, func(i, j int) bool { return docs[i].Path < docs[j].Path })

	return docs, "snapshot", s.persistSnapshot(docs, current), nil
}

// extractOne reads and tokenizes eligibility for a single storage-relative
// path using the same policy as the full text walk. Ineligible, unreadable,
// or vanished files are skipped (they simply drop out of the index).
func (s *RebuildIndexService) extractOne(rel string) (ports.TextDocument, bool) {
	if !policy.IsIndexableTextPath(rel) {
		return ports.TextDocument{}, false
	}
	rc, _, err := s.open(rel)
	if err != nil {
		return ports.TextDocument{}, false
	}
	defer rc.Close()
	content, ok := policy.ReadIndexableText(rel, rc)
	if !ok {
		return ports.TextDocument{}, false
	}
	return ports.TextDocument{Path: rel, Content: content}, true
}

// persistSnapshot stores the merged document set with fresh fingerprints.
// It returns a non-nil error only for logging: a failed save must never
// block boot (the next boot simply re-extracts more).
func (s *RebuildIndexService) persistSnapshot(
	docs []ports.TextDocument,
	current map[string]*models.FileInfo,
) error {
	if s.snapshots == nil {
		return nil
	}
	snap := &ports.IndexSnapshot{TakenAt: time.Now()}
	var embedded int
	for _, d := range docs {
		fi, ok := current[d.Path]
		if !ok {
			continue
		}
		embed := embedded+len(d.Content) <= maxSnapshotBytes
		snap.Docs = append(snap.Docs, ports.IndexSnapshotDoc{
			Path:     d.Path,
			Size:     fi.Size,
			Modified: fi.Modified,
			Content:  d.Content,
			Embedded: embed,
		})
		if embed {
			embedded += len(d.Content)
		}
	}
	if err := s.snapshots.Save(snap); err != nil {
		return fmt.Errorf("persist index snapshot: %w", err)
	}
	return nil
}
