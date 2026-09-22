package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// ListSharesService returns the share links visible to a user: an admin sees
// every link in the system, a regular user only their own.
type ListSharesService struct {
	shareRepo ports.ShareRepository
	scoper    ports.PathScoper
}

func NewListSharesService(shareRepo ports.ShareRepository, scoper ports.PathScoper) *ListSharesService {
	return &ListSharesService{shareRepo: shareRepo, scoper: scoper}
}

func (s *ListSharesService) Execute(user *models.User) ([]*models.Share, error) {
	if user == nil || user.IsAdmin {
		return s.shareRepo.ListAll()
	}
	return s.shareRepo.ListByOwner(user.Username)
}
