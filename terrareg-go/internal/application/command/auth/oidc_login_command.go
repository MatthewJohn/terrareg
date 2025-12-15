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
) *OidcLoginCommand {
	// Initialize OIDC service
	var oidcService *service.OIDCService
	if config != nil && config.OpenIDConnectIssuer != "" {
		oidcService, _ = service.NewOIDCService(context.Background(), config)
	}

	return &OidcLoginCommand{
		authFactory:    authFactory,
		sessionService: sessionService,
		config:         config,
		oidcService:    oidcService,
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
		authURL, _, err := c.oidcService.GetAuthURL(ctx, req.State, req.RedirectURL)
		if err != nil {
			return nil, fmt.Errorf("failed to generate OIDC authorization URL: %w", err)
		}

		// TODO: Store OIDC session in secure storage (Redis/database)
		// For now, we'll rely on the state parameter validation in the callback

		return &OidcLoginResponse{
			AuthURL: authURL,
			State:   req.State,
		}, nil
	}

	// Fallback to basic URL construction if OIDC service is not available
	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		c.config.OpenIDConnectIssuer,
		url.QueryEscape(c.config.OpenIDConnectClientID),
		url.QueryEscape(req.RedirectURL),
		url.QueryEscape("openid profile email"),
		url.QueryEscape(req.State),
	)

	return &OidcLoginResponse{
		AuthURL: authURL,
		State:   req.State,
	}, nil
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

// IsConfigured checks if OIDC authentication is properly configured
func (c *OidcLoginCommand) IsConfigured() bool {
	return c.config != nil &&
		c.config.OpenIDConnectIssuer != "" &&
		c.config.OpenIDConnectClientID != "" &&
		c.config.OpenIDConnectClientSecret != "" &&
		c.authFactory != nil &&
		c.sessionService != nil
}
