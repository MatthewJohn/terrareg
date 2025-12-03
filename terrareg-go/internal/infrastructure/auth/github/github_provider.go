package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"terrareg/internal/domain/identity/model"
	"terrareg/internal/domain/identity/service"
)

var (
	ErrGitHubNotConfigured    = errors.New("GitHub provider not configured")
	ErrGitHubInvalidToken    = errors.New("invalid GitHub token")
	ErrGitHubAuthentication = errors.New("GitHub authentication failed")
)

// GitHubProvider implements GitHub OAuth authentication
type GitHubProvider struct {
	client     *github.Client
	oauth2Config *oauth2.Config
	config     GitHubConfig
}

// GitHubConfig holds GitHub configuration
type GitHubConfig struct {
	ClientID         string
	ClientSecret     string
	RedirectURL      string
	Scopes           []string
	AuthURL          string
	TokenURL         string
	UserInfoURL      string
	SessionTimeout    time.Duration
	Organization     string // Optional: restrict to organization
	Teams            []string // Optional: restrict to specific teams
}

// GitHubUser represents GitHub user information
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL  string `json:"avatar_url"`
	Company   string `json:"company"`
	Blog      string `json:"blog"`
	Location  string `json:"location"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GitHubTokenResponse represents the OAuth token response from GitHub
type GitHubTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope       string `json:"scope"`
}

// NewGitHubProvider creates a new GitHub provider
func NewGitHubProvider(config GitHubConfig) (*GitHubProvider, error) {
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, ErrGitHubNotConfigured
	}

	// Default GitHub OAuth endpoints
	if config.AuthURL == "" {
		config.AuthURL = "https://github.com/login/oauth/authorize"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://github.com/login/oauth/access_token"
	}

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.AuthURL,
			TokenURL: config.TokenURL,
		},
		Scopes:      config.Scopes,
		RedirectURL: config.RedirectURL,
	}

	return &GitHubProvider{
		oauth2Config: oauth2Config,
		config:       config,
	}, nil
}

// GetAuthURL returns the GitHub authentication URL
func (p *GitHubProvider) GetAuthURL(ctx context.Context, state string) (string, error) {
	// Generate auth code URL with state
	authURL := p.oauth2Config.AuthCodeURL(state)
	return authURL, nil
}

// Authenticate handles GitHub authentication response
func (p *GitHubProvider) Authenticate(ctx context.Context, request *http.Request) (*service.AuthResult, error) {
	// Extract authorization code from request
	code := request.FormValue("code")
	if code == "" {
		return nil, ErrGitHubAuthentication
	}

	// Exchange authorization code for tokens
	token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Get user information from GitHub API
	userInfo, err := p.getUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub user info: %w", err)
	}

	// Check organization restrictions if configured
	if p.config.Organization != "" {
		if !p.checkOrganizationMembership(ctx, token.AccessToken, p.config.Organization) {
			return nil, fmt.Errorf("user is not a member of required organization: %s", p.config.Organization)
		}
	}

	// Calculate expires in (GitHub tokens don't expire unless revoked)
	expiresIn := int64(p.config.SessionTimeout.Seconds())

	// Create auth result
	return &service.AuthResult{
		UserID:         strconv.FormatInt(userInfo.ID, 10),
		Username:       userInfo.Login,
		Email:          userInfo.Email,
		DisplayName:    userInfo.Name,
		ExternalID:     strconv.FormatInt(userInfo.ID, 10),
		AuthProviderID: "github",
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken, // GitHub doesn't typically provide refresh tokens
		ExpiresIn:      expiresIn,
	}, nil
}

// RefreshToken refreshes an access token using refresh token
func (p *GitHubProvider) RefreshToken(ctx context.Context, refreshToken string) (*service.AuthResult, error) {
	// GitHub doesn't typically support refresh tokens for OAuth apps
	// Users need to re-authenticate
	return nil, errors.New("GitHub OAuth doesn't support token refresh, please re-authenticate")
}

// GetUserInfo returns user information from GitHub API
func (p *GitHubProvider) GetUserInfo(ctx context.Context, accessToken string) (*service.UserInfo, error) {
	userInfo, err := p.getUserInfo(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	return &service.UserInfo{
		ID:          strconv.FormatInt(userInfo.ID, 10),
		Username:    userInfo.Login,
		Email:       userInfo.Email,
		DisplayName: userInfo.Name,
		AvatarURL:   userInfo.AvatarURL,
		Groups:      p.getUserTeams(ctx, accessToken), // Add GitHub orgs/teams as groups
	}, nil
}

// getUserInfo fetches user information from GitHub API
func (p *GitHubProvider) getUserInfo(ctx context.Context, accessToken string) (*GitHubUser, error) {
	client := p.createClient(accessToken)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub user: %w", err)
	}

	// Get user email (GitHub requires separate API call for primary email)
	email, _, err := client.Users.GetPrimaryEmail(ctx)
	if err == nil && email != nil {
		user.Email = email.GetEmail()
	}

	return &GitHubUser{
		ID:        user.GetID(),
		Login:     user.GetLogin(),
		Name:      user.GetName(),
		Email:     user.GetEmail(),
		AvatarURL:  user.GetAvatarURL(),
		Company:   user.GetCompany(),
		Blog:      user.GetBlog(),
		Location:  user.GetLocation(),
		Bio:       user.GetBio(),
		CreatedAt: user.GetCreatedAt().Format(time.RFC3339),
		UpdatedAt: user.GetUpdatedAt().Format(time.RFC3339),
	}, nil
}

// checkOrganizationMembership checks if user is a member of the specified organization
func (p *GitHubProvider) checkOrganizationMembership(ctx context.Context, accessToken, org string) bool {
	client := p.createClient(accessToken)

	membership, _, err := client.Organizations.GetOrgMembership(ctx, org)
	if err != nil {
		return false
	}

	return membership != nil
}

// getUserTeams gets user's teams (used for group membership)
func (p *GitHubProvider) getUserTeams(ctx context.Context, accessToken string) []string {
	client := p.createClient(accessToken)

	teams, _, err := client.Teams.ListUserTeams(ctx, nil)
	if err != nil {
		return []string{}
	}

	var teamNames []string
	for _, team := range teams {
		teamNames = append(teamNames, team.GetSlug())
	}

	return teamNames
}

// createClient creates a GitHub client with the access token
func (p *GitHubProvider) createClient(accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}