package services

import (
	"strings"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	secondaryauth "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
)

// timeNow is a seam so tests can pin the clock when needed.
var timeNow = time.Now

// backupCodeCount is how many single-use recovery codes enrollment issues.
const backupCodeCount = 10

// totpIssuer appears in authenticator provisioning URIs.
const totpIssuer = "Simple File Share"

// TOTPService owns the two-factor lifecycle: enrollment, confirmation,
// disable, admin reset, and second-factor verification at login. State lives
// on models.User and persists through the regular user repository.
type TOTPService struct {
	users  ports.UserRepository
	hasher ports.PasswordHasher
	roles  *RoleCatalog
}

// NewTOTPService wires the user repository, password hasher (reused for
// backup codes) and the role catalog for the admin reset gate.
func NewTOTPService(users ports.UserRepository, hasher ports.PasswordHasher, roles *RoleCatalog) *TOTPService {
	return &TOTPService{users: users, hasher: hasher, roles: roles}
}

// Enroll starts (or restarts) enrollment for the actor and returns the
// fresh secret plus its otpauth:// provisioning URI. The account stays
// single-factor until Confirm succeeds.
func (s *TOTPService) Enroll(actor *models.User) (secret, uri string, err error) {
	if actor == nil {
		return "", "", &domainerrors.ValidationError{Field: "actor", Message: "authentication required"}
	}
	if actor.TOTPEnabled {
		return "", "", &domainerrors.ValidationError{Field: "code", Message: "two-factor is already enabled"}
	}
	secret, err = secondaryauth.GenerateTOTPSecret()
	if err != nil {
		return "", "", err
	}
	actor.TOTPSecret = secret
	actor.TOTPEnabled = false
	if err := s.users.UpdateUser(actor, actor.Username); err != nil {
		return "", "", err
	}
	return secret, secondaryauth.TOTPProvisioningURI(actor.Username, totpIssuer, secret), nil
}

// Confirm verifies the first code, flips the account to fully enrolled, and
// returns the fresh backup codes exactly once (they are stored hashed).
func (s *TOTPService) Confirm(actor *models.User, code string) ([]string, error) {
	if actor == nil {
		return nil, &domainerrors.ValidationError{Field: "actor", Message: "authentication required"}
	}
	if actor.TOTPEnabled {
		return nil, &domainerrors.ValidationError{Field: "code", Message: "two-factor is already enabled"}
	}
	if actor.TOTPSecret == "" {
		return nil, &domainerrors.ValidationError{Field: "code", Message: "no enrollment in progress"}
	}
	if !secondaryauth.ValidateTOTP(actor.TOTPSecret, code, 1, timeNow()) {
		return nil, &domainerrors.ValidationError{Field: "code", Message: "invalid two-factor code"}
	}
	raw, err := secondaryauth.GenerateBackupCodes(backupCodeCount)
	if err != nil {
		return nil, err
	}
	hashed := make([]string, 0, len(raw))
	for _, c := range raw {
		h, err := s.hasher.Hash(c)
		if err != nil {
			return nil, err
		}
		hashed = append(hashed, h)
	}
	actor.TOTPEnabled = true
	actor.BackupCodes = hashed
	if err := s.users.UpdateUser(actor, actor.Username); err != nil {
		return nil, err
	}
	return raw, nil
}

// Disable turns two-factor off. While enrollment is pending (secret set but
// never confirmed) no code is needed; once enabled the current TOTP code or
// an unused backup code must be presented.
func (s *TOTPService) Disable(actor *models.User, code string) error {
	if actor == nil {
		return &domainerrors.ValidationError{Field: "actor", Message: "authentication required"}
	}
	if !actor.TOTPEnabled && actor.TOTPSecret != "" {
		// Pending enrollment — nothing to prove.
		actor.TOTPSecret = ""
		actor.BackupCodes = nil
		return s.users.UpdateUser(actor, actor.Username)
	}
	if !actor.TOTPEnabled {
		return &domainerrors.ValidationError{Field: "code", Message: "two-factor is not enabled"}
	}
	if secondaryauth.ValidateTOTP(actor.TOTPSecret, code, 1, timeNow()) {
		return s.clear(actor)
	}
	if secondaryauth.LooksLikeBackupCode(code) && s.consumeBackupCode(actor, code) {
		return s.clear(actor)
	}
	return &domainerrors.ValidationError{Field: "code", Message: "invalid two-factor code"}
}

// AdminReset clears two-factor state for a target account (device lost,
// codes unavailable). Requires users.update.
func (s *TOTPService) AdminReset(admin *models.User, target string) error {
	if admin == nil || !admin.HasPermission(models.PermUsersUpdate, s.roles) {
		return forbiddenErr("reset two-factor")
	}
	user, err := s.users.FindByUsername(target)
	if err != nil {
		return err
	}
	user.TOTPEnabled = false
	user.TOTPSecret = ""
	user.BackupCodes = nil
	return s.users.UpdateUser(user, user.Username)
}

// VerifyLogin enforces the second factor for password logins. Empty otp
// yields ErrTwoFactorRequired (the challenge); an invalid code yields
// ErrInvalidCredentials so the response stays indistinguishable from a bad
// password. Successful backup-code use consumes the code.
func (s *TOTPService) VerifyLogin(user *models.User, otp string) error {
	if user == nil || !user.TOTPEnabled {
		return nil
	}
	otp = strings.TrimSpace(otp)
	if otp == "" {
		return domainerrors.ErrTwoFactorRequired
	}
	if secondaryauth.ValidateTOTP(user.TOTPSecret, otp, 1, timeNow()) {
		return nil
	}
	if secondaryauth.LooksLikeBackupCode(otp) {
		if s.consumeBackupCode(user, otp) {
			return nil
		}
	}
	return domainerrors.ErrInvalidCredentials
}

func (s *TOTPService) clear(actor *models.User) error {
	actor.TOTPEnabled = false
	actor.TOTPSecret = ""
	actor.BackupCodes = nil
	return s.users.UpdateUser(actor, actor.Username)
}

// consumeBackupCode verifies a backup code and removes it on success
// (single use). Returns whether a code matched.
func (s *TOTPService) consumeBackupCode(user *models.User, code string) bool {
	matched := -1
	for i, hashed := range user.BackupCodes {
		if s.hasher.Verify(code, hashed) {
			matched = i
			break
		}
	}
	if matched < 0 {
		return false
	}
	user.BackupCodes = append(user.BackupCodes[:matched], user.BackupCodes[matched+1:]...)
	if err := s.users.UpdateUser(user, user.Username); err != nil {
		// Persistence failed: the code still works once more, but login
		// proceeds — never lock a user out because of a save hiccup.
		return true
	}
	return true
}
