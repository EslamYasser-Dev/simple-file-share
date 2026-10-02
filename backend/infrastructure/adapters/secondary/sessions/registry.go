// Package sessions persists the active-session list as JSON under the
// storage root's metadata directory (`.file-share/sessions.json`), which the
// search index already skips. Writes are atomic and guarded by a mutex; the
// list is pruned on load and bounded per subject so it cannot grow forever.
package sessions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// maxPerSubject caps retained sessions for one account (newest kept).
const maxPerSubject = 50

// Registry is a file-backed ports.SessionRegistry.
type Registry struct {
	path string
	mu   sync.Mutex
}

var _ ports.SessionRegistry = (*Registry)(nil)

// NewRegistry stores the list at ROOT_DIR/.file-share/sessions.json.
func NewRegistry(rootDir string) *Registry {
	return &Registry{path: filepath.Join(rootDir, ".file-share", "sessions.json")}
}

type document struct {
	Sessions []ports.Session `json:"sessions"`
}

// Record upserts by JTI, prunes expired sessions, enforces the per-subject
// cap, and saves atomically.
func (r *Registry) Record(sess ports.Session) error {
	if sess.JTI == "" {
		return fmt.Errorf("session jti required")
	}
	if sess.IssuedAt.IsZero() {
		sess.IssuedAt = time.Now().UTC()
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, err := r.loadLocked()
	if err != nil {
		return err
	}
	doc.Sessions = upsert(doc.Sessions, sess)
	doc.Sessions = pruneExpired(doc.Sessions)
	doc.Sessions = capPerSubject(doc.Sessions)
	return r.saveLocked(doc)
}

// List returns a subject's sessions newest first.
func (r *Registry) List(subject string) ([]ports.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	out := make([]ports.Session, 0, len(doc.Sessions))
	for _, s := range doc.Sessions {
		if subject == "" || s.Subject == subject {
			out = append(out, s)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].IssuedAt.After(out[j].IssuedAt)
	})
	return out, nil
}

// Find returns the session with the given JTI.
func (r *Registry) Find(jti string) (*ports.Session, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc, err := r.loadLocked()
	if err != nil {
		return nil, false, err
	}
	for i := range doc.Sessions {
		if doc.Sessions[i].JTI == jti {
			s := doc.Sessions[i]
			return &s, true, nil
		}
	}
	return nil, false, nil
}

// Remove drops a session (idempotent).
func (r *Registry) Remove(jti string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	doc, err := r.loadLocked()
	if err != nil {
		return err
	}
	kept := doc.Sessions[:0]
	found := false
	for _, s := range doc.Sessions {
		if s.JTI == jti {
			found = true
			continue
		}
		kept = append(kept, s)
	}
	if !found {
		return nil
	}
	return r.saveLocked(document{Sessions: kept})
}

func (r *Registry) loadLocked() (document, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return document{Sessions: []ports.Session{}}, nil
		}
		return document{}, fmt.Errorf("read sessions: %w", err)
	}
	var doc document
	if err := json.Unmarshal(data, &doc); err != nil {
		// A corrupt list must not brick login; start over (revocation still
		// works through the JWT revocation store).
		return document{Sessions: []ports.Session{}}, nil
	}
	if doc.Sessions == nil {
		doc.Sessions = []ports.Session{}
	}
	return doc, nil
}

func (r *Registry) saveLocked(doc document) error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create sessions dir: %w", err)
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sessions: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".sessions-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp sessions: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write sessions: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("sync sessions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close sessions: %w", err)
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		os.Remove(tmpName)
		return nil // best effort — the file is already private to the dir
	}
	if err := os.Rename(tmpName, r.path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("replace sessions: %w", err)
	}
	return nil
}

func upsert(sessions []ports.Session, sess ports.Session) []ports.Session {
	for i := range sessions {
		if sessions[i].JTI == sess.JTI {
			sessions[i] = sess
			return sessions
		}
	}
	return append(sessions, sess)
}

func pruneExpired(sessions []ports.Session) []ports.Session {
	now := time.Now()
	kept := sessions[:0]
	for _, s := range sessions {
		if s.ExpiresAt.After(now) {
			kept = append(kept, s)
		}
	}
	return kept
}

// capPerSubject keeps the newest maxPerSubject sessions for each subject.
func capPerSubject(sessions []ports.Session) []ports.Session {
	bySubject := make(map[string][]ports.Session)
	order := make([]string, 0)
	for _, s := range sessions {
		if _, ok := bySubject[s.Subject]; !ok {
			order = append(order, s.Subject)
		}
		bySubject[s.Subject] = append(bySubject[s.Subject], s)
	}
	var out []ports.Session
	for _, subject := range order {
		list := bySubject[subject]
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].IssuedAt.After(list[j].IssuedAt)
		})
		if len(list) > maxPerSubject {
			list = list[:maxPerSubject]
		}
		out = append(out, list...)
	}
	return out
}
