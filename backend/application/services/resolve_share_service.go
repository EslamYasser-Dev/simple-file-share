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
// must validate the token's existence, validity window, password, and
// download budget before anything is streamed.
type ResolveShareService struct {
	shareRepo ports.ShareRepository
	scoper    ports.PathScoper
	downloads *DownloadService
	hasher    ports.PasswordHasher
	now       func() time.Time
}

func NewResolveShareService(shareRepo ports.ShareRepository, scoper ports.PathScoper, downloads *DownloadService, hasher ports.PasswordHasher) *ResolveShareService {
	return &ResolveShareService{
		shareRepo: shareRepo,
		scoper:    scoper,
		downloads: downloads,
		hasher:    hasher,
		now:       time.Now,
	}
}

// Execute resolves token to a download. password is the presented link
// password (from X-Share-Password or the password query parameter); links
// without a password ignore it.
func (s *ResolveShareService) Execute(token, password string) (*models.Download, error) {
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

	if share.PasswordProtected() {
		if password == "" {
			return nil, &domainerrors.SharePasswordError{Token: token, Missing: true}
		}
		if !s.hasher.Verify(password, share.PasswordHash) {
			return nil, &domainerrors.SharePasswordError{Token: token}
		}
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
	download, err := s.downloads.Execute(owner, share.Path)
	if err != nil {
		return nil, err
	}

	// Charge the budget only once the file is known to be servable, and
	// atomically so concurrent requests cannot overshoot the limit.
	if share.Limited() {
		if err := s.shareRepo.ConsumeDownload(token); err != nil {
			if errors.Is(err, domainerrors.ErrShareLimitReached) {
				_ = download.Stream.Close()
				return nil, err
			}
			_ = download.Stream.Close()
			return nil, err
		}
	}
	return download, nil
}
