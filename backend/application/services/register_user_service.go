package services

import (
	"regexp"
	"strings"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{2,31}$`)

const minPasswordLength = 4

// RegisterUserService creates new accounts. The first account ever created
// becomes an admin.
type RegisterUserService struct {
	users             ports.UserRepository
	hasher            ports.PasswordHasher
	fileRepo          ports.FileRepository
	scoper            ports.PathScoper
	signupEnabled     bool
	defaultQuotaBytes int64
}

func NewRegisterUserService(
	users ports.UserRepository,
	hasher ports.PasswordHasher,
	fileRepo ports.FileRepository,
	scoper ports.PathScoper,
	signupEnabled bool,
	defaultQuotaBytes int64,
) *RegisterUserService {
	return &RegisterUserService{
		users:             users,
		hasher:            hasher,
		fileRepo:          fileRepo,
		scoper:            scoper,
		signupEnabled:     signupEnabled,
		defaultQuotaBytes: defaultQuotaBytes,
	}
}

// Execute validates the credentials, hashes the password, and creates the
// account plus its private home directory. Returns the created user with the
// password hash redacted.
func (s *RegisterUserService) Execute(username, password string) (*models.User, error) {
	username = strings.TrimSpace(username)
	if !s.signupEnabled {
		return nil, &domainerrors.ForbiddenError{Action: "register"}
	}
	if !usernamePattern.MatchString(username) {
		return nil, domainerrors.NewValidationError("username", username, "username must be 3-32 characters (letters, digits, dots, dashes, underscores)")
	}
	if len(password) < minPasswordLength {
		return nil, domainerrors.NewValidationError("password", "", "password must be at least 4 characters")
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		PasswordHash: hash,
		Role:         models.RoleMember,
		IsAdmin:      false,
		Enabled:      true,
		QuotaBytes:   s.defaultQuotaBytes,
		CreatedAt:    time.Now().UTC(),
	}

	count, err := s.users.CountUsers()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		user.Role = models.RoleAdmin
		user.IsAdmin = true
	}

	if err := s.users.CreateUser(user); err != nil {
		return nil, err
	}

	// Pre-create the user's private home so listing `/` works immediately.
	_ = s.fileRepo.CreateDirectory(s.scoper.PrivatePrefix(user.Username))

	user.PasswordHash = ""
	return user, nil
}
