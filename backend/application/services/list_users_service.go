package services

import (
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// ListUsersService returns every account plus per-user storage usage for the
// admin console.
type ListUsersService struct {
	users  ports.UserRepository
	index  ports.FileIndexRepository
	scoper ports.PathScoper
}

func NewListUsersService(users ports.UserRepository, index ports.FileIndexRepository, scoper ports.PathScoper) *ListUsersService {
	return &ListUsersService{users: users, index: index, scoper: scoper}
}

func (s *ListUsersService) Execute() ([]models.UserStats, error) {
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
			Username:  u.Username,
			IsAdmin:   u.IsAdmin,
			CreatedAt: u.CreatedAt,
			Files:     files,
			Size:      size,
		})
	}
	return stats, nil
}
