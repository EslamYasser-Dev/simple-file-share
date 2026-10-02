package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
)

// maxShareValidity bounds the allowed link lifetime so a drifting clock or a
// malformed request can never produce a wildly offset expiration. "Never
// expire" (0) remains supported via noExpirySentinel.
const maxShareValidity = 3650 * 24 * time.Hour

// SharePolicy carries the optional restrictions applied when minting a link.
type SharePolicy struct {
	// ExpiresInSeconds is the link validity; 0 means never expires.
	ExpiresInSeconds int64
	// Password protects the link (checked at serve time). Empty = public.
	Password string
	// MaxDownloads caps serves; 0 = unlimited.
	MaxDownloads int
}

// maxSharePassword bounds stored link passwords (PBKDF2 accepts anything,
// but unbounded input is an easy memory-amplification vector).
const maxSharePassword = 256

// CreateShareService generates an unguessable public link to a file or
// directory the user is allowed to read, with caller-chosen policies.
type CreateShareService struct {
	fileRepo  ports.FileRepository
	shareRepo ports.ShareRepository
	scoper    ports.PathScoper
	hasher    ports.PasswordHasher
	now       func() time.Time
	tokenLen  int
	bus       *events.Bus
}

func NewCreateShareService(fileRepo ports.FileRepository, shareRepo ports.ShareRepository, scoper ports.PathScoper, hasher ports.PasswordHasher) *CreateShareService {
	return &CreateShareService{
		fileRepo:  fileRepo,
		shareRepo: shareRepo,
		scoper:    scoper,
		hasher:    hasher,
		now:       time.Now,
		tokenLen:  32,
	}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *CreateShareService) SetEventBus(bus *events.Bus) { s.bus = bus }

// Execute creates a share for the given virtual path under the given policy.
// Returns the created share.
func (s *CreateShareService) Execute(user *models.User, path string, policy SharePolicy) (*models.Share, error) {
	if policy.ExpiresInSeconds < 0 {
		return nil, domainerrors.NewValidationError("expiresInSeconds", policy.ExpiresInSeconds, "must be zero or positive")
	}
	if time.Duration(policy.ExpiresInSeconds)*time.Second > maxShareValidity {
		return nil, domainerrors.NewValidationError("expiresInSeconds", policy.ExpiresInSeconds, "validity exceeds the maximum allowed")
	}
	if policy.MaxDownloads < 0 {
		return nil, domainerrors.NewValidationError("maxDownloads", policy.MaxDownloads, "must be zero or positive")
	}
	if len(policy.Password) > maxSharePassword {
		return nil, domainerrors.NewValidationError("password", nil, "must be at most 256 characters")
	}

	fp, err := valueobjects.NewFilePath(path)
	if err != nil {
		return nil, err
	}

	owner := ""
	if user != nil {
		owner = user.Username
	}
	// The share records the *virtual* path the client already knows, so the UI
	// can match links to items directly. The physical path is re-derived from
	// that virtual path (scoped by the same owner) every time the link is served.
	virtual := fp.Relative()
	physical, err := s.scoper.ReadPath(user, virtual)
	if err != nil {
		return nil, err
	}

	exists, err := s.fileRepo.FileExists(physical)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &domainerrors.NotFoundError{Path: path}
	}

	token, err := s.newToken()
	if err != nil {
		return nil, err
	}

	now := s.now()
	share := &models.Share{
		Token:        token,
		Path:         virtual,
		Owner:        owner,
		CreatedAt:    now,
		MaxDownloads: policy.MaxDownloads,
	}
	if user != nil {
		share.IsAdmin = user.IsAdmin
	}
	if policy.ExpiresInSeconds > 0 {
		share.ExpiresAt = now.Add(time.Duration(policy.ExpiresInSeconds) * time.Second)
	}
	if policy.Password != "" {
		hash, err := s.hasher.Hash(policy.Password)
		if err != nil {
			return nil, err
		}
		share.PasswordHash = hash
	}

	if err := s.shareRepo.Create(share); err != nil {
		return nil, err
	}
	publishEvent(s.bus, events.TypeShare, virtual, user)
	return share, nil
}

// newToken returns a cryptographically random, URL-safe token. Collisions are
// astronomically unlikely, but they are retried if the repository rejects one.
func (s *CreateShareService) newToken() (string, error) {
	for attempt := 0; attempt < 3; attempt++ {
		raw := make([]byte, s.tokenLen)
		if _, err := rand.Read(raw); err != nil {
			return "", err
		}
		token := base64.RawURLEncoding.EncodeToString(raw)
		if _, err := s.shareRepo.FindByToken(token); err != nil {
			if errors.Is(err, domainerrors.ErrShareNotFound) {
				return token, nil
			}
			return "", err
		}
	}
	return "", errors.New("could not generate a unique share token")
}
