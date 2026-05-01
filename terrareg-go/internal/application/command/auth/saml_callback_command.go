package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// SamlCallbackCommand handles the SAML authentication callback
// Follows DDD principles by encapsulating the complete SAML callback flow
type SamlCallbackCommand struct {
	authFactory    *service.AuthFactory
	sessionService *service.SessionService
	config         *infraConfig.InfrastructureConfig
	samlService    *service.SAMLService
}

// SamlCallbackRequest represents the input for SAML callback
type SamlCallbackRequest struct {
	// SAMLResponse returned by the SAML provider
	SAMLResponse string
	// RelayState returned by the SAML provider
	RelayState string
}

// SamlCallbackResponse represents the output of SAML callback
type SamlCallbackResponse struct {
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

// NewSamlCallbackCommand creates a new SAML callback command
func NewSamlCallbackCommand(
	authFactory *service.AuthFactory,
	sessionService *service.SessionService,
	config *infraConfig.InfrastructureConfig,
	samlService *service.SAMLService,
) *SamlCallbackCommand {
	return &SamlCallbackCommand{
		authFactory:    authFactory,
		sessionService: sessionService,
		config:         config,
		samlService:    samlService,
	}
}

// Execute executes the SAML callback command
// Implements the complete SAML authentication callback flow
func (c *SamlCallbackCommand) Execute(ctx context.Context, req *SamlCallbackRequest) (*SamlCallbackResponse, error) {
	// Validate request
	if err := c.ValidateRequest(req); err != nil {
		return &SamlCallbackResponse{
			Authenticated: false,
			ErrorMessage:  fmt.Sprintf("Invalid request: %v", err),
		}, nil
	}

	// Check if SAML authentication is configured
	if !c.IsConfigured() {
		return &SamlCallbackResponse{
			Authenticated: false,
			ErrorMessage:  "SAML authentication is not configured",
		}, nil
	}

	// Use real SAML service if available
	if c.samlService != nil {
		// Exchange SAML response for user information
		userInfo, err := c.samlService.ProcessResponse(ctx, req.SAMLResponse, req.RelayState)
		if err != nil {
			return &SamlCallbackResponse{
				Authenticated: false,
				ErrorMessage:  fmt.Sprintf("Failed to process SAML response: %v", err),
			}, nil
		}

		// Create session with user information
		ttl := 24 * time.Hour
		authMethod := "SAML"

		providerData := map[string]interface{}{
			"entity_id":  c.config.SAML2EntityID,
			"name_id":    userInfo.NameID,
			"username":   userInfo.Username,
			"email":      userInfo.Email,
			"name":       userInfo.Name,
			"groups":     userInfo.Groups,
			"attributes": userInfo.Attributes,
		}
		providerDataBytes, _ := json.Marshal(providerData)

		session, err := c.sessionService.CreateSession(ctx, authMethod, providerDataBytes, &ttl)
		if err != nil {
			return &SamlCallbackResponse{
				Authenticated: false,
				ErrorMessage:  fmt.Sprintf("Failed to create session: %v", err),
			}, nil
		}

		return &SamlCallbackResponse{
			Authenticated: true,
			SessionID:     session.ID,
			Expiry:        session.Expiry,
			Username:      userInfo.Username,
			Email:         userInfo.Email,
			Groups:        userInfo.Groups,
		}, nil
	}

	// Fallback to placeholder response if SAML service is not available
	username := "saml-user"
	email := "user@example.com"
	groups := []string{"saml-users"}

	// Create session with a default TTL
	ttl := 24 * time.Hour
	authMethod := "SAML"

	providerData := map[string]interface{}{
		"entity_id": c.config.SAML2EntityID,
		"username":  username,
		"email":     email,
		"groups":    groups,
	}
	providerDataBytes, _ := json.Marshal(providerData)

	session, err := c.sessionService.CreateSession(ctx, authMethod, providerDataBytes, &ttl)
	if err != nil {
		return &SamlCallbackResponse{
			Authenticated: false,
			ErrorMessage:  fmt.Sprintf("Failed to create session: %v", err),
		}, nil
	}

	return &SamlCallbackResponse{
		Authenticated: true,
		SessionID:     session.ID,
		Expiry:        session.Expiry,
		Username:      username,
		Email:         email,
		Groups:        groups,
	}, nil
}

// ValidateRequest validates the SAML callback request before execution
func (c *SamlCallbackCommand) ValidateRequest(req *SamlCallbackRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.SAMLResponse == "" {
		return fmt.Errorf("SAML response cannot be empty")
	}

	return nil
}

// IsConfigured checks if SAML authentication is properly configured
func (c *SamlCallbackCommand) IsConfigured() bool {
	return c.config != nil &&
		c.config.SAML2EntityID != "" &&
		c.config.SAML2IDPMetadataURL != "" &&
		c.authFactory != nil &&
		c.sessionService != nil &&
		c.samlService != nil
}
