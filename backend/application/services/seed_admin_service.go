package services

import (
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// SeedAdminService bootstraps the first admin account so existing deployments
// (which previously used static USERNAME/PASSWORD credentials) keep working.
// It is a no-op once any user exists.
type SeedAdminService struct {
	users    ports.UserRepository
	hasher   ports.PasswordHasher
	fileRepo ports.FileRepository
	scoper   ports.PathScoper
}

func NewSeedAdminService(
	users ports.UserRepository,
	hasher ports.PasswordHasher,
	fileRepo ports.FileRepository,
	scoper ports.PathScoper,
) *SeedAdminService {
	return &SeedAdminService{
		users:    users,
		hasher:   hasher,
		fileRepo: fileRepo,
		scoper:   scoper,
	}
}

// Execute creates the bootstrap admin when the store is empty. It reports
// whether an account was actually created.
func (s *SeedAdminService) Execute(username, password string) (bool, error) {
	username = strings.TrimSpace(username)
	if username == "" || len(password) < minPasswordLength {
		return false, nil
	}

	count, err := s.users.CountUsers()
	if err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return false, err
	}

	if err := s.users.CreateUser(&models.User{
		Username:     username,
		PasswordHash: hash,
		Role:         models.RoleAdmin,
		IsAdmin:      true,
		Enabled:      true,
		CreatedAt:    time.Now().UTC(),
	}); err != nil {
		return false, err
	}

	_ = s.fileRepo.CreateDirectory(s.scoper.PrivatePrefix(username))
	return true, nil
}
