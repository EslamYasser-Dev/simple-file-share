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

// CreateShareService generates an unguessable public link to a file or
// directory the user is allowed to read, with a caller-chosen validity window.
type CreateShareService struct {
	fileRepo  ports.FileRepository
	shareRepo ports.ShareRepository
	scoper    ports.PathScoper
	now       func() time.Time
	tokenLen  int
	bus       *events.Bus
}

func NewCreateShareService(fileRepo ports.FileRepository, shareRepo ports.ShareRepository, scoper ports.PathScoper) *CreateShareService {
	return &CreateShareService{
		fileRepo:  fileRepo,
		shareRepo: shareRepo,
		scoper:    scoper,
		now:       time.Now,
		tokenLen:  32,
	}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *CreateShareService) SetEventBus(bus *events.Bus) { s.bus = bus }

// Execute creates a share for the given virtual path. expiresInSeconds of 0
// means the link never expires. Returns the created share.
func (s *CreateShareService) Execute(user *models.User, path string, expiresInSeconds int64) (*models.Share, error) {
	if expiresInSeconds < 0 {
		return nil, domainerrors.NewValidationError("expiresInSeconds", expiresInSeconds, "must be zero or positive")
	}
	if time.Duration(expiresInSeconds)*time.Second > maxShareValidity {
		return nil, domainerrors.NewValidationError("expiresInSeconds", expiresInSeconds, "validity exceeds the maximum allowed")
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
		Token:     token,
		Path:      virtual,
		Owner:     owner,
		CreatedAt: now,
	}
	if user != nil {
		share.IsAdmin = user.IsAdmin
	}
	if expiresInSeconds > 0 {
		share.ExpiresAt = now.Add(time.Duration(expiresInSeconds) * time.Second)
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
