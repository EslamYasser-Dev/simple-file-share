package dto

import (
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

type UserResponse struct {
	Username  string `json:"username"`
	IsAdmin   bool   `json:"isAdmin"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type MeResponse struct {
	Username   string `json:"username,omitempty"`
	IsAdmin    bool   `json:"isAdmin"`
	QuotaBytes int64  `json:"quotaBytes"`
	Size       int64  `json:"size"`
	Files      int    `json:"files"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

type UserStatsResponse struct {
	Username   string `json:"username"`
	IsAdmin    bool   `json:"isAdmin"`
	QuotaBytes int64  `json:"quotaBytes,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
	Files      int    `json:"files"`
	Size       int64  `json:"size"`
}

type AuthInfoResponse struct {
	SignupEnabled bool `json:"signupEnabled"`
}

func FromUser(u *models.User) UserResponse {
	if u == nil {
		return UserResponse{}
	}
	return UserResponse{
		Username:  strings.TrimSpace(u.Username),
		IsAdmin:   u.IsAdmin,
		CreatedAt: formatTime(u.CreatedAt),
	}
}

func FromUserStats(stats []models.UserStats) []UserStatsResponse {
	out := make([]UserStatsResponse, 0, len(stats))
	for _, s := range stats {
		out = append(out, UserStatsResponse{
			Username:   s.Username,
			IsAdmin:    s.IsAdmin,
			QuotaBytes: s.QuotaBytes,
			CreatedAt:  formatTime(s.CreatedAt),
			Files:      s.Files,
			Size:       s.Size,
		})
	}
	return out
}

func FromUserStatsSingle(s models.UserStats) MeResponse {
	return MeResponse{
		Username:   s.Username,
		IsAdmin:    s.IsAdmin,
		QuotaBytes: s.QuotaBytes,
		Size:       s.Size,
		Files:      s.Files,
		CreatedAt:  formatTime(s.CreatedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
