package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// AdminLoginCommand handles the admin login use case
// Follows DDD principles by encapsulating the complete login flow
type AdminLoginCommand struct {
	authFactory    *service.AuthFactory
	sessionService *service.SessionService
	config         *infraConfig.InfrastructureConfig
}

// AdminLoginRequest represents the input for admin login
type AdminLoginRequest struct {
	// ApiKey extracted from X-Terrareg-ApiKey header
	ApiKey string
}

// AdminLoginResponse represents the output of admin login
type AdminLoginResponse struct {
	Authenticated bool      `json:"authenticated"`
	SessionID     string    `json:"session_id"`
	Expiry        time.Time `json:"expiry"`
}

// NewAdminLoginCommand creates a new admin login command
func NewAdminLoginCommand(
	authFactory *service.AuthFactory,
	sessionService *service.SessionService,
	config *infraConfig.InfrastructureConfig,
) *AdminLoginCommand {
	return &AdminLoginCommand{
		authFactory:    authFactory,
		sessionService: sessionService,
		config:         config,
	}
}

// Execute executes the admin login command
// Implements the complete authentication flow using the auth factory
func (c *AdminLoginCommand) Execute(ctx context.Context, req *AdminLoginRequest) (*AdminLoginResponse, error) {
	// Validate request
	if req.ApiKey == "" {
		return &AdminLoginResponse{
			Authenticated: false,
		}, fmt.Errorf("missing API key")
	}

	// Use the auth factory to validate the API key
	headers := map[string]string{
		"X-Terrareg-ApiKey": req.ApiKey,
	}

	response, err := c.authFactory.AuthenticateRequest(ctx, headers, nil, nil)
	if err != nil {
		return &AdminLoginResponse{
			Authenticated: false,
		}, fmt.Errorf("authentication failed: %w", err)
	}

	// Verify authentication succeeded and this is an admin
	if !response.Success || !response.IsAdmin {
		return &AdminLoginResponse{
			Authenticated: false,
		}, fmt.Errorf("not an admin authentication token")
	}

	ttl := time.Duration(c.config.AdminSessionExpiryMins) * time.Minute

	// Create provider data for the session
	providerData := map[string]interface{}{
		"auth_method": "ADMIN_API_KEY",
		"is_admin":    true,
		"username":    response.Username,
	}
	providerDataBytes, _ := json.Marshal(providerData)

	session, err := c.sessionService.CreateSession(ctx, "ADMIN_API_KEY", providerDataBytes, &ttl)
	if err != nil {
		return &AdminLoginResponse{
			Authenticated: false,
		}, fmt.Errorf("failed to create session: %w", err)
	}

	// Return successful response
	return &AdminLoginResponse{
		Authenticated: true,
		SessionID:     session.ID,
		Expiry:        session.Expiry,
	}, nil
}

// ExecuteWithRequest is a convenience method that extracts API key from HTTP request
// This keeps the command pure while providing ease of use for HTTP handlers
func (c *AdminLoginCommand) ExecuteWithRequest(ctx context.Context, r *http.Request) (*AdminLoginResponse, error) {
	apiKey := r.Header.Get("X-Terrareg-ApiKey")
	if apiKey == "" {
		return &AdminLoginResponse{
			Authenticated: false,
		}, fmt.Errorf("missing X-Terrareg-ApiKey header")
	}

	return c.Execute(ctx, &AdminLoginRequest{
		ApiKey: apiKey,
	})
}

// ValidateRequest validates the admin login request before execution
func (c *AdminLoginCommand) ValidateRequest(req *AdminLoginRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.ApiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	if !c.IsConfigured() {
		return fmt.Errorf("admin authentication is not configured")
	}

	return nil
}

// IsConfigured checks if admin authentication is properly configured
func (c *AdminLoginCommand) IsConfigured() bool {
	// Check if config-based admin authentication is available
	return c.config.AdminAuthenticationToken != "" && c.sessionService != nil
}
