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

// UserFileRepository persists user accounts as a JSON document under the
// storage root's metadata directory (`.file-share/users.json`), which the
// search index already skips. Writes are atomic (temp file + rename) and
// guarded by an in-process mutex.
type UserFileRepository struct {
	path string
	mu   sync.RWMutex
}

var _ ports.UserRepository = (*UserFileRepository)(nil)

func NewUserFileRepository(rootDir string) *UserFileRepository {
	return &UserFileRepository{
		path: filepath.Join(rootDir, ".file-share", "users.json"),
	}
}

type userDocument struct {
	Username      string    `json:"username"`
	PasswordHash  string    `json:"passwordHash"`
	IsAdmin       bool      `json:"isAdmin"`
	QuotaBytes    int64     `json:"quotaBytes,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	OAuthProvider string    `json:"oauthProvider,omitempty"`
	OAuthSubject  string    `json:"oauthSubject,omitempty"`
}

func (d userDocument) toUser() *models.User {
	return &models.User{
		Username:      d.Username,
		PasswordHash:  d.PasswordHash,
		IsAdmin:       d.IsAdmin,
		QuotaBytes:    d.QuotaBytes,
		CreatedAt:     d.CreatedAt,
		OAuthProvider: d.OAuthProvider,
		OAuthSubject:  d.OAuthSubject,
	}
}

func namedDoc(u *models.User) userDocument {
	return userDocument{
		Username:      u.Username,
		PasswordHash:  u.PasswordHash,
		IsAdmin:       u.IsAdmin,
		QuotaBytes:    u.QuotaBytes,
		CreatedAt:     u.CreatedAt,
		OAuthProvider: u.OAuthProvider,
		OAuthSubject:  u.OAuthSubject,
	}
}

func (r *UserFileRepository) CreateUser(user *models.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	docs, err := r.loadLocked()
	if err != nil {
		return err
	}
	if _, ok := docs[user.Username]; ok {
		return domainerrors.ErrUserAlreadyExists
	}

	docs[user.Username] = namedDoc(user)
	return r.saveLocked(docs)
}

func (r *UserFileRepository) FindByUsername(username string) (*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	docs, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	doc, ok := docs[username]
	if !ok {
		return nil, domainerrors.ErrUserNotFound
	}
	return doc.toUser(), nil
}

func (r *UserFileRepository) FindByOAuth(provider, subject string) (*models.User, error) {
	if provider == "" || subject == "" {
		return nil, domainerrors.ErrUserNotFound
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	docs, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	for _, doc := range docs {
		if doc.OAuthProvider == provider && doc.OAuthSubject == subject {
			return doc.toUser(), nil
		}
	}
	return nil, domainerrors.ErrUserNotFound
}

func (r *UserFileRepository) ListUsers() ([]*models.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	docs, err := r.loadLocked()
	if err != nil {
		return nil, err
	}
	users := make([]*models.User, 0, len(docs))
	for _, doc := range docs {
		users = append(users, doc.toUser())
	}
	sort.Slice(users, func(i, j int) bool { return users[i].Username < users[j].Username })
	return users, nil
}

func (r *UserFileRepository) CountUsers() (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	docs, err := r.loadLocked()
	if err != nil {
		return 0, err
	}
	return len(docs), nil
}

func (r *UserFileRepository) GetQuotaBytes(username string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	docs, err := r.loadLocked()
	if err != nil {
		return 0, err
	}
	doc, ok := docs[username]
	if !ok {
		return 0, domainerrors.ErrUserNotFound
	}
	return doc.QuotaBytes, nil
}

func (r *UserFileRepository) SetQuotaBytes(username string, quotaBytes int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	docs, err := r.loadLocked()
	if err != nil {
		return err
	}
	doc, ok := docs[username]
	if !ok {
		return domainerrors.ErrUserNotFound
	}
	doc.QuotaBytes = quotaBytes
	docs[username] = doc
	return r.saveLocked(docs)
}

func (r *UserFileRepository) loadLocked() (map[string]userDocument, error) {
	data, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]userDocument{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read users file: %w", err)
	}

	var docs map[string]userDocument
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("parse users file: %w", err)
	}
	if docs == nil {
		docs = map[string]userDocument{}
	}
	return docs, nil
}

func (r *UserFileRepository) saveLocked(docs map[string]userDocument) error {
	if err := os.MkdirAll(filepath.Dir(r.path), storageDirPerm); err != nil {
		return fmt.Errorf("create users dir: %w", err)
	}

	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return fmt.Errorf("encode users: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(r.path), ".users-*.tmp")
	if err != nil {
		return fmt.Errorf("create users temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write users temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync users temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close users temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), r.path); err != nil {
		return fmt.Errorf("replace users file: %w", err)
	}
	return nil
}
