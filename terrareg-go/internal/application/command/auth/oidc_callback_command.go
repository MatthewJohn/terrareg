package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// OidcCallbackCommand handles the OIDC authentication callback
// Follows DDD principles by encapsulating the complete OIDC callback flow
type OidcCallbackCommand struct {
	authFactory    *service.AuthFactory
	sessionService *service.SessionService
	config         *infraConfig.InfrastructureConfig
	oidcService    *service.OIDCService
	loginCommand   *OidcLoginCommand // Reference to login command for session access
}

// OidcCallbackRequest represents the input for OIDC callback
type OidcCallbackRequest struct {
	// Authorization code returned by the OIDC provider
	Code string
	// State parameter returned by the OIDC provider
	State string
	// Optional error parameter if the user denied access
	Error string
	// Optional error description
	ErrorDescription string
}

// OidcCallbackResponse represents the output of OIDC callback
type OidcCallbackResponse struct {
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
	// Groups the user belongs to
	Groups []string `json:"groups,omitempty"`
	// ErrorMessage contains any error details
	ErrorMessage string `json:"error_message,omitempty"`
}

// NewOidcCallbackCommand creates a new OIDC callback command
func NewOidcCallbackCommand(
	authFactory *service.AuthFactory,
	sessionService *service.SessionService,
	config *infraConfig.InfrastructureConfig,
	oidcService *service.OIDCService,
	loginCommand *OidcLoginCommand,
) *OidcCallbackCommand {
	return &OidcCallbackCommand{
		authFactory:    authFactory,
		sessionService: sessionService,
		config:         config,
		oidcService:    oidcService,
		loginCommand:   loginCommand,
	}
}

// Execute executes the OIDC callback command
// Implements the complete OIDC authentication callback flow
func (c *OidcCallbackCommand) Execute(ctx context.Context, req *OidcCallbackRequest) (*OidcCallbackResponse, error) {
	// Validate request
	if err := c.ValidateRequest(req); err != nil {
		return &OidcCallbackResponse{
			Authenticated: false,
			ErrorMessage:  fmt.Sprintf("Invalid request: %v", err),
		}, nil
	}

	// Check if OIDC authentication is configured
	if !c.IsConfigured() {
		return &OidcCallbackResponse{
			Authenticated: false,
			ErrorMessage:  "OIDC authentication is not configured",
		}, nil
	}

	// Check for authentication error
	if req.Error != "" {
		errorMsg := req.Error
		if req.ErrorDescription != "" {
			errorMsg = fmt.Sprintf("%s: %s", req.Error, req.ErrorDescription)
		}
		return &OidcCallbackResponse{
			Authenticated: false,
			ErrorMessage:  errorMsg,
		}, nil
	}

	// Use real OIDC service if available
	if c.oidcService != nil {
		// Retrieve stored OIDC session for state validation
		var oidcSession *service.OIDCSession
		if c.loginCommand != nil {
			oidcSession, _ = c.loginCommand.GetOIDCSession(req.State)
			// Clean up the session after retrieval
			defer c.loginCommand.RemoveOIDCSession(req.State)
		}

		// If no stored session found, create a minimal one for compatibility
		if oidcSession == nil {
			oidcSession = &service.OIDCSession{
				State:     req.State,
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(10 * time.Minute),
			}
		}

		// Exchange code for user information
		userInfo, err := c.oidcService.ExchangeCode(ctx, oidcSession, req.Code, req.State)
		if err != nil {
			return &OidcCallbackResponse{
				Authenticated: false,
				ErrorMessage:  fmt.Sprintf("Failed to exchange OIDC code: %v", err),
			}, nil
		}

		// Create session with user information
		ttl := 24 * time.Hour
		authMethod := "OIDC"

		providerData := map[string]interface{}{
			"issuer":     c.config.OpenIDConnectIssuer,
			"client_id":  c.config.OpenIDConnectClientID,
			"subject":    userInfo.Subject,
			"username":   userInfo.Username,
			"email":      userInfo.Email,
			"name":       userInfo.Name,
			"groups":     userInfo.Groups,
			"raw_claims": userInfo.RawClaims,
		}
		providerDataBytes, _ := json.Marshal(providerData)

		session, err := c.sessionService.CreateSession(ctx, authMethod, providerDataBytes, &ttl)
		if err != nil {
			return &OidcCallbackResponse{
				Authenticated: false,
				ErrorMessage:  fmt.Sprintf("Failed to create session: %v", err),
			}, nil
		}

		return &OidcCallbackResponse{
			Authenticated: true,
			SessionID:     session.ID,
			Expiry:        session.Expiry,
			Username:      userInfo.Username,
			Email:         userInfo.Email,
			Groups:        userInfo.Groups,
		}, nil
	}

	// OIDC service must be available for proper OIDC authentication
	return &OidcCallbackResponse{
		Authenticated: false,
		ErrorMessage:  "OIDC service is not properly configured - please check OIDC configuration settings",
	}, nil
}

// ValidateRequest validates the OIDC callback request before execution
func (c *OidcCallbackCommand) ValidateRequest(req *OidcCallbackRequest) error {
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

// IsConfigured checks if OIDC authentication is properly configured
func (c *OidcCallbackCommand) IsConfigured() bool {
	return c.config != nil &&
		c.config.OpenIDConnectIssuer != "" &&
		c.config.OpenIDConnectClientID != "" &&
		c.config.OpenIDConnectClientSecret != "" &&
		c.authFactory != nil &&
		c.sessionService != nil &&
		c.oidcService != nil
}
