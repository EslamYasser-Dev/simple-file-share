package services

import (
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// UpdateUserQuotaService sets an account's storage quota. Only the system view
// (admin, or auth-disabled nil user) may change quotas; the gate lives here so
// both primary adapters enforce it.
type UpdateUserQuotaService struct {
	users  ports.UserRepository
	index  ports.FileIndexRepository
	scoper ports.PathScoper
}

func NewUpdateUserQuotaService(users ports.UserRepository, index ports.FileIndexRepository, scoper ports.PathScoper) *UpdateUserQuotaService {
	return &UpdateUserQuotaService{users: users, index: index, scoper: scoper}
}

// Execute parses the quota (bytes or a human size, 0/unlimited meaning no
// limit), stores it, and returns the updated account with current usage.
func (s *UpdateUserQuotaService) Execute(actor *models.User, username, rawQuota string) (*models.UserStats, error) {
	if !actor.IsSystemView() {
		return nil, &domainerrors.ForbiddenError{Action: "set quota"}
	}

	quota, err := valueobjects.ParseStorageQuota(rawQuota)
	if err != nil {
		return nil, domainerrors.NewValidationError("quota", rawQuota, err.Error())
	}

	if err := s.users.SetQuotaBytes(username, quota.Bytes()); err != nil {
		return nil, err
	}

	return s.loadStats(username)
}

// loadStats builds the read model for one account including storage usage.
func (s *UpdateUserQuotaService) loadStats(username string) (*models.UserStats, error) {
	u, err := s.users.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	files, size, err := s.index.PrefixStats(s.scoper.PrivatePrefix(username))
	if err != nil {
		return nil, err
	}
	return &models.UserStats{
		Username:   u.Username,
		IsAdmin:    u.IsAdmin,
		QuotaBytes: u.QuotaBytes,
		CreatedAt:  u.CreatedAt,
		Files:      files,
		Size:       size,
	}, nil
}
