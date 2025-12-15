package auth

import (
	"context"
	"fmt"
	"net/url"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// OidcLoginCommand handles the OIDC login use case
// Follows DDD principles by encapsulating the complete OIDC authentication flow
type OidcLoginCommand struct {
	authFactory    *service.AuthFactory
	sessionService *service.SessionService
	config         *infraConfig.InfrastructureConfig
	oidcService    *service.OIDCService
	// Temporary in-memory storage for OIDC sessions
	// In production, this should be replaced with Redis or database storage
	oidcSessions   map[string]*service.OIDCSession
}

// OidcLoginRequest represents the input for OIDC login
type OidcLoginRequest struct {
	// RedirectURL is the URL to redirect to after successful authentication
	RedirectURL string
	// State parameter to prevent CSRF attacks
	State string
}

// OidcLoginResponse represents the output of OIDC login
type OidcLoginResponse struct {
	// AuthURL is the URL to redirect the user to for authentication
	AuthURL string `json:"auth_url"`
	// State parameter returned for CSRF protection
	State string `json:"state"`
}

// NewOidcLoginCommand creates a new OIDC login command
func NewOidcLoginCommand(
	authFactory *service.AuthFactory,
	sessionService *service.SessionService,
	config *infraConfig.InfrastructureConfig,
	oidcService *service.OIDCService,
) *OidcLoginCommand {
	return &OidcLoginCommand{
		authFactory:    authFactory,
		sessionService: sessionService,
		config:         config,
		oidcService:    oidcService,
		oidcSessions:   make(map[string]*service.OIDCSession),
	}
}

// Execute executes the OIDC login command
// Implements the complete OIDC authentication initiation flow
func (c *OidcLoginCommand) Execute(ctx context.Context, req *OidcLoginRequest) (*OidcLoginResponse, error) {
	// Validate request
	if err := c.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if OIDC authentication is configured
	if !c.IsConfigured() {
		return nil, fmt.Errorf("OIDC authentication is not configured")
	}

	// Use the real OIDC service if available
	if c.oidcService != nil {
		authURL, oidcSession, err := c.oidcService.GetAuthURL(ctx, req.State, req.RedirectURL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate OIDC authorization URL: %w", err)
		}

		// Store OIDC session for state validation in callback
		// In production, this should be stored in Redis or database
		c.oidcSessions[req.State] = oidcSession

		return &OidcLoginResponse{
			AuthURL: authURL,
			State:   req.State,
		}, nil
	}

	// OIDC service must be available for proper OIDC authentication
	return nil, fmt.Errorf("OIDC service is not properly configured - please check OIDC configuration settings")
}

// ValidateRequest validates the OIDC login request before execution
func (c *OidcLoginCommand) ValidateRequest(req *OidcLoginRequest) error {
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

// GetOIDCSession retrieves stored OIDC session by state
func (c *OidcLoginCommand) GetOIDCSession(state string) (*service.OIDCSession, bool) {
	session, exists := c.oidcSessions[state]
	return session, exists
}

// RemoveOIDCSession removes a stored OIDC session
func (c *OidcLoginCommand) RemoveOIDCSession(state string) {
	delete(c.oidcSessions, state)
}

// IsConfigured checks if OIDC authentication is properly configured
func (c *OidcLoginCommand) IsConfigured() bool {
	return c.config != nil &&
		c.config.OpenIDConnectIssuer != "" &&
		c.config.OpenIDConnectClientID != "" &&
		c.config.OpenIDConnectClientSecret != "" &&
		c.authFactory != nil &&
		c.sessionService != nil &&
		c.oidcService != nil
}
