package errors

import "errors"

// Sentinel errors returned by user-related ports. Callers map them to HTTP
// status codes via errors.Is.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrTwoFactorRequired is returned when the account has TOTP enabled but
	// the login request carried no (or an invalid-channel) second factor.
	// Handlers surface it as 401 {"error":"totp_required"} so clients can
	// prompt for a code without treating it as a failed password attempt.
	ErrTwoFactorRequired = errors.New("totp_required")
)
