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
) *OidcCallbackCommand {
	return &OidcCallbackCommand{
		authFactory:    authFactory,
		sessionService: sessionService,
		config:         config,
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

	// For now, return a placeholder response
	// In a full implementation, this would:
	// 1. Validate the state parameter
	// 2. Exchange the authorization code for tokens
	// 3. Validate the ID token
	// 4. Extract user information from the token
	// 5. Create a session for the user
	// 6. Return the session details

	// Placeholder user data
	username := "oidc-user"
	email := "user@example.com"
	groups := []string{"oidc-users"}

	// Create session with a default TTL
	ttl := 24 * time.Hour
	authMethod := "OIDC"

	providerData := map[string]interface{}{
		"issuer":       c.config.OpenIDConnectIssuer,
		"client_id":    c.config.OpenIDConnectClientID,
		"username":     username,
		"email":        email,
		"groups":       groups,
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
		Username:      username,
		Email:         email,
		Groups:        groups,
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
		c.sessionService != nil
}