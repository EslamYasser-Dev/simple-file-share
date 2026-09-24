package services

import (
	"errors"
	"io"
	"path"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// DefaultResumableChunkSize is the preferred client chunk size advertised by
// Start. Chunks may be any size as long as they arrive at the expected offset.
const DefaultResumableChunkSize int64 = 4 << 20 // 4 MiB

// DefaultUploadSessionTTL is how long an idle session stays resumable.
const DefaultUploadSessionTTL = 24 * time.Hour

// ResumableUploadService implements chunked uploads that survive network
// drops and page reloads. Sessions stage bytes under the metadata directory;
// Complete promotes the staged file through the normal FileRepository path so
// versions, quota accounting, and the search index stay consistent.
type ResumableUploadService struct {
	sessions  ports.UploadSessionRepository
	fileRepo  ports.FileRepository
	scoper    ports.PathScoper
	index     ports.FileIndexRepository
	users     ports.UserRepository
	maxBytes  int64
	chunkSize int64
	ttl       time.Duration
	bus       *events.Bus
	now       func() time.Time
}

func NewResumableUploadService(
	sessions ports.UploadSessionRepository,
	fileRepo ports.FileRepository,
	scoper ports.PathScoper,
	index ports.FileIndexRepository,
	users ports.UserRepository,
	maxBytes int64,
) *ResumableUploadService {
	return &ResumableUploadService{
		sessions:  sessions,
		fileRepo:  fileRepo,
		scoper:    scoper,
		index:     index,
		users:     users,
		maxBytes:  maxBytes,
		chunkSize: DefaultResumableChunkSize,
		ttl:       DefaultUploadSessionTTL,
		now:       time.Now,
	}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *ResumableUploadService) SetEventBus(bus *events.Bus) { s.bus = bus }

// ChunkSize returns the preferred chunk size advertised to clients.
func (s *ResumableUploadService) ChunkSize() int64 { return s.chunkSize }

func ownerKey(user *models.User) string {
	if user == nil {
		return ""
	}
	return user.Username
}

// Start creates a resumable session, or reuses an active one that matches
// owner+fingerprint so a reload resumes from the staged offset.
func (s *ResumableUploadService) Start(user *models.User, destination, filename, fingerprint string, size int64) (*models.UploadSession, error) {
	if filename == "" {
		return nil, domainerrors.NewValidationError("filename", filename, "filename is required")
	}
	if size < 0 {
		return nil, domainerrors.NewValidationError("size", size, "size must be non-negative")
	}
	if fingerprint != "" {
		if existing, err := s.sessions.FindByFingerprint(ownerKey(user), fingerprint); err == nil {
			if existing.Size == size && existing.Destination == destination && existing.Filename == filename {
				return existing, nil
			}
		} else if !isSessionNotFound(err) {
			return nil, err
		}
	}

	virtual := filename
	if destination != "" {
		virtual = path.Join(destination, filename)
	}
	fp, err := valueobjects.NewFilePath(virtual)
	if err != nil {
		return nil, err
	}
	if _, err := s.scoper.WritePath(user, fp.Relative()); err != nil {
		return nil, err
	}
	if s.maxBytes > 0 && size > s.maxBytes {
		return nil, domainerrors.NewValidationError("size", size, "upload exceeds size limit")
	}
	if err := s.ensureQuota(user, size); err != nil {
		return nil, err
	}

	now := s.now()
	session := &models.UploadSession{
		Owner:       ownerKey(user),
		Fingerprint: fingerprint,
		Destination: destination,
		Filename:    filename,
		Size:        size,
		Offset:      0,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   now.Add(s.ttl),
	}
	if err := s.sessions.Create(session); err != nil {
		return nil, err
	}
	return session, nil
}

// Status returns the current session (for resume probes).
func (s *ResumableUploadService) Status(user *models.User, id string) (*models.UploadSession, error) {
	return s.ownedSession(user, id)
}

// List returns the caller's active sessions (pending / resumable uploads),
// newest first. Expired sessions are omitted (and left for PurgeExpired).
func (s *ResumableUploadService) List(user *models.User) ([]*models.UploadSession, error) {
	all, err := s.sessions.List(ownerKey(user))
	if err != nil {
		return nil, err
	}
	now := s.now()
	out := make([]*models.UploadSession, 0, len(all))
	for _, session := range all {
		if now.After(session.ExpiresAt) {
			continue
		}
		out = append(out, session)
	}
	return out, nil
}

// Append writes one chunk at the expected offset. A mismatched offset returns
// a conflict-style validation error carrying the true staged size via
// ErrOffsetMismatch so clients can resynchronize.
func (s *ResumableUploadService) Append(user *models.User, id string, offset int64, body io.Reader) (*models.UploadSession, error) {
	session, err := s.ownedSession(user, id)
	if err != nil {
		return nil, err
	}
	if offset != session.Offset {
		return nil, &ErrOffsetMismatch{Expected: session.Offset, Got: offset}
	}
	if session.Offset >= session.Size && session.Size > 0 {
		return nil, domainerrors.NewValidationError("offset", offset, "upload already complete; call complete")
	}
	if s.maxBytes > 0 && session.Size > s.maxBytes {
		return nil, domainerrors.NewValidationError("size", session.Size, "upload exceeds size limit")
	}

	newOffset, err := s.sessions.Append(id, offset, body)
	if err != nil {
		// The repository reports the true staged size on offset conflicts.
		if mismatch := asOffsetMismatch(err); mismatch != nil {
			return nil, mismatch
		}
		return nil, err
	}
	if newOffset > session.Size {
		_ = s.sessions.Delete(id)
		return nil, domainerrors.NewValidationError("size", newOffset, "upload exceeds declared size")
	}

	now := s.now()
	if err := s.sessions.UpdateOffset(id, newOffset, now); err != nil {
		return nil, err
	}
	session.Offset = newOffset
	session.UpdatedAt = now
	return session, nil
}

// Complete promotes the fully staged file into the destination namespace.
func (s *ResumableUploadService) Complete(user *models.User, id string) (*models.FileUpload, error) {
	session, err := s.ownedSession(user, id)
	if err != nil {
		return nil, err
	}
	if session.Offset != session.Size {
		return nil, &ErrOffsetMismatch{Expected: session.Size, Got: session.Offset}
	}
	if err := s.ensureQuota(user, session.Size); err != nil {
		return nil, err
	}

	virtual := session.Filename
	if session.Destination != "" {
		virtual = path.Join(session.Destination, session.Filename)
	}
	fp, err := valueobjects.NewFilePath(virtual)
	if err != nil {
		return nil, err
	}
	physical, err := s.scoper.WritePath(user, fp.Relative())
	if err != nil {
		return nil, err
	}
	if err := s.ensureParent(physical); err != nil {
		return nil, err
	}

	staged, err := s.sessions.OpenStaging(id, false)
	if err != nil {
		return nil, err
	}
	written, writeErr := s.fileRepo.WriteFile(physical, staged)
	staged.Close()
	_ = s.sessions.Delete(id)
	if writeErr != nil {
		return nil, writeErr
	}
	if written != session.Size {
		return nil, domainerrors.NewValidationError("size", written, "staged size mismatch")
	}

	virtualName := s.scoper.PhysicalToVirtual(user, physical)
	result := &models.FileUpload{Filename: virtualName, Size: written}
	publishEventBytes(s.bus, events.TypeUpload, virtualName, user, written)
	return result, nil
}

// Abort discards a session and its staged bytes.
func (s *ResumableUploadService) Abort(user *models.User, id string) error {
	if _, err := s.ownedSession(user, id); err != nil {
		return err
	}
	return s.sessions.Delete(id)
}

// PurgeExpired removes sessions past their TTL (callers wire this at startup
// and on a background ticker).
func (s *ResumableUploadService) PurgeExpired() (int, error) {
	return s.sessions.PurgeExpired(s.now())
}

func (s *ResumableUploadService) ownedSession(user *models.User, id string) (*models.UploadSession, error) {
	session, err := s.sessions.Get(id)
	if err != nil {
		return nil, err
	}
	if session.Owner != ownerKey(user) {
		// Same not-found as an unknown id: never confirm another account's session.
		return nil, &domainerrors.NotFoundError{Path: "upload session " + id}
	}
	if s.now().After(session.ExpiresAt) {
		_ = s.sessions.Delete(id)
		return nil, &domainerrors.NotFoundError{Path: "upload session " + id}
	}
	return session, nil
}

func (s *ResumableUploadService) ensureParent(physical string) error {
	dir := path.Dir(path.Clean(physical))
	if dir == "." || dir == "/" {
		return nil
	}
	return s.fileRepo.CreateDirectory(dir)
}

// ensureQuota rejects a session whose declared size does not fit the
// remaining account quota (quota <= 0 means unlimited; nil user is system view).
func (s *ResumableUploadService) ensureQuota(user *models.User, size int64) error {
	if user == nil || size <= 0 {
		return nil
	}
	quota, err := s.users.GetQuotaBytes(user.Username)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil
		}
		return err
	}
	if quota <= 0 {
		return nil
	}
	_, used, err := s.index.PrefixStats(s.scoper.PrivatePrefix(user.Username))
	if err != nil {
		return err
	}
	if used+size > quota {
		return &domainerrors.QuotaExceededError{Action: "upload", Path: ""}
	}
	return nil
}

// ErrOffsetMismatch reports that a chunk arrived at the wrong position; the
// client should resume from Expected.
type ErrOffsetMismatch struct {
	Expected int64
	Got      int64
}

func (e *ErrOffsetMismatch) Error() string {
	return "upload offset mismatch"
}

func asOffsetMismatch(err error) *ErrOffsetMismatch {
	var mismatch *ErrOffsetMismatch
	if errors.As(err, &mismatch) {
		return mismatch
	}
	return nil
}

func isSessionNotFound(err error) bool {
	var notFound *domainerrors.NotFoundError
	return errors.As(err, &notFound) || errors.Is(err, domainerrors.ErrNotFound)
}
