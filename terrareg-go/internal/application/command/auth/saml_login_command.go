package auth

import (
	"context"
	"fmt"
	"net/url"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// SamlLoginCommand handles the SAML login use case
// Follows DDD principles by encapsulating the complete SAML authentication flow
type SamlLoginCommand struct {
	authFactory    *service.AuthFactory
	sessionService *service.SessionService
	config         *infraConfig.InfrastructureConfig
	samlService    *service.SAMLService
}

// SamlLoginRequest represents the input for SAML login
type SamlLoginRequest struct {
	// RelayState is the URL to redirect to after successful authentication
	RelayState string
	// Optional IDP identifier for multi-IDP configurations
	IDP string
}

// SamlLoginResponse represents the output of SAML login
type SamlLoginResponse struct {
	// AuthURL is the SAML request URL to redirect the user to
	AuthURL string `json:"auth_url"`
	// SAMLRequest is the base64 encoded SAML request
	SAMLRequest string `json:"saml_request"`
	// RelayState is the state parameter to maintain session state
	RelayState string `json:"relay_state"`
}

// NewSamlLoginCommand creates a new SAML login command
func NewSamlLoginCommand(
	authFactory *service.AuthFactory,
	sessionService *service.SessionService,
	config *infraConfig.InfrastructureConfig,
	samlService *service.SAMLService,
) *SamlLoginCommand {
	return &SamlLoginCommand{
		authFactory:    authFactory,
		sessionService: sessionService,
		config:         config,
		samlService:    samlService,
	}
}

// Execute executes the SAML login command
// Implements the complete SAML authentication initiation flow
func (c *SamlLoginCommand) Execute(ctx context.Context, req *SamlLoginRequest) (*SamlLoginResponse, error) {
	// Validate request
	if err := c.ValidateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if SAML authentication is configured
	if !c.IsConfigured() {
		return nil, fmt.Errorf("SAML authentication is not configured")
	}

	// Use the real SAML service if available
	if c.samlService != nil {
		// Create SAML authentication request
		samlAuthReq, err := c.samlService.CreateAuthRequest(ctx, req.RelayState)
		if err != nil {
			return nil, fmt.Errorf("failed to create SAML authentication request: %w", err)
		}

		return &SamlLoginResponse{
			AuthURL:     samlAuthReq.AuthURL,
			SAMLRequest: samlAuthReq.SAMLRequest,
			RelayState:  samlAuthReq.RelayState,
		}, nil
	}

	// SAML service must be available for proper SAML authentication
	return nil, fmt.Errorf("SAML service is not properly configured - please check SAML configuration settings")
}

// ValidateRequest validates the SAML login request before execution
func (c *SamlLoginCommand) ValidateRequest(req *SamlLoginRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// RelayState is optional, but if provided, validate it's a proper URL
	if req.RelayState != "" {
		if _, err := url.Parse(req.RelayState); err != nil {
			return fmt.Errorf("invalid relay state URL format: %w", err)
		}
	}

	return nil
}

// IsConfigured checks if SAML authentication is properly configured
func (c *SamlLoginCommand) IsConfigured() bool {
	return c.config != nil &&
		c.config.SAML2EntityID != "" &&
		c.config.SAML2IDPMetadataURL != "" &&
		c.authFactory != nil &&
		c.sessionService != nil &&
		c.samlService != nil
}
