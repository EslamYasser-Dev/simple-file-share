package fs

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// UploadSessionRepository persists resumable-upload metadata under
// `.file-share/upload_sessions.json` and staged bytes under
// `.file-share/uploads/<id>.part`. Writes are atomic (temp + rename) and
// guarded by an in-process mutex, mirroring ShareFileRepository.
type UploadSessionRepository struct {
	mu       sync.Mutex
	metaPath string
	staging  string
}

var _ ports.UploadSessionRepository = (*UploadSessionRepository)(nil)

func NewUploadSessionRepository(rootDir string) *UploadSessionRepository {
	meta := filepath.Join(rootDir, ".file-share", "upload_sessions.json")
	return &UploadSessionRepository{
		metaPath: meta,
		staging:  filepath.Join(rootDir, ".file-share", "uploads"),
	}
}

type uploadSessionDocument struct {
	ID          string    `json:"id"`
	Owner       string    `json:"owner"`
	Fingerprint string    `json:"fingerprint"`
	Destination string    `json:"destination"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	Offset      int64     `json:"offset"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

func fromSession(s *models.UploadSession) uploadSessionDocument {
	return uploadSessionDocument{
		ID:          s.ID,
		Owner:       s.Owner,
		Fingerprint: s.Fingerprint,
		Destination: s.Destination,
		Filename:    s.Filename,
		Size:        s.Size,
		Offset:      s.Offset,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		ExpiresAt:   s.ExpiresAt,
	}
}

func (d uploadSessionDocument) toSession() *models.UploadSession {
	return &models.UploadSession{
		ID:          d.ID,
		Owner:       d.Owner,
		Fingerprint: d.Fingerprint,
		Destination: d.Destination,
		Filename:    d.Filename,
		Size:        d.Size,
		Offset:      d.Offset,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
		ExpiresAt:   d.ExpiresAt,
	}
}

func (r *UploadSessionRepository) Create(session *models.UploadSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	docs, err := r.loadLocked()
	if err != nil {
		return err
	}

	if session.ID == "" {
		id, genErr := newSessionID()
		if genErr != nil {
			return genErr
		}
		session.ID = id
	}

	if err := r.ensureStagingDir(); err != nil {
		return err
	}
	part := r.partPath(session.ID)
	f, err := os.OpenFile(part, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create staging file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close staging file: %w", err)
	}

	docs[session.ID] = fromSession(session)
	return r.saveLocked(docs)
}

func (r *UploadSessionRepository) Get(id string) (*models.UploadSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	docs, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	doc, ok := docs[id]
	if !ok {
		return nil, &domainerrors.NotFoundError{Path: "upload session " + id}
	}
	return doc.toSession(), nil
}

func (r *UploadSessionRepository) FindByFingerprint(owner, fingerprint string) (*models.UploadSession, error) {
	if fingerprint == "" {
		return nil, &domainerrors.NotFoundError{Path: "upload fingerprint"}
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	docs, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	var best *uploadSessionDocument
	for _, doc := range docs {
		if doc.Owner != owner || doc.Fingerprint != fingerprint {
			continue
		}
		if time.Now().After(doc.ExpiresAt) {
			continue
		}
		if best == nil || doc.UpdatedAt.After(best.UpdatedAt) {
			candidate := doc
			best = &candidate
		}
	}
	if best == nil {
		return nil, &domainerrors.NotFoundError{Path: "upload fingerprint"}
	}
	return best.toSession(), nil
}

func (r *UploadSessionRepository) List(owner string) ([]*models.UploadSession, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	docs, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	out := make([]*models.UploadSession, 0, len(docs))
	for _, doc := range docs {
		if doc.Owner != owner {
			continue
		}
		if now.After(doc.ExpiresAt) {
			continue
		}
		out = append(out, doc.toSession())
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

func (r *UploadSessionRepository) UpdateOffset(id string, offset int64, updatedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	docs, err := r.loadLocked()
	if err != nil {
		return err
	}
	doc, ok := docs[id]
	if !ok {
		return &domainerrors.NotFoundError{Path: "upload session " + id}
	}
	doc.Offset = offset
	doc.UpdatedAt = updatedAt
	docs[id] = doc
	return r.saveLocked(docs)
}

func (r *UploadSessionRepository) OpenStaging(id string, write bool) (io.ReadWriteCloser, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureStagingDir(); err != nil {
		return nil, err
	}
	flags := os.O_RDONLY
	if write {
		flags = os.O_RDWR
	}
	return os.OpenFile(r.partPath(id), flags, 0o600)
}

func (r *UploadSessionRepository) Append(id string, offset int64, body io.Reader) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureStagingDir(); err != nil {
		return 0, err
	}
	f, err := os.OpenFile(r.partPath(id), os.O_RDWR, 0o600)
	if err != nil {
		return 0, fmt.Errorf("open staging file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat staging file: %w", err)
	}
	if offset != info.Size() {
		return info.Size(), fmt.Errorf("offset mismatch: got %d want %d", offset, info.Size())
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return 0, fmt.Errorf("seek staging file: %w", err)
	}
	written, err := io.Copy(f, body)
	if err != nil {
		return info.Size(), fmt.Errorf("append staging: %w", err)
	}
	if err := f.Sync(); err != nil {
		return info.Size(), fmt.Errorf("sync staging: %w", err)
	}
	return info.Size() + written, nil
}

func (r *UploadSessionRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	docs, err := r.loadLocked()
	if err != nil {
		return err
	}
	if _, ok := docs[id]; !ok {
		return &domainerrors.NotFoundError{Path: "upload session " + id}
	}
	delete(docs, id)
	_ = os.Remove(r.partPath(id))
	return r.saveLocked(docs)
}

func (r *UploadSessionRepository) PurgeExpired(now time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	docs, err := r.loadLocked()
	if err != nil {
		return 0, err
	}
	removed := 0
	for id, doc := range docs {
		if now.After(doc.ExpiresAt) {
			delete(docs, id)
			_ = os.Remove(r.partPath(id))
			removed++
		}
	}
	if removed == 0 {
		return 0, nil
	}
	return removed, r.saveLocked(docs)
}

func (r *UploadSessionRepository) partPath(id string) string {
	return filepath.Join(r.staging, id+".part")
}

func (r *UploadSessionRepository) ensureStagingDir() error {
	return os.MkdirAll(r.staging, storageDirPerm)
}

func (r *UploadSessionRepository) loadLocked() (map[string]uploadSessionDocument, error) {
	data, err := os.ReadFile(r.metaPath)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]uploadSessionDocument{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read upload sessions file: %w", err)
	}
	var docs map[string]uploadSessionDocument
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("parse upload sessions file: %w", err)
	}
	if docs == nil {
		docs = map[string]uploadSessionDocument{}
	}
	return docs, nil
}

func (r *UploadSessionRepository) saveLocked(docs map[string]uploadSessionDocument) error {
	if err := os.MkdirAll(filepath.Dir(r.metaPath), storageDirPerm); err != nil {
		return fmt.Errorf("create upload sessions dir: %w", err)
	}
	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode upload sessions: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(r.metaPath), ".upload-sessions-*.tmp")
	if err != nil {
		return fmt.Errorf("create upload sessions temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write upload sessions temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync upload sessions temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close upload sessions temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), r.metaPath); err != nil {
		return fmt.Errorf("replace upload sessions file: %w", err)
	}
	return nil
}

func newSessionID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate upload session id: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}
