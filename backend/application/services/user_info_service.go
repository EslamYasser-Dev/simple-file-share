package services

import (
	"errors"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// UserInfoService assembles the authenticated account's profile plus its
// storage quota and usage for the /me endpoint. With auth disabled the actor is
// nil and the service reports the anonymous system view.
type UserInfoService struct {
	users  ports.UserRepository
	index  ports.FileIndexRepository
	scoper ports.PathScoper
}

func NewUserInfoService(users ports.UserRepository, index ports.FileIndexRepository, scoper ports.PathScoper) *UserInfoService {
	return &UserInfoService{users: users, index: index, scoper: scoper}
}

func (s *UserInfoService) Execute(user *models.User) (*models.UserStats, error) {
	if user == nil {
		return &models.UserStats{IsAdmin: true}, nil
	}

	quota, err := s.users.GetQuotaBytes(user.Username)
	if err != nil {
		// A vanished account record still resolves to an unlimited, empty space.
		if !errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, err
		}
		quota = 0
	}
	files, size, err := s.index.PrefixStats(s.scoper.PrivatePrefix(user.Username))
	if err != nil {
		return nil, err
	}
	return &models.UserStats{
		Username:   user.Username,
		IsAdmin:    user.IsAdmin,
		QuotaBytes: quota,
		CreatedAt:  user.CreatedAt,
		Files:      files,
		Size:       size,
	}, nil
}
