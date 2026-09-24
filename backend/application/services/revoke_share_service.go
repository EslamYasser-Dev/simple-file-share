package services

import (
	"errors"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// RevokeShareService invalidates a share link. Only the owner (or an account
// with shares.manage / system view) may revoke a link.
type RevokeShareService struct {
	shareRepo ports.ShareRepository
	scoper    ports.PathScoper
	bus       *events.Bus
	roles     *RoleCatalog
}

func NewRevokeShareService(shareRepo ports.ShareRepository, scoper ports.PathScoper, roles *RoleCatalog) *RevokeShareService {
	return &RevokeShareService{shareRepo: shareRepo, scoper: scoper, roles: roles}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *RevokeShareService) SetEventBus(bus *events.Bus) { s.bus = bus }

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

	if user == nil || user.IsSystemView() || user.HasPermission(models.PermSharesManage, s.roles) {
		if err := s.shareRepo.Delete(token); err != nil {
			return err
		}
		publishEvent(s.bus, events.TypeShareRevoke, share.Path, user)
		return nil
	}
	if share.Owner != user.Username {
		return &domainerrors.ForbiddenError{Action: "revoke share", Path: share.Token}
	}
	if err := s.shareRepo.Delete(token); err != nil {
		return err
	}
	publishEvent(s.bus, events.TypeShareRevoke, share.Path, user)
	return nil
}
