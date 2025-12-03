package terraform

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/service"
)

// AuthenticateOIDCTokenCommand handles OIDC token authentication
type AuthenticateOIDCTokenCommand struct {
	terraformAuthService *service.TerraformAuthServiceIntegrated
}

// NewAuthenticateOIDCTokenCommand creates a new OIDC token authentication command
func NewAuthenticateOIDCTokenCommand(terraformAuthService *service.TerraformAuthServiceIntegrated) *AuthenticateOIDCTokenCommand {
	return &AuthenticateOIDCTokenCommand{
		terraformAuthService: terraformAuthService,
	}
}

// Request represents the OIDC authentication request
type AuthenticateOIDCTokenRequest struct {
	AccessToken string `json:"access_token"`
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
	if req.AccessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	// Extract subject from token (this would be done by infrastructure layer)
	// For now, assume we have a way to extract subject
	subject := "terraform-oidc-user" // This should be extracted from JWT

	// Authenticate using domain service
	user, err := c.terraformAuthService.AuthenticateOIDCIdentity(ctx, subject, req.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("OIDC authentication failed: %w", err)
	}

	// Convert user permissions to strings
	permissions := make([]string, len(user.Permissions()))
	for i, perm := range user.Permissions() {
		permissions[i] = fmt.Sprintf("%s:%s:%s", perm.ResourceType(), perm.ResourceID(), perm.Action())
	}

	return &AuthenticateOIDCTokenResponse{
		IdentityID:   user.ID(),
		Subject:      user.Username(),
		Permissions:  permissions,
		IdentityType: user.AuthMethod().String(),
	}, nil
}

// ValidateTokenCommand handles token validation
type ValidateTokenCommand struct {
	terraformAuthService *service.TerraformAuthServiceIntegrated
}

// NewValidateTokenCommand creates a new token validation command
func NewValidateTokenCommand(terraformAuthService *service.TerraformAuthServiceIntegrated) *ValidateTokenCommand {
	return &ValidateTokenCommand{
		terraformAuthService: terraformAuthService,
	}
}

// Request represents the token validation request
type ValidateTokenRequest struct {
	Token               string   `json:"token"`
	TokenType           string   `json:"token_type"`
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
	if req.Token == "" {
		return &ValidateTokenResponse{Valid: false}, fmt.Errorf("token is required")
	}

	var user *model.User
	var err error

	switch req.TokenType {
	case "oidc":
		// For OIDC, we need to extract subject from token
		subject := "terraform-oidc-user" // This should be extracted from JWT
		user, err = c.terraformAuthService.AuthenticateOIDCIdentity(ctx, subject, req.Token)
	case "analytics", "internal-extraction", "deployment":
		user, err = c.terraformAuthService.AuthenticateStaticToken(ctx, req.Token, req.TokenType)
	default:
		return &ValidateTokenResponse{Valid: false}, fmt.Errorf("unsupported token type: %s", req.TokenType)
	}

	if err != nil {
		return &ValidateTokenResponse{Valid: false}, nil // Don't leak error details
	}

	// Validate required permissions
	if len(req.RequiredPermissions) > 0 {
		if err := c.terraformAuthService.ValidatePermissions(ctx, user, req.RequiredPermissions); err != nil {
			return &ValidateTokenResponse{Valid: false}, nil
		}
	}

	// Convert user permissions to strings
	permissions := make([]string, len(user.Permissions()))
	for i, perm := range user.Permissions() {
		permissions[i] = fmt.Sprintf("%s:%s:%s", perm.ResourceType(), perm.ResourceID(), perm.Action())
	}

	return &ValidateTokenResponse{
		Valid:        true,
		IdentityID:   user.ID(),
		Subject:      user.Username(),
		Permissions:  permissions,
		IdentityType: user.AuthMethod().String(),
	}, nil
}

// GetUserCommand handles user retrieval
type GetUserCommand struct {
	userRepo repository.UserRepository
}

// NewGetUserCommand creates a new get user command
func NewGetUserCommand(userRepo repository.UserRepository) *GetUserCommand {
	return &GetUserCommand{
		userRepo: userRepo,
	}
}

// Response represents the user response
type GetUserResponse struct {
	IdentityID   string   `json:"identity_id"`
	Subject      string   `json:"subject"`
	Permissions  []string `json:"permissions"`
	IdentityType string   `json:"identity_type"`
	IsValid      bool     `json:"is_valid"`
	Metadata     map[string]string `json:"metadata"`
}

// Execute retrieves a user by ID
func (c *GetUserCommand) Execute(ctx context.Context, userID string) (*GetUserResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := c.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert user permissions to strings
	permissions := make([]string, len(user.Permissions()))
	for i, perm := range user.Permissions() {
		permissions[i] = fmt.Sprintf("%s:%s:%s", perm.ResourceType(), perm.ResourceID(), perm.Action())
	}

	// Create metadata for Terraform users
	metadata := make(map[string]string)
	if user.AuthMethod() == model.AuthMethodTerraform {
		metadata["terraform_user"] = "true"
	}

	return &GetUserResponse{
		IdentityID:   user.ID(),
		Subject:      user.Username(),
		Permissions:  permissions,
		IdentityType: user.AuthMethod().String(),
		IsValid:      true, // Users in the system are considered valid
		Metadata:     metadata,
	}, nil
}