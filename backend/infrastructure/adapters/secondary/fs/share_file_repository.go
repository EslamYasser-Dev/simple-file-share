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
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// ShareFileRepository persists public share links as a JSON document under the
// storage root's metadata directory (`.file-share/shares.json`), which the
// search index already skips. Writes are atomic (temp file + rename) and
// guarded by an in-process mutex, mirroring UserFileRepository.
type ShareFileRepository struct {
	path string
	mu   sync.RWMutex
}

var _ ports.ShareRepository = (*ShareFileRepository)(nil)

func NewShareFileRepository(rootDir string) *ShareFileRepository {
	return &ShareFileRepository{
		path: filepath.Join(rootDir, ".file-share", "shares.json"),
	}
}

type shareDocument struct {
	Token     string    `json:"token"`
	Path      string    `json:"path"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"createdAt"`
	IsAdmin   bool      `json:"isAdmin,omitempty"`
	ExpiresAt time.Time `json:"expiresAt,omitempty"`
}

func fromShare(s *models.Share) shareDocument {
	return shareDocument{
		Token:     s.Token,
		Path:      s.Path,
		Owner:     s.Owner,
		CreatedAt: s.CreatedAt,
		IsAdmin:   s.IsAdmin,
		ExpiresAt: s.ExpiresAt,
	}
}

func (d shareDocument) toShare() *models.Share {
	return &models.Share{
		Token:     d.Token,
		Path:      d.Path,
		Owner:     d.Owner,
		CreatedAt: d.CreatedAt,
		IsAdmin:   d.IsAdmin,
		ExpiresAt: d.ExpiresAt,
	}
}

func (r *ShareFileRepository) Create(share *models.Share) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	shares, err := r.loadLocked()
	if err != nil {
		return err
	}

	// Replace any existing share with the same token (keeps the file tidy).
	out := make([]shareDocument, 0, len(shares)+1)
	added := false
	for _, existing := range shares {
		if existing.Token == share.Token {
			out = append(out, fromShare(share))
			added = true
			continue
		}
		out = append(out, existing)
	}
	if !added {
		out = append(out, fromShare(share))
	}
	return r.saveLocked(out)
}

func (r *ShareFileRepository) FindByToken(token string) (*models.Share, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shares, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	for _, doc := range shares {
		if doc.Token == token {
			return doc.toShare(), nil
		}
	}
	return nil, domainerrors.ErrShareNotFound
}

func (r *ShareFileRepository) ListByOwner(owner string) ([]*models.Share, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shares, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	out := make([]*models.Share, 0, len(shares))
	for _, doc := range shares {
		if doc.Owner == owner {
			out = append(out, doc.toShare())
		}
	}
	sortShares(out)
	return out, nil
}

func (r *ShareFileRepository) ListAll() ([]*models.Share, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shares, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	out := make([]*models.Share, 0, len(shares))
	for _, doc := range shares {
		out = append(out, doc.toShare())
	}
	sortShares(out)
	return out, nil
}

func (r *ShareFileRepository) Delete(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	shares, err := r.loadLocked()
	if err != nil {
		return err
	}
	out := make([]shareDocument, 0, len(shares))
	removed := false
	for _, doc := range shares {
		if doc.Token == token {
			removed = true
			continue
		}
		out = append(out, doc)
	}
	if !removed {
		return domainerrors.ErrShareNotFound
	}
	return r.saveLocked(out)
}

// PurgeExpired atomically removes shares whose validity window has passed and
// returns the number removed. Callers may invoke it opportunistically (e.g. on
// startup and each public fetch) to keep the store from accumulating dead links.
func (r *ShareFileRepository) PurgeExpired(now time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	shares, err := r.loadLocked()
	if err != nil {
		return 0, err
	}
	out := make([]shareDocument, 0, len(shares))
	removed := 0
	for _, doc := range shares {
		if expiredAt(now, doc.ExpiresAt) {
			removed++
			continue
		}
		out = append(out, doc)
	}
	if removed == 0 {
		return 0, nil
	}
	if err := r.saveLocked(out); err != nil {
		return 0, err
	}
	return removed, nil
}

func expiredAt(now time.Time, expiresAt time.Time) bool {
	return !expiresAt.IsZero() && now.After(expiresAt)
}

// sortShares orders shares newest first.
func sortShares(shares []*models.Share) {
	sort.Slice(shares, func(i, j int) bool {
		return shares[i].CreatedAt.After(shares[j].CreatedAt)
	})
}

func (r *ShareFileRepository) loadLocked() ([]shareDocument, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return []shareDocument{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read shares file: %w", err)
	}

	var docs []shareDocument
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("parse shares file: %w", err)
	}
	if docs == nil {
		docs = []shareDocument{}
	}
	return docs, nil
}

func (r *ShareFileRepository) saveLocked(docs []shareDocument) error {
	if err := os.MkdirAll(filepath.Dir(r.path), storageDirPerm); err != nil {
		return fmt.Errorf("create shares dir: %w", err)
	}

	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode shares: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(r.path), ".shares-*.tmp")
	if err != nil {
		return fmt.Errorf("create shares temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write shares temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync shares temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close shares temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), r.path); err != nil {
		return fmt.Errorf("replace shares file: %w", err)
	}
	return nil
}
