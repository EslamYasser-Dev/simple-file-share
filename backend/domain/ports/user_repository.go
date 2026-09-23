package ports

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// UserRepository stores and retrieves user accounts.
type UserRepository interface {
	// CreateUser persists a new user. Returns domainerrors.ErrUserAlreadyExists
	// when a user with the same username already exists.
	CreateUser(user *models.User) error
	// FindByUsername looks up a user, returning domainerrors.ErrUserNotFound
	// when no account matches.
	FindByUsername(username string) (*models.User, error)
	// FindByOAuth looks up a user bound to a provider subject, returning
	// domainerrors.ErrUserNotFound when no account matches.
	FindByOAuth(provider, subject string) (*models.User, error)
	// ListUsers returns all accounts ordered by username.
	ListUsers() ([]*models.User, error)
	// CountUsers returns the number of stored accounts.
	CountUsers() (int, error)
	// GetQuotaBytes returns an account's storage quota in bytes (0 = unlimited),
	// returning domainerrors.ErrUserNotFound when no account matches.
	GetQuotaBytes(username string) (int64, error)
	// SetQuotaBytes updates an account's storage quota in bytes. 0 means
	// unlimited.
	SetQuotaBytes(username string, quotaBytes int64) error
}
