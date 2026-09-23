package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var ErrOAuthNotConfigured = errors.New("oauth provider not configured")

// OAuthProfile is the normalized identity returned by a provider.
// Subject is the provider-stable user id (never an email or display login).
type OAuthProfile struct {
	Username string
	Email    string
	Subject  string
}

// OAuthProvider builds the authorize URL and exchanges a code for a profile.
type OAuthProvider struct {
	name         string
	clientID     string
	clientSecret string
	authURL      string
	tokenURL     string
	userURL      string
	scope        string
	httpClient   *http.Client
}

func (p *OAuthProvider) Name() string { return p.name }

func (p *OAuthProvider) AuthURL(redirectURI, state string) string {
	q := url.Values{}
	q.Set("client_id", p.clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", p.scope)
	q.Set("state", state)
	return p.authURL + "?" + q.Encode()
}

func (p *OAuthProvider) Exchange(code, redirectURI string) (*OAuthProfile, error) {
	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest(http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	if tokenResp.AccessToken == "" || tokenResp.Error != "" {
		return nil, errors.New("no access token in oauth response")
	}

	return p.fetchProfile(tokenResp.AccessToken)
}

func (p *OAuthProvider) fetchProfile(accessToken string) (*OAuthProfile, error) {
	req, err := http.NewRequest(http.MethodGet, p.userURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch profile: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo returned %d", resp.StatusCode)
	}

	switch p.name {
	case "github":
		var gh struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal(body, &gh); err != nil {
			return nil, err
		}
		if gh.Login == "" || gh.ID == 0 {
			return nil, errors.New("github profile missing login or id")
		}
		return &OAuthProfile{
			Username: gh.Login,
			Email:    gh.Email,
			Subject:  strconv.FormatInt(gh.ID, 10),
		}, nil
	case "google":
		var g struct {
			Sub           string `json:"sub"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err := json.Unmarshal(body, &g); err != nil {
			return nil, err
		}
		if g.Sub == "" {
			return nil, errors.New("google profile missing sub")
		}
		if g.Email == "" || !g.EmailVerified {
			return nil, errors.New("google email missing or not verified")
		}
		username := g.Email
		if at := strings.Index(username, "@"); at > 0 {
			username = username[:at]
		}
		return &OAuthProfile{
			Username: username,
			Email:    g.Email,
			Subject:  g.Sub,
		}, nil
	default:
		return nil, errors.New("unknown oauth provider")
	}
}

// NewOAuthProviders builds the configured providers from the environment.
// Unconfigured providers are omitted so /api/auth/info only advertises live ones.
func NewOAuthProviders() map[string]*OAuthProvider {
	httpClient := &http.Client{Timeout: 15 * time.Second}
	providers := map[string]*OAuthProvider{}

	if id, secret := os.Getenv("OAUTH_GITHUB_CLIENT_ID"), os.Getenv("OAUTH_GITHUB_CLIENT_SECRET"); id != "" && secret != "" {
		providers["github"] = &OAuthProvider{
			name:         "github",
			clientID:     id,
			clientSecret: secret,
			authURL:      "https://github.com/login/oauth/authorize",
			tokenURL:     "https://github.com/login/oauth/access_token",
			userURL:      "https://api.github.com/user",
			scope:        "read:user",
			httpClient:   httpClient,
		}
	}
	if id, secret := os.Getenv("OAUTH_GOOGLE_CLIENT_ID"), os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"); id != "" && secret != "" {
		providers["google"] = &OAuthProvider{
			name:         "google",
			clientID:     id,
			clientSecret: secret,
			authURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			tokenURL:     "https://oauth2.googleapis.com/token",
			userURL:      "https://openidconnect.googleapis.com/v1/userinfo",
			scope:        "openid email profile",
			httpClient:   httpClient,
		}
	}
	return providers
}

// OAuthProviderNames returns the configured provider names in stable order.
func OAuthProviderNames(providers map[string]*OAuthProvider) []string {
	known := []string{"github", "google"}
	out := make([]string, 0, len(providers))
	for _, name := range known {
		if _, ok := providers[name]; ok {
			out = append(out, name)
		}
	}
	return out
}

// ValidateRedirectBase ensures an explicit OAuth public base is an absolute
// http(s) URL with no path/query. Empty is allowed only outside production
// callers (see ValidateProductionOAuthConfig).
func ValidateRedirectBase(base string) error {
	base = strings.TrimSpace(base)
	if base == "" {
		return errors.New("OAUTH_REDIRECT_BASE is empty")
	}
	u, err := url.Parse(base)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errors.New("OAUTH_REDIRECT_BASE must be an absolute http(s) origin")
	}
	if u.Path != "" && u.Path != "/" {
		return errors.New("OAUTH_REDIRECT_BASE must not include a path")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return errors.New("OAUTH_REDIRECT_BASE must not include query or fragment")
	}
	return nil
}

// OAuthCallbackPath is the fixed callback path for a provider.
func OAuthCallbackPath(provider string) string {
	return "/api/auth/oauth/" + provider + "/callback"
}
