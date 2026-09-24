package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// ListSharesService returns the share links visible to a user: an account with
// shares.manage (or the system view) sees every link; others only their own.
type ListSharesService struct {
	shareRepo ports.ShareRepository
	scoper    ports.PathScoper
	roles     *RoleCatalog
}

func NewListSharesService(shareRepo ports.ShareRepository, scoper ports.PathScoper, roles *RoleCatalog) *ListSharesService {
	return &ListSharesService{shareRepo: shareRepo, scoper: scoper, roles: roles}
}

func (s *ListSharesService) Execute(user *models.User) ([]*models.Share, error) {
	if user == nil || user.IsSystemView() || user.HasPermission(models.PermSharesManage, s.roles) {
		return s.shareRepo.ListAll()
	}
	return s.shareRepo.ListByOwner(user.Username)
}
