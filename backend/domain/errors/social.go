package errors

import "errors"

var (
	// ErrFollowNotFound is returned when no follow edge exists.
	ErrFollowNotFound = errors.New("follow not found")
	// ErrAlreadyFollowing is returned when the follow edge already exists.
	ErrAlreadyFollowing = errors.New("already following")
	// ErrCannotFollowSelf is returned when follower and followee match.
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
)
