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

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// FollowFileRepository persists follow edges as a JSON document under the
// storage root's metadata directory (`.file-share/follows.json`), which the
// search index already skips. Writes are atomic (temp file + rename) and
// guarded by an in-process mutex, mirroring ShareFileRepository.
type FollowFileRepository struct {
	path string
	mu   sync.RWMutex
	// loaded guards the in-memory edge index. The JSON file stays the
	// source of truth across restarts; within the process every read and
	// write goes through the maps below, so feed filtering (one
	// IsFollowing per link-scoped entry) is O(1) instead of a file
	// re-parse plus linear scan.
	loaded    bool
	following map[string]map[string]time.Time
}

var _ ports.FollowRepository = (*FollowFileRepository)(nil)

func NewFollowFileRepository(rootDir string) *FollowFileRepository {
	return &FollowFileRepository{
		path: filepath.Join(rootDir, ".file-share", "follows.json"),
	}
}

type followDocument struct {
	Follower  string    `json:"follower"`
	Followee  string    `json:"followee"`
	CreatedAt time.Time `json:"createdAt"`
}

func (r *FollowFileRepository) Follow(follower, followee string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureLoadedLocked(); err != nil {
		return err
	}
	set, ok := r.following[follower]
	if ok && !set[followee].IsZero() {
		return domainerrors.ErrAlreadyFollowing
	}
	if !ok {
		set = make(map[string]time.Time)
		r.following[follower] = set
	}
	set[followee] = time.Now().UTC()
	return r.saveLocked(r.documentsLocked())
}

func (r *FollowFileRepository) Unfollow(follower, followee string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureLoadedLocked(); err != nil {
		return err
	}
	set, ok := r.following[follower]
	if !ok || set[followee].IsZero() {
		return domainerrors.ErrFollowNotFound
	}
	delete(set, followee)
	if len(set) == 0 {
		delete(r.following, follower)
	}
	return r.saveLocked(r.documentsLocked())
}

func (r *FollowFileRepository) IsFollowing(follower, followee string) (bool, error) {
	if err := r.ensureLoaded(); err != nil {
		return false, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	set, ok := r.following[follower]
	return ok && !set[followee].IsZero(), nil
}

func (r *FollowFileRepository) Followers(followee string) ([]string, error) {
	if err := r.ensureLoaded(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []string
	for follower, set := range r.following {
		if !set[followee].IsZero() {
			out = append(out, follower)
		}
	}
	sort.Strings(out)
	return out, nil
}

func (r *FollowFileRepository) Following(follower string) ([]string, error) {
	if err := r.ensureLoaded(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []string
	for followee := range r.following[follower] {
		out = append(out, followee)
	}
	sort.Strings(out)
	return out, nil
}

// ensureLoaded loads the file into the index on first use. No lock is held
// across the upgrade, so concurrent first-use readers cannot deadlock.
func (r *FollowFileRepository) ensureLoaded() error {
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

// ensureLoadedLocked loads the file into the index; the caller holds the
// write lock.
func (r *FollowFileRepository) ensureLoadedLocked() error {
	if r.loaded {
		return nil
	}
	docs, err := r.loadLocked()
	if err != nil {
		return err
	}
	r.following = make(map[string]map[string]time.Time, len(docs))
	for _, doc := range docs {
		set, ok := r.following[doc.Follower]
		if !ok {
			set = make(map[string]time.Time)
			r.following[doc.Follower] = set
		}
		set[doc.Followee] = doc.CreatedAt
	}
	r.loaded = true
	return nil
}

// documentsLocked flattens the index back to documents, sorted for stable
// file output.
func (r *FollowFileRepository) documentsLocked() []followDocument {
	docs := make([]followDocument, 0)
	for follower, set := range r.following {
		for followee, created := range set {
			docs = append(docs, followDocument{Follower: follower, Followee: followee, CreatedAt: created})
		}
	}
	sort.Slice(docs, func(i, j int) bool {
		if docs[i].Follower != docs[j].Follower {
			return docs[i].Follower < docs[j].Follower
		}
		return docs[i].Followee < docs[j].Followee
	})
	return docs
}

func (r *FollowFileRepository) loadLocked() ([]followDocument, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return []followDocument{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read follows file: %w", err)
	}

	var docs []followDocument
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("parse follows file: %w", err)
	}
	if docs == nil {
		docs = []followDocument{}
	}
	return docs, nil
}

func (r *FollowFileRepository) saveLocked(docs []followDocument) error {
	if err := os.MkdirAll(filepath.Dir(r.path), storageDirPerm); err != nil {
		return fmt.Errorf("create follows dir: %w", err)
	}

	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode follows: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(r.path), ".follows-*.tmp")
	if err != nil {
		return fmt.Errorf("create follows temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write follows temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync follows temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close follows temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), r.path); err != nil {
		return fmt.Errorf("replace follows file: %w", err)
	}
	return nil
}
