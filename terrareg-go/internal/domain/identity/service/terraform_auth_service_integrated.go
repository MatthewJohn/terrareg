package service

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
)

// TerraformAuthServiceIntegrated handles Terraform authentication using the existing User system
type TerraformAuthServiceIntegrated struct {
	userRepo repository.UserRepository
}

// NewTerraformAuthServiceIntegrated creates a new Terraform authentication service
func NewTerraformAuthServiceIntegrated(userRepo repository.UserRepository) *TerraformAuthServiceIntegrated {
	return &TerraformAuthServiceIntegrated{
		userRepo: userRepo,
	}
}

// AuthenticateOIDCIdentity authenticates an OIDC identity using the existing User system
func (s *TerraformAuthServiceIntegrated) AuthenticateOIDCIdentity(ctx context.Context, subject string, accessToken string) (*model.User, error) {
	if subject == "" {
		return nil, fmt.Errorf("subject cannot be empty")
	}

	// Check if user already exists by auth provider ID
	existingUser, err := s.userRepo.FindByAuthProviderID(ctx, model.AuthMethodTerraform, subject)
	if err == nil && existingUser != nil {
		// Update access token for existing user
		existingUser.SetAccessToken(accessToken)
		if err := s.userRepo.Update(ctx, existingUser); err != nil {
			return nil, fmt.Errorf("failed to update existing OIDC user: %w", err)
		}
		return existingUser, nil
	}

	// Create new OIDC user
	username := fmt.Sprintf("terraform-oidc-%s", subject)
	displayName := fmt.Sprintf("Terraform OIDC User (%s)", subject)

	user, err := model.NewUser(
		username,
		displayName,
		"", // email optional for Terraform
		model.AuthMethodTerraform,
		"TERRAFORM_OIDC", // auth provider ID
		subject, // external ID
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC user: %w", err)
	}

	// Set access token
	user.SetAccessToken(accessToken)

	// Add Terraform-specific permissions
	user.AddPermissionSimple(model.PermissionReadModules)
	user.AddPermissionSimple(model.PermissionReadProviders)

	// Save to repository
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save OIDC user: %w", err)
	}

	return user, nil
}

// AuthenticateStaticToken authenticates using static token with the existing User system
func (s *TerraformAuthServiceIntegrated) AuthenticateStaticToken(ctx context.Context, token, tokenType string) (*model.User, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	var permissions []model.Permission
	var username string
	var displayName string
	var externalID string

	switch tokenType {
	case "analytics":
		permissions = []model.Permission{model.PermissionReadAnalytics}
		username = "terraform-analytics-token"
		displayName = "Terraform Analytics Token"
		externalID = "terraform-analytics"
	case "internal-extraction":
		permissions = []model.Permission{model.PermissionReadModules, model.PermissionReadProviders}
		username = "terraform-internal-extraction-token"
		displayName = "Terraform Internal Extraction"
		externalID = "terraform-internal-extraction"
	case "deployment":
		permissions = []model.Permission{
			model.PermissionReadModules,
			model.PermissionReadProviders,
			model.PermissionPublishModules,
			model.PermissionPublishProviders,
		}
		username = "terraform-deployment-token"
		displayName = "Terraform Deployment Token"
		externalID = "terraform-deployment"
	default:
		return nil, fmt.Errorf("unsupported token type: %s", tokenType)
	}

	// Check if user already exists by auth provider ID
	existingUser, err := s.userRepo.FindByAuthProviderID(ctx, model.AuthMethodTerraform, externalID)
	if err == nil && existingUser != nil {
		// Validate the token matches (in a real implementation, this would check against config)
		// For now, assume the token is valid if we get here
		if existingUser.AccessToken() == token {
			return existingUser, nil
		}
		return nil, fmt.Errorf("invalid static token")
	}

	// Create new static token user
	user, err := model.NewUser(
		username,
		displayName,
		"", // email optional for Terraform tokens
		model.AuthMethodTerraform,
		"TERRAFORM_TOKEN", // auth provider ID
		externalID, // external ID
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create static token user: %w", err)
	}

	// Set the token as access token
	user.SetAccessToken(token)

	// Add permissions based on token type
	for _, permission := range permissions {
		user.AddPermissionSimple(permission)
	}

	// Save to repository
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save static token user: %w", err)
	}

	return user, nil
}

// ValidatePermissions checks if a user has all required permissions
func (s *TerraformAuthServiceIntegrated) ValidatePermissions(ctx context.Context, user *model.User, requiredPermissions []string) error {
	// Convert string permissions to model.Permission
	modelPermissions := make([]model.Permission, len(requiredPermissions))
	for i, perm := range requiredPermissions {
		switch perm {
		case "read:modules":
			modelPermissions[i] = model.PermissionReadModules
		case "read:providers":
			modelPermissions[i] = model.PermissionReadProviders
		case "write:modules":
			modelPermissions[i] = model.PermissionPublishModules
		case "write:providers":
			modelPermissions[i] = model.PermissionPublishProviders
		case "read:analytics":
			modelPermissions[i] = model.PermissionReadAnalytics
		default:
			// Skip unknown permissions
			modelPermissions[i] = model.Permission(999) // Invalid permission
		}
	}

	// Check each required permission
	for _, requiredPerm := range modelPermissions {
		if requiredPerm == model.Permission(999) {
			continue // Skip invalid permissions
		}
		if !user.HasPermission(requiredPerm) {
			return fmt.Errorf("user lacks required permission: %v", requiredPerm)
		}
	}

	return nil
}

// GetUserByAccessToken retrieves a user by access token
func (s *TerraformAuthServiceIntegrated) GetUserByAccessToken(ctx context.Context, accessToken string) (*model.User, error) {
	user, err := s.userRepo.FindByAccessToken(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by access token: %w", err)
	}

	// Verify this is a Terraform user
	if user.AuthMethod() != model.AuthMethodTerraform {
		return nil, fmt.Errorf("access token does not belong to a Terraform user")
	}

	return user, nil
}

// GetUserByAuthProviderID retrieves a user by auth provider ID
func (s *TerraformAuthServiceIntegrated) GetUserByAuthProviderID(ctx context.Context, authProviderID string) (*model.User, error) {
	user, err := s.userRepo.FindByAuthProviderID(ctx, model.AuthMethodTerraform, authProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by auth provider ID: %w", err)
	}

	return user, nil
}