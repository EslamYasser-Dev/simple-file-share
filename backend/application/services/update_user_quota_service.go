package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// UpdateUserQuotaService sets an account's storage quota. Only quota.manage
// (or auth-disabled nil user) may change quotas; the gate lives here so both
// primary adapters enforce it.
type UpdateUserQuotaService struct {
	users  ports.UserRepository
	index  ports.FileIndexRepository
	scoper ports.PathScoper
	bus    *events.Bus
	roles  *RoleCatalog
}

func NewUpdateUserQuotaService(users ports.UserRepository, index ports.FileIndexRepository, scoper ports.PathScoper, roles *RoleCatalog) *UpdateUserQuotaService {
	return &UpdateUserQuotaService{users: users, index: index, scoper: scoper, roles: roles}
}

// SetEventBus attaches a live-update bus (nil disables publishing).
func (s *UpdateUserQuotaService) SetEventBus(bus *events.Bus) { s.bus = bus }

// Execute parses the quota (bytes or a human size, 0/unlimited meaning no
// limit), stores it, and returns the updated account with current usage.
func (s *UpdateUserQuotaService) Execute(actor *models.User, username, rawQuota string) (*models.UserStats, error) {
	if !actor.HasPermission(models.PermQuotaManage, s.roles) {
		return nil, &domainerrors.ForbiddenError{Action: "set quota"}
	}

	quota, err := valueobjects.ParseStorageQuota(rawQuota)
	if err != nil {
		return nil, domainerrors.NewValidationError("quota", rawQuota, err.Error())
	}

	if err := s.users.SetQuotaBytes(username, quota.Bytes()); err != nil {
		return nil, err
	}

	stats, err := s.loadStats(username)
	if err == nil {
		publishEvent(s.bus, events.TypeQuota, "", actor)
	}
	return stats, err
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
		Role:       u.Role,
		IsAdmin:    u.IsAdmin,
		Enabled:    u.Enabled,
		QuotaBytes: u.QuotaBytes,
		CreatedAt:  u.CreatedAt,
		Files:      files,
		Size:       size,
	}, nil
}
