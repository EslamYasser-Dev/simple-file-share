package dto

import (
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

// FollowRequest is the body of POST/DELETE /api/follows.
type FollowRequest struct {
	Username string `json:"username"`
}

// FollowResponse reports the follow edge state after the call.
type FollowResponse struct {
	Username  string `json:"username"`
	Following bool   `json:"following"`
}

// UsernamesResponse lists follower/following usernames, ordered by username.
type UsernamesResponse struct {
	Usernames []string `json:"usernames"`
}

// SetVisibilityRequest is the body of PUT /api/visibility.
type SetVisibilityRequest struct {
	Path string `json:"path"`
	// Level is one of "private", "link", "public".
	Level string `json:"level"`
	// AllowStream gates media playback for non-owners.
	AllowStream bool `json:"allowStream"`
}

// VisibilityItem is a per-file privacy setting as seen by the API.
type VisibilityItem struct {
	Owner       string `json:"owner"`
	Path        string `json:"path"`
	Level       string `json:"level"`
	AllowStream bool   `json:"allowStream"`
	UpdatedAt   string `json:"updatedAt"`
}

func FromVisibility(v *models.FileVisibility) VisibilityItem {
	updated := ""
	if !v.UpdatedAt.IsZero() {
		updated = v.UpdatedAt.UTC().Format(time.RFC3339)
	}
	return VisibilityItem{
		Owner:       v.Owner,
		Path:        v.Path,
		Level:       string(v.Level),
		AllowStream: v.AllowStream,
		UpdatedAt:   updated,
	}
}

// FeedItem is one timeline entry as seen by the API: metadata only, never
// raw content.
type FeedItem struct {
	ID         string `json:"id"`
	Owner      string `json:"owner"`
	Kind       string `json:"kind"`
	Path       string `json:"path"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Visibility string `json:"visibility"`
	CreatedAt  string `json:"createdAt"`
}

func FromTimelineEvent(e *models.TimelineEvent) FeedItem {
	created := ""
	if !e.CreatedAt.IsZero() {
		created = e.CreatedAt.UTC().Format(time.RFC3339)
	}
	return FeedItem{
		ID:         e.ID,
		Owner:      e.Owner,
		Kind:       e.Kind,
		Path:       e.Path,
		Name:       e.Name,
		Size:       e.Size,
		Visibility: string(e.Visibility),
		CreatedAt:  created,
	}
}

func FromTimelineEvents(events []*models.TimelineEvent) []FeedItem {
	items := make([]FeedItem, 0, len(events))
	for _, e := range events {
		items = append(items, FromTimelineEvent(e))
	}
	return items
}

// FeedResponse is the GET /api/feed envelope. NextCursor is the last event
// ID; empty means the feed is exhausted.
type FeedResponse struct {
	Events     []FeedItem `json:"events"`
	NextCursor string     `json:"nextCursor"`
}
