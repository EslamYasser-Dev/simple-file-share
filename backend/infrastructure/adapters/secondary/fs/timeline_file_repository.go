package fs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// maxTimelineEvents bounds the feed store: timeline is a live feed, not an
// archive, so the oldest entries are pruned on append past this cap.
const maxTimelineEvents = 5000

// TimelineFileRepository persists feed events as a JSON document under the
// storage root's metadata directory (`.file-share/timeline.json`). Writes are
// atomic (temp file + rename) and guarded by an in-process mutex, mirroring
// ShareFileRepository.
type TimelineFileRepository struct {
	path string
	mu   sync.RWMutex
	// loaded guards the in-memory event log, kept oldest-first like the file
	// so appends stay O(1) amortized. Feed pages iterate from the tail
	// instead of re-parsing and re-sorting the JSON file on every request;
	// the file stays the source of truth across restarts.
	loaded bool
	events []*models.TimelineEvent
}

var _ ports.TimelineRepository = (*TimelineFileRepository)(nil)

func NewTimelineFileRepository(rootDir string) *TimelineFileRepository {
	return &TimelineFileRepository{
		path: filepath.Join(rootDir, ".file-share", "timeline.json"),
	}
}

type timelineDocument struct {
	ID         string    `json:"id"`
	Owner      string    `json:"owner"`
	Kind       string    `json:"kind"`
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Size       int64     `json:"size,omitempty"`
	Visibility string    `json:"visibility"`
	CreatedAt  time.Time `json:"createdAt"`
}

func fromTimelineEvent(e *models.TimelineEvent) timelineDocument {
	return timelineDocument{
		ID:         e.ID,
		Owner:      e.Owner,
		Kind:       e.Kind,
		Path:       e.Path,
		Name:       e.Name,
		Size:       e.Size,
		Visibility: string(e.Visibility),
		CreatedAt:  e.CreatedAt,
	}
}

func (d timelineDocument) toTimelineEvent() *models.TimelineEvent {
	level := models.VisibilityLevel(d.Visibility)
	if !level.Valid() {
		level = models.VisibilityPrivate
	}
	return &models.TimelineEvent{
		ID:         d.ID,
		Owner:      d.Owner,
		Kind:       d.Kind,
		Path:       d.Path,
		Name:       d.Name,
		Size:       d.Size,
		Visibility: level,
		CreatedAt:  d.CreatedAt,
	}
}

func (r *TimelineFileRepository) Append(event *models.TimelineEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureLoadedLocked(); err != nil {
		return err
	}
	cp := *event
	r.events = append(r.events, &cp)
	if overflow := len(r.events) - maxTimelineEvents; overflow > 0 {
		// Oldest entries are appended first, so trim from the front.
		r.events = append([]*models.TimelineEvent{}, r.events[overflow:]...)
	}
	return r.saveLocked(r.documentsLocked())
}

func (r *TimelineFileRepository) List(after string, limit int) ([]*models.TimelineEvent, error) {
	if err := r.ensureLoaded(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*models.TimelineEvent, 0, limit)
	skipping := after != ""
	for i := len(r.events) - 1; i >= 0; i-- {
		e := r.events[i]
		if skipping {
			if e.ID == after {
				skipping = false
			}
			continue
		}
		cp := *e
		out = append(out, &cp)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (r *TimelineFileRepository) UpdateVisibility(owner, path string, level models.VisibilityLevel) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureLoadedLocked(); err != nil {
		return err
	}
	for _, e := range r.events {
		if e.Owner == owner && e.Path == path {
			e.Visibility = level
		}
	}
	return r.saveLocked(r.documentsLocked())
}

// ensureLoaded loads the file into the log on first use, newest-first. No
// lock is held across the upgrade, so concurrent first-use readers cannot
// deadlock.
func (r *TimelineFileRepository) ensureLoaded() error {
	r.mu.RLock()
	if r.loaded {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ensureLoadedLocked()
}

// ensureLoadedLocked loads the file into the log; the caller holds the
// write lock.
func (r *TimelineFileRepository) ensureLoadedLocked() error {
	if r.loaded {
		return nil
	}
	docs, err := r.loadLocked()
	if err != nil {
		return err
	}
	r.events = make([]*models.TimelineEvent, 0, len(docs))
	for _, doc := range docs {
		r.events = append(r.events, doc.toTimelineEvent())
	}
	// Stable ascending sort: equal timestamps keep file (insertion) order,
	// which keeps cursor pagination stable across restarts.
	sort.SliceStable(r.events, func(i, j int) bool {
		return r.events[i].CreatedAt.Before(r.events[j].CreatedAt)
	})
	r.loaded = true
	return nil
}

// documentsLocked flattens the log back to documents in stored order.
func (r *TimelineFileRepository) documentsLocked() []timelineDocument {
	docs := make([]timelineDocument, 0, len(r.events))
	for _, e := range r.events {
		docs = append(docs, fromTimelineEvent(e))
	}
	return docs
}

func (r *TimelineFileRepository) loadLocked() ([]timelineDocument, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return []timelineDocument{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read timeline file: %w", err)
	}

	var docs []timelineDocument
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("parse timeline file: %w", err)
	}
	if docs == nil {
		docs = []timelineDocument{}
	}
	return docs, nil
}

func (r *TimelineFileRepository) saveLocked(docs []timelineDocument) error {
	if err := os.MkdirAll(filepath.Dir(r.path), storageDirPerm); err != nil {
		return fmt.Errorf("create timeline dir: %w", err)
	}

	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode timeline: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(r.path), ".timeline-*.tmp")
	if err != nil {
		return fmt.Errorf("create timeline temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write timeline temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync timeline temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close timeline temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), r.path); err != nil {
		return fmt.Errorf("replace timeline file: %w", err)
	}
	return nil
}
