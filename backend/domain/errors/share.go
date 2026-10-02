package errors

import "errors"

var (
	// ErrShareNotFound is returned by the share repository when no share with
	// the given token exists.
	ErrShareNotFound = errors.New("share not found")
)

// ShareNotFoundError indicates a share token that does not exist or was
// revoked, so the linked content can no longer be served.
type ShareNotFoundError struct {
	Token string
}

func (e *ShareNotFoundError) Error() string {
	return "share link not found"
}

// ShareExpiredError indicates a share whose validity window has passed.
type ShareExpiredError struct {
	Token string
}

func (e *ShareExpiredError) Error() string {
	return "share link has expired"
}

// ErrShareLimitReached indicates a share link whose download budget is used
// up (MaxDownloads consumed).
var ErrShareLimitReached = errors.New("share download limit reached")

// SharePasswordError indicates a password-protected share link presented
// without (Missing) or with the wrong (Missing=false) password.
type SharePasswordError struct {
	Token   string
	Missing bool
}

func (e *SharePasswordError) Error() string {
	if e.Missing {
		return "share password required"
	}
	return "share password incorrect"
}
