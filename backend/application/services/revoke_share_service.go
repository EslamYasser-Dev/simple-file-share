package services

import (
	"errors"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// RevokeShareService invalidates a share link. Only the owner (or an admin)
// may revoke a link.
type RevokeShareService struct {
	shareRepo ports.ShareRepository
	scoper    ports.PathScoper
}

func NewRevokeShareService(shareRepo ports.ShareRepository, scoper ports.PathScoper) *RevokeShareService {
	return &RevokeShareService{shareRepo: shareRepo, scoper: scoper}
}

func (s *RevokeShareService) Execute(user *models.User, token string) error {
	if _, err := valueobjects.NewShareToken(token); err != nil {
		return &domainerrors.ShareNotFoundError{Token: token}
	}

	share, err := s.shareRepo.FindByToken(token)
	if err != nil {
		if errors.Is(err, domainerrors.ErrShareNotFound) {
			return &domainerrors.ShareNotFoundError{Token: token}
		}
		return err
	}

	if user == nil || user.IsAdmin {
		return s.shareRepo.Delete(token)
	}
	if share.Owner != user.Username {
		return &domainerrors.ForbiddenError{Action: "revoke share", Path: share.Token}
	}
	return s.shareRepo.Delete(token)
}
