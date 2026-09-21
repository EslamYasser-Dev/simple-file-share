package errors

import "errors"

// Sentinel errors returned by user-related ports. Callers map them to HTTP
// status codes via errors.Is.
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
