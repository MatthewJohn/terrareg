package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// GithubOAuthCommand handles the GitHub OAuth authentication use cases
// Follows DDD principles by encapsulating the complete GitHub OAuth flow
type GithubOAuthCommand struct {
	authFactory    *service.AuthFactory
	sessionService *service.SessionService
	config         *infraConfig.InfrastructureConfig
}

// GithubOAuthLoginRequest represents the input for GitHub OAuth login
type GithubOAuthLoginRequest struct {
	// RedirectURL is the URL to redirect to after successful authentication
	RedirectURL string
	// State parameter to prevent CSRF attacks
	State string
	// Optional scopes to request (defaults to "user:email")
	Scopes []string
}

// GithubOAuthLoginResponse represents the output of GitHub OAuth login initiation
type GithubOAuthLoginResponse struct {
	// AuthURL is the GitHub OAuth authorization URL
	AuthURL string `json:"auth_url"`
	// State parameter returned for CSRF protection
	State string `json:"state"`
}

// GithubOAuthCallbackRequest represents the input for GitHub OAuth callback
type GithubOAuthCallbackRequest struct {
	// Authorization code returned by GitHub
	Code string
	// State parameter returned by GitHub
	State string
	// Optional error parameter if the user denied access
	Error string
	// Optional error description
	ErrorDescription string
}

// GithubOAuthCallbackResponse represents the output of GitHub OAuth callback
type GithubOAuthCallbackResponse struct {
	// Authenticated indicates whether the authentication was successful
	Authenticated bool `json:"authenticated"`
	// SessionID is the ID of the created session
	SessionID string `json:"session_id,omitempty"`
	// Expiry is the session expiry time
	Expiry time.Time `json:"expiry,omitempty"`
	// Username of the authenticated user
	Username string `json:"username,omitempty"`
	// Email of the authenticated user
	Email string `json:"email,omitempty"`
	// Organizations the user belongs to
	Organizations []string `json:"organizations,omitempty"`
	// ErrorMessage contains any error details
	ErrorMessage string `json:"error_message,omitempty"`
}

// NewGithubOAuthCommand creates a new GitHub OAuth command
func NewGithubOAuthCommand(
	authFactory *service.AuthFactory,
	sessionService *service.SessionService,
	config *infraConfig.InfrastructureConfig,
) *GithubOAuthCommand {
	return &GithubOAuthCommand{
		authFactory:    authFactory,
		sessionService: sessionService,
		config:         config,
	}
}

// ExecuteLogin executes the GitHub OAuth login initiation
func (c *GithubOAuthCommand) ExecuteLogin(ctx context.Context, req *GithubOAuthLoginRequest) (*GithubOAuthLoginResponse, error) {
	// Validate request
	if err := c.ValidateLoginRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if GitHub OAuth is configured
	if !c.IsConfigured() {
		return nil, fmt.Errorf("GitHub OAuth authentication is not configured")
	}

	// For now, return a placeholder response
	// In a full implementation, this would:
	// 1. Generate the GitHub OAuth authorization URL
	// 2. Store state parameter in session/cookie
	// 3. Return the authorization URL to redirect the user

	// Default scopes if none provided
	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = []string{"user:email"}
	}

	// Build the authorization URL
	// Note: GitHub client ID would be retrieved from config when available
	githubClientID := "placeholder-github-client-id" // Would be c.config.GithubClientID
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		"https://github.com/login/oauth/authorize",
		url.QueryEscape(githubClientID),
		url.QueryEscape(req.RedirectURL),
		url.QueryEscape(fmt.Sprintf("%v", scopes)),
		url.QueryEscape(req.State),
	)

	return &GithubOAuthLoginResponse{
		AuthURL: authURL,
		State:   req.State,
	}, nil
}

// ExecuteCallback executes the GitHub OAuth callback
func (c *GithubOAuthCommand) ExecuteCallback(ctx context.Context, req *GithubOAuthCallbackRequest) (*GithubOAuthCallbackResponse, error) {
	// Validate request
	if err := c.ValidateCallbackRequest(req); err != nil {
		return &GithubOAuthCallbackResponse{
			Authenticated: false,
			ErrorMessage:  fmt.Sprintf("Invalid request: %v", err),
		}, nil
	}

	// Check if GitHub OAuth is configured
	if !c.IsConfigured() {
		return &GithubOAuthCallbackResponse{
			Authenticated: false,
			ErrorMessage:  "GitHub OAuth authentication is not configured",
		}, nil
	}

	// Check for authentication error
	if req.Error != "" {
		errorMsg := req.Error
		if req.ErrorDescription != "" {
			errorMsg = fmt.Sprintf("%s: %s", req.Error, req.ErrorDescription)
		}
		return &GithubOAuthCallbackResponse{
			Authenticated: false,
			ErrorMessage:  errorMsg,
		}, nil
	}

	// For now, return a placeholder response
	// In a full implementation, this would:
	// 1. Validate the state parameter
	// 2. Exchange the authorization code for an access token
	// 3. Use the token to fetch user information
	// 4. Fetch user organizations if needed for authorization
	// 5. Create a session for the user
	// 6. Return the session details

	// Placeholder user data
	username := "github-user"
	email := "user@example.com"
	organizations := []string{"example-org", "terraform-modules"}

	// Create session with a default TTL
	ttl := 24 * time.Hour
	authMethod := "GITHUB_OAUTH"

	providerData := map[string]interface{}{
		"client_id":     "placeholder-github-client-id", // Would be c.config.GithubClientID
		"username":      username,
		"email":         email,
		"organizations": organizations,
	}
	providerDataBytes, _ := json.Marshal(providerData)

	session, err := c.sessionService.CreateSession(ctx, authMethod, providerDataBytes, &ttl)
	if err != nil {
		return &GithubOAuthCallbackResponse{
			Authenticated: false,
			ErrorMessage:  fmt.Sprintf("Failed to create session: %v", err),
		}, nil
	}

	return &GithubOAuthCallbackResponse{
		Authenticated: true,
		SessionID:     session.ID,
		Expiry:        session.Expiry,
		Username:      username,
		Email:         email,
		Organizations: organizations,
	}, nil
}

// ValidateLoginRequest validates the GitHub OAuth login request
func (c *GithubOAuthCommand) ValidateLoginRequest(req *GithubOAuthLoginRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.RedirectURL == "" {
		return fmt.Errorf("redirect URL cannot be empty")
	}

	if req.State == "" {
		return fmt.Errorf("state parameter cannot be empty")
	}

	// Validate redirect URL format
	if _, err := url.Parse(req.RedirectURL); err != nil {
		return fmt.Errorf("invalid redirect URL format: %w", err)
	}

	return nil
}

// ValidateCallbackRequest validates the GitHub OAuth callback request
func (c *GithubOAuthCommand) ValidateCallbackRequest(req *GithubOAuthCallbackRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// If there's an error, we don't need to validate other fields
	if req.Error != "" {
		return nil
	}

	if req.Code == "" {
		return fmt.Errorf("authorization code cannot be empty")
	}

	if req.State == "" {
		return fmt.Errorf("state parameter cannot be empty")
	}

	return nil
}

// IsConfigured checks if GitHub OAuth is properly configured
func (c *GithubOAuthCommand) IsConfigured() bool {
	// For now, return true if basic services are available
	// In a full implementation, this would check for GitHub client credentials:
	// c.config.GithubClientID != "" && c.config.GithubClientSecret != ""
	return c.config != nil &&
		c.authFactory != nil &&
		c.sessionService != nil
}
