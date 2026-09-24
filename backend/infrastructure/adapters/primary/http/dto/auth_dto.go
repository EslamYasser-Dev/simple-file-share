package dto

import (
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

type UserResponse struct {
	Username  string `json:"username"`
	Role      string `json:"role,omitempty"`
	IsAdmin   bool   `json:"isAdmin"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type MeResponse struct {
	Username    string   `json:"username,omitempty"`
	Role        string   `json:"role,omitempty"`
	IsAdmin     bool     `json:"isAdmin"`
	Enabled     bool     `json:"enabled"`
	Permissions []string `json:"permissions,omitempty"`
	QuotaBytes  int64    `json:"quotaBytes"`
	Size        int64    `json:"size"`
	Files       int      `json:"files"`
	CreatedAt   string   `json:"createdAt,omitempty"`
}

type UserStatsResponse struct {
	Username   string `json:"username"`
	Role       string `json:"role,omitempty"`
	IsAdmin    bool   `json:"isAdmin"`
	Enabled    bool   `json:"enabled"`
	QuotaBytes int64  `json:"quotaBytes,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
	Files      int    `json:"files"`
	Size       int64  `json:"size"`
}

type AuthInfoResponse struct {
	SignupEnabled bool     `json:"signupEnabled"`
	OAuth         []string `json:"oauth,omitempty"`
}

// TokenResponse is the wire shape of a freshly issued access token.
type TokenResponse struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int64  `json:"expiresIn"`
}

func FromTokenPair(pair *services.TokenPair) TokenResponse {
	if pair == nil {
		return TokenResponse{}
	}
	expiresIn := int64(0)
	if !pair.ExpiresAt.IsZero() {
		expiresIn = int64(time.Until(pair.ExpiresAt).Seconds())
		if expiresIn < 0 {
			expiresIn = 0
		}
	}
	return TokenResponse{
		AccessToken: pair.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}
}

func FromUser(u *models.User) UserResponse {
	if u == nil {
		return UserResponse{}
	}
	return UserResponse{
		Username:  strings.TrimSpace(u.Username),
		Role:      u.Role,
		IsAdmin:   u.IsAdmin,
		Enabled:   u.Enabled,
		CreatedAt: formatTime(u.CreatedAt),
	}
}

func FromUserStats(stats []models.UserStats) []UserStatsResponse {
	out := make([]UserStatsResponse, 0, len(stats))
	for _, s := range stats {
		out = append(out, UserStatsResponse{
			Username:   s.Username,
			Role:       s.Role,
			IsAdmin:    s.IsAdmin,
			Enabled:    s.Enabled,
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
		Username:    s.Username,
		Role:        s.Role,
		IsAdmin:     s.IsAdmin,
		Enabled:     s.Enabled,
		Permissions: models.AllPermissionsIfAdmin(s.IsAdmin),
		QuotaBytes:  s.QuotaBytes,
		Size:        s.Size,
		Files:       s.Files,
		CreatedAt:   formatTime(s.CreatedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
