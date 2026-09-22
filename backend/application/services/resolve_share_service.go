package services

import (
	"errors"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// ResolveShareService turns a public share token into a streamable download. It
// is the only service that serves content without an authenticated user, so it
// must validate both the token's existence and its validity window.
type ResolveShareService struct {
	shareRepo ports.ShareRepository
	scoper    ports.PathScoper
	downloads *DownloadService
	now       func() time.Time
}

func NewResolveShareService(shareRepo ports.ShareRepository, scoper ports.PathScoper, downloads *DownloadService) *ResolveShareService {
	return &ResolveShareService{
		shareRepo: shareRepo,
		scoper:    scoper,
		downloads: downloads,
		now:       time.Now,
	}
}

func (s *ResolveShareService) Execute(token string) (*models.Download, error) {
	// Reject structurally invalid tokens before anything touches storage, so
	// crafted garbage can never cost a repository or filesystem probe.
	if _, err := valueobjects.NewShareToken(token); err != nil {
		return nil, &domainerrors.ShareNotFoundError{Token: token}
	}

	share, err := s.shareRepo.FindByToken(token)
	if err != nil {
		if errors.Is(err, domainerrors.ErrShareNotFound) {
			return nil, &domainerrors.ShareNotFoundError{Token: token}
		}
		return nil, err
	}

	if share.Expired(s.now()) {
		return nil, &domainerrors.ShareExpiredError{Token: token}
	}

	// Replay the owner's read scope so the stored virtual path resolves to the
	// same physical location it pointed at when the link was created. A share
	// created by the system view (no owner) stays unscoped. The explicit scope
	// check here also guards against a tampered/legacy path in the store.
	owner := (*models.User)(nil)
	if share.Owner != "" {
		owner = &models.User{Username: share.Owner, IsAdmin: share.IsAdmin}
	}
	if _, err := s.scoper.ReadPath(owner, share.Path); err != nil {
		return nil, err
	}

	// If the underlying item was deleted, this returns a 404.
	return s.downloads.Execute(owner, share.Path)
}
