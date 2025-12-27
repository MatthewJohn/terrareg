package terraform

import (
	"context"
	"fmt"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
)

// AuthenticateOIDCTokenCommand handles OIDC token authentication
type AuthenticateOIDCTokenCommand struct {
	authFactory *authservice.AuthFactory
}

// NewAuthenticateOIDCTokenCommand creates a new OIDC token authentication command
func NewAuthenticateOIDCTokenCommand(authFactory *authservice.AuthFactory) *AuthenticateOIDCTokenCommand {
	return &AuthenticateOIDCTokenCommand{
		authFactory: authFactory,
	}
}

// Request represents the OIDC authentication request
type AuthenticateOIDCTokenRequest struct {
	AuthorizationHeader string `json:"authorization_header"`
}

// Response represents the OIDC authentication response
type AuthenticateOIDCTokenResponse struct {
	IdentityID   string   `json:"identity_id"`
	Subject      string   `json:"subject"`
	Permissions  []string `json:"permissions"`
	IdentityType string   `json:"identity_type"`
}

// Execute authenticates an OIDC token and returns identity information
func (c *AuthenticateOIDCTokenCommand) Execute(ctx context.Context, req AuthenticateOIDCTokenRequest) (*AuthenticateOIDCTokenResponse, error) {
	if req.AuthorizationHeader == "" {
		return nil, fmt.Errorf("authorization header is required")
	}

	// Create mock request headers for Terraform OIDC validation
	headers := map[string]string{
		"Authorization": req.AuthorizationHeader,
	}

	// Authenticate the request
	_, err := c.authFactory.AuthenticateRequest(ctx, headers, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Check if this is Terraform OIDC
	authMethod := c.authFactory.GetCurrentAuthMethod()
	if authMethod == nil || authMethod.GetProviderType() != auth.AuthMethodTerraformOIDC {
		return nil, fmt.Errorf("expected Terraform OIDC authentication, got: %s", authMethod.GetProviderType())
	}

	// Get authentication context
	authCtx := c.authFactory.GetCurrentAuthContext()
	if authCtx == nil || !authCtx.IsAuthenticated() {
		return nil, fmt.Errorf("authentication failed")
	}

	// Convert permissions to string slice
	permissions := make([]string, 0)
	for namespace, permType := range authCtx.GetAllNamespacePermissions() {
	permissions = append(permissions, fmt.Sprintf("%s:%s", namespace, permType))
	}

	// Generate identity ID
	identityID := "terraform-oidc-user"
	// Try to get session ID from provider data
	if providerData := authCtx.GetProviderData(); providerData != nil {
		if sessionID, ok := providerData["session_id"].(string); ok && sessionID != "" {
			identityID = sessionID
		}
	}

	return &AuthenticateOIDCTokenResponse{
		IdentityID:   identityID,
		Subject:      authCtx.GetUsername(),
		Permissions:  permissions,
		IdentityType: string(authMethod.GetProviderType()),
	}, nil
}

// ValidateTokenCommand handles token validation
type ValidateTokenCommand struct {
	authFactory *authservice.AuthFactory
}

// NewValidateTokenCommand creates a new token validation command
func NewValidateTokenCommand(authFactory *authservice.AuthFactory) *ValidateTokenCommand {
	return &ValidateTokenCommand{
		authFactory: authFactory,
	}
}

// Request represents the token validation request
type ValidateTokenRequest struct {
	AuthorizationHeader string   `json:"authorization_header"`
	RequiredPermissions []string `json:"required_permissions"`
}

// Response represents the token validation response
type ValidateTokenResponse struct {
	Valid        bool     `json:"valid"`
	IdentityID   string   `json:"identity_id,omitempty"`
	Subject      string   `json:"subject,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
	IdentityType string   `json:"identity_type,omitempty"`
}

// Execute validates a token and returns validation result
func (c *ValidateTokenCommand) Execute(ctx context.Context, req ValidateTokenRequest) (*ValidateTokenResponse, error) {
	if req.AuthorizationHeader == "" {
		return &ValidateTokenResponse{Valid: false}, fmt.Errorf("authorization header is required")
	}

	// Create mock request headers
	headers := map[string]string{
		"Authorization": req.AuthorizationHeader,
	}

	// Authenticate the request
	_, err := c.authFactory.AuthenticateRequest(ctx, headers, nil, nil)
	if err != nil {
		return &ValidateTokenResponse{Valid: false}, nil
	}

	// Get auth method
	authMethod := c.authFactory.GetCurrentAuthMethod()
	if authMethod == nil {
		return &ValidateTokenResponse{Valid: false}, nil
	}

	// Get auth context
	authCtx := c.authFactory.GetCurrentAuthContext()
	if authCtx == nil || !authCtx.IsAuthenticated() {
		return &ValidateTokenResponse{Valid: false}, nil
	}

	// Validate required permissions
	if len(req.RequiredPermissions) > 0 {
		for _, requiredPerm := range req.RequiredPermissions {
			// Parse required permission format "namespace:permission_type"
			parts := strings.SplitN(requiredPerm, ":", 2)
			if len(parts) == 2 {
				// Check if user has the specific permission
				if !authCtx.CheckNamespaceAccess(parts[1], parts[0]) {
					return &ValidateTokenResponse{Valid: false}, nil
				}
			}
		}
	}

	// Convert permissions to string slice
	permissions := make([]string, 0)
	for namespace, permType := range authCtx.GetAllNamespacePermissions() {
	permissions = append(permissions, fmt.Sprintf("%s:%s", namespace, permType))
	}

	// Generate identity ID
	identityID := "terraform-user"
	// Try to get session ID from provider data
	if providerData := authCtx.GetProviderData(); providerData != nil {
		if sessionID, ok := providerData["session_id"].(string); ok && sessionID != "" {
			identityID = sessionID
		}
	}

	return &ValidateTokenResponse{
		Valid:        true,
		IdentityID:   identityID,
		Subject:      authCtx.GetUsername(),
		Permissions:  permissions,
		IdentityType: string(authMethod.GetProviderType()),
	}, nil
}

// GetUserCommand handles user retrieval (legacy - to be removed)
// This is kept for compatibility but should use auth context instead
type GetUserCommand struct {
	authFactory *authservice.AuthFactory
}

// NewGetUserCommand creates a new get user command
func NewGetUserCommand(authFactory *authservice.AuthFactory) *GetUserCommand {
	return &GetUserCommand{
		authFactory: authFactory,
	}
}

// Response represents the user response
type GetUserResponse struct {
	IdentityID   string            `json:"identity_id"`
	Subject      string            `json:"subject"`
	Permissions  []string          `json:"permissions"`
	IdentityType string            `json:"identity_type"`
	IsValid      bool              `json:"is_valid"`
	Metadata     map[string]string `json:"metadata"`
}

// Execute retrieves authentication context information
func (c *GetUserCommand) Execute(ctx context.Context, identityID string) (*GetUserResponse, error) {
	if identityID == "" {
		return nil, fmt.Errorf("identity ID is required")
	}

	// Get current auth context (since we don't have user lookup by ID in new system)
	authCtx := c.authFactory.GetCurrentAuthContext()
	if authCtx == nil {
		return &GetUserResponse{
			IdentityID: identityID,
			IsValid:    false,
		}, nil
	}

	// Check if session ID matches (using SessionID as identity identifier)
	sessionID := ""
	if providerData := authCtx.GetProviderData(); providerData != nil {
		if id, ok := providerData["session_id"].(string); ok {
			sessionID = id
		}
	}
	sessionMatches := sessionID != "" && sessionID == identityID
	if !sessionMatches && identityID != "current-user" {
		return &GetUserResponse{
			IdentityID: identityID,
			IsValid:    false,
		}, nil
	}

	authMethod := c.authFactory.GetCurrentAuthMethod()
	if authMethod == nil {
		return &GetUserResponse{
			IdentityID: identityID,
			IsValid:    false,
		}, nil
	}

	// Convert permissions to string slice
	permissions := make([]string, 0)
	for namespace, permType := range authCtx.GetAllNamespacePermissions() {
	permissions = append(permissions, fmt.Sprintf("%s:%s", namespace, permType))
	}

	// Create metadata
	metadata := make(map[string]string)
	metadata["auth_method"] = string(authCtx.GetProviderType())
	if authCtx.IsAdmin() {
		metadata["is_admin"] = "true"
	}

	// Use provided identityID or fallback to session ID
	responseIdentityID := identityID
	if sessionMatches && sessionID != "" {
		responseIdentityID = sessionID
	}

	return &GetUserResponse{
		IdentityID:   responseIdentityID,
		Subject:      authCtx.GetUsername(),
		Permissions:  permissions,
		IdentityType: string(authMethod.GetProviderType()),
		IsValid:      true,
		Metadata:     metadata,
	}, nil
}
