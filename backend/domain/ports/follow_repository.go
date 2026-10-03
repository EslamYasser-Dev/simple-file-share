package ports

// FollowRepository stores one-way follow edges. Implementations must be safe
// for concurrent use.
type FollowRepository interface {
	// Follow persists follower -> followee, returning
	// domainerrors.ErrAlreadyFollowing when the edge exists.
	Follow(follower, followee string) error
	// Unfollow removes follower -> followee, returning
	// domainerrors.ErrFollowNotFound when no edge exists.
	Unfollow(follower, followee string) error
	// IsFollowing reports whether follower -> followee exists.
	IsFollowing(follower, followee string) (bool, error)
	// Followers returns usernames following followee, ordered by username.
	Followers(followee string) ([]string, error)
	// Following returns usernames followee follows, ordered by username.
	Following(follower string) ([]string, error)
}
