package services

import (
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// ListUsersService returns every account plus per-user storage usage for the
// admin console. Only the system view (admin, or auth-disabled nil user) may
// list accounts; the check lives here so both primary adapters enforce it.
type ListUsersService struct {
	users  ports.UserRepository
	index  ports.FileIndexRepository
	scoper ports.PathScoper
}

func NewListUsersService(users ports.UserRepository, index ports.FileIndexRepository, scoper ports.PathScoper) *ListUsersService {
	return &ListUsersService{users: users, index: index, scoper: scoper}
}

func (s *ListUsersService) Execute(user *models.User) ([]models.UserStats, error) {
	if !user.IsSystemView() {
		return nil, &domainerrors.ForbiddenError{Action: "list users"}
	}

	list, err := s.users.ListUsers()
	if err != nil {
		return nil, err
	}

	stats := make([]models.UserStats, 0, len(list))
	for _, u := range list {
		files, size, err := s.index.PrefixStats(s.scoper.PrivatePrefix(u.Username))
		if err != nil {
			return nil, err
		}
		stats = append(stats, models.UserStats{
			Username:   u.Username,
			IsAdmin:    u.IsAdmin,
			QuotaBytes: u.QuotaBytes,
			CreatedAt:  u.CreatedAt,
			Files:      files,
			Size:       size,
		})
	}
	return stats, nil
}
