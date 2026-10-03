package models

import "time"

// Follow is a one-way subscription: Follower sees Followee's link/public
// uploads in their timeline feed. Open-follow: no approval step. Blocks are
// out of scope; unfollowing is the remedy.
type Follow struct {
	Follower  string
	Followee  string
	CreatedAt time.Time
}
