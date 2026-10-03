package services

import (
	"strings"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// FollowService manages open-follow edges: one tap to follow, no approval.
// Both endpoints must be real accounts; following yourself is rejected.
type FollowService struct {
	follows ports.FollowRepository
	users   ports.UserRepository
}

func NewFollowService(follows ports.FollowRepository, users ports.UserRepository) *FollowService {
	return &FollowService{follows: follows, users: users}
}

// Follow records follower -> followee.
func (s *FollowService) Follow(follower, followee string) error {
	follower = strings.TrimSpace(follower)
	followee = strings.TrimSpace(followee)
	if follower == "" || followee == "" {
		return domainerrors.NewValidationError("followee", followee, "must be a non-empty username")
	}
	if follower == followee {
		return domainerrors.ErrCannotFollowSelf
	}
	if _, err := s.users.FindByUsername(follower); err != nil {
		return err
	}
	if _, err := s.users.FindByUsername(followee); err != nil {
		return err
	}
	return s.follows.Follow(follower, followee)
}

// Unfollow removes follower -> followee.
func (s *FollowService) Unfollow(follower, followee string) error {
	follower = strings.TrimSpace(follower)
	followee = strings.TrimSpace(followee)
	if follower == "" || followee == "" {
		return domainerrors.NewValidationError("followee", followee, "must be a non-empty username")
	}
	return s.follows.Unfollow(follower, followee)
}

// IsFollowing reports whether follower -> followee exists.
func (s *FollowService) IsFollowing(follower, followee string) (bool, error) {
	return s.follows.IsFollowing(follower, followee)
}

// Followers returns usernames following followee, ordered by username.
func (s *FollowService) Followers(followee string) ([]string, error) {
	if _, err := s.users.FindByUsername(followee); err != nil {
		return nil, err
	}
	return s.follows.Followers(followee)
}

// Following returns usernames follower follows, ordered by username.
func (s *FollowService) Following(follower string) ([]string, error) {
	if _, err := s.users.FindByUsername(follower); err != nil {
		return nil, err
	}
	return s.follows.Following(follower)
}
