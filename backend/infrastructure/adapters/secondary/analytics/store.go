package analytics

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// Store is a JSONL-backed AnalyticsStore under `.file-share/events.jsonl`.
// Writes are append-only and mutex-guarded; reads load the file and aggregate
// in memory (fine for single-node deployments this app targets).
type Store struct {
	path string
	mu   sync.Mutex
}

var _ ports.AnalyticsStore = (*Store)(nil)

func NewStore(rootDir string) *Store {
	return &Store{path: filepath.Join(rootDir, ".file-share", "events.jsonl")}
}

func (s *Store) Record(e ports.AnalyticsEvent) error {
	if e.Type == "" {
		return nil
	}
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	e.At = e.At.UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return err
}

func (s *Store) load() ([]ports.AnalyticsEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var events []ports.AnalyticsEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e ports.AnalyticsEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		events = append(events, e)
	}
	return events, sc.Err()
}

func inWindow(at, since, until time.Time) bool {
	if at.IsZero() {
		return false
	}
	if !since.IsZero() && at.Before(since) {
		return false
	}
	if !until.IsZero() && !at.Before(until) {
		return false
	}
	return true
}

func (s *Store) Overview(since, until time.Time) (*ports.AnalyticsOverview, error) {
	events, err := s.load()
	if err != nil {
		return nil, err
	}
	ov := &ports.AnalyticsOverview{
		ByType:      map[string]int64{},
		WindowStart: since,
		WindowEnd:   until,
	}
	users := map[string]struct{}{}
	for _, e := range events {
		if !inWindow(e.At, since, until) {
			continue
		}
		ov.TotalEvents++
		ov.ByType[e.Type]++
		if e.User != "" {
			users[e.User] = struct{}{}
		}
		switch e.Type {
		case "upload", "update", "restore":
			ov.Uploads++
			ov.BytesUploaded += e.Bytes
		case "download":
			ov.Downloads++
		case "delete":
			ov.Deletes++
		case "share":
			ov.Shares++
		case "login":
			ov.Logins++
		}
	}
	ov.ActiveUsers = int64(len(users))
	return ov, nil
}

func dayStart(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (s *Store) Timeline(since, until time.Time) ([]ports.AnalyticsBucket, error) {
	events, err := s.load()
	if err != nil {
		return nil, err
	}
	byDay := map[int64]*ports.AnalyticsBucket{}
	for _, e := range events {
		if !inWindow(e.At, since, until) {
			continue
		}
		key := dayStart(e.At).UnixNano()
		b := byDay[key]
		if b == nil {
			b = &ports.AnalyticsBucket{Start: dayStart(e.At)}
			byDay[key] = b
		}
		b.Count++
		b.Bytes += e.Bytes
	}
	out := make([]ports.AnalyticsBucket, 0, len(byDay))
	for _, b := range byDay {
		out = append(out, *b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out, nil
}

func (s *Store) TopFiles(since, until time.Time, limit int) ([]ports.AnalyticsTopFile, error) {
	if limit <= 0 {
		limit = 10
	}
	events, err := s.load()
	if err != nil {
		return nil, err
	}
	type acc struct {
		count int64
		bytes int64
	}
	byPath := map[string]*acc{}
	for _, e := range events {
		if !inWindow(e.At, since, until) || e.Path == "" {
			continue
		}
		switch e.Type {
		case "upload", "download", "update", "delete", "restore":
		default:
			continue
		}
		a := byPath[e.Path]
		if a == nil {
			a = &acc{}
			byPath[e.Path] = a
		}
		a.count++
		a.bytes += e.Bytes
	}
	out := make([]ports.AnalyticsTopFile, 0, len(byPath))
	for p, a := range byPath {
		out = append(out, ports.AnalyticsTopFile{Path: p, Count: a.count, Bytes: a.bytes})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Path < out[j].Path
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
