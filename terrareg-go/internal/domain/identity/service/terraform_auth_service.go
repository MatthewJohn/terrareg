package service

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
)

// TerraformAuthService handles Terraform authentication domain logic
type TerraformAuthService struct {
	identityRepo repository.IdentityRepository
}

// NewTerraformAuthService creates a new Terraform authentication service
func NewTerraformAuthService(identityRepo repository.IdentityRepository) *TerraformAuthService {
	return &TerraformAuthService{
		identityRepo: identityRepo,
	}
}

// AuthenticateOIDCIdentity authenticates an OIDC identity
func (s *TerraformAuthService) AuthenticateOIDCIdentity(ctx context.Context, subject string, accessToken string) (*model.TerraformIdentity, error) {
	if subject == "" {
		return nil, fmt.Errorf("subject cannot be empty")
	}

	// Generate unique ID for this OIDC identity
	id := fmt.Sprintf("oidc-%s", subject)

	// Check if identity already exists
	existing, err := s.identityRepo.FindByID(ctx, id)
	if err == nil && existing != nil {
		// Update access token for existing identity
		if terraformIdentity, ok := existing.(*model.TerraformIdentity); ok {
			terraformIdentity.SetAccessToken(accessToken, nil)
			if err := s.identityRepo.Save(ctx, terraformIdentity); err != nil {
				return nil, fmt.Errorf("failed to update existing OIDC identity: %w", err)
			}
			return terraformIdentity, nil
		}
	}

	// Create new OIDC identity
	terraformIdentity, err := model.NewTerraformIdentity(id, subject, model.TerraformIdentityTypeOIDC)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC identity: %w", err)
	}

	// Set access token
	terraformIdentity.SetAccessToken(accessToken, nil)

	// Add OIDC-specific permissions
	terraformIdentity.AddPermission("read:modules")
	terraformIdentity.AddPermission("read:providers")

	// Save to repository
	if err := s.identityRepo.Save(ctx, terraformIdentity); err != nil {
		return nil, fmt.Errorf("failed to save OIDC identity: %w", err)
	}

	return terraformIdentity, nil
}

// AuthenticateStaticToken authenticates using static token
func (s *TerraformAuthService) AuthenticateStaticToken(ctx context.Context, token, tokenType string) (*model.TerraformIdentity, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	var identityType model.TerraformIdentityType
	var permissions []string
	var subject string

	switch tokenType {
	case "analytics":
		identityType = model.TerraformIdentityTypeAnalytics
		permissions = []string{"read:analytics"}
		subject = "terraform-analytics"
	case "internal-extraction":
		identityType = model.TerraformIdentityTypeInternalExtraction
		permissions = []string{"read:modules", "read:providers", "no-analytics"}
		subject = "terraform-internal-extraction"
	case "deployment":
		identityType = model.TerraformIdentityTypeDeployment
		permissions = []string{"read:modules", "read:providers", "write:modules", "write:providers"}
		subject = "terraform-deployment"
	default:
		return nil, fmt.Errorf("unsupported token type: %s", tokenType)
	}

	// Generate unique ID for this token identity
	id := fmt.Sprintf("%s-%s", tokenType, subject)

	// Check if identity already exists
	existing, err := s.identityRepo.FindByID(ctx, id)
	if err == nil && existing != nil {
		if terraformIdentity, ok := existing.(*model.TerraformIdentity); ok {
			// Validate the token matches
			if terraformIdentity.AccessToken() == token {
				terraformIdentity.SetAccessToken(token, nil)
				if err := s.identityRepo.Save(ctx, terraformIdentity); err != nil {
					return nil, fmt.Errorf("failed to update existing static token identity: %w", err)
				}
				return terraformIdentity, nil
			}
		}
		return nil, fmt.Errorf("invalid static token")
	}

	// For static tokens, we should validate against configuration
	// This would be handled by the infrastructure layer
	// For now, create the identity structure
	terraformIdentity, err := model.NewTerraformIdentity(id, subject, identityType)
	if err != nil {
		return nil, fmt.Errorf("failed to create static token identity: %w", err)
	}

	// Set the token as access token
	terraformIdentity.SetAccessToken(token, nil)

	// Add permissions based on token type
	for _, permission := range permissions {
		terraformIdentity.AddPermission(permission)
	}

	// Set metadata
	terraformIdentity.SetMetadata("token_type", tokenType)

	return terraformIdentity, nil
}

// ValidatePermissions checks if an identity has all required permissions
func (s *TerraformAuthService) ValidatePermissions(ctx context.Context, identity *model.TerraformIdentity, requiredPermissions []string) error {
	if !identity.IsValid() {
		return fmt.Errorf("identity is invalid or expired")
	}

	if !identity.HasAllPermissions(requiredPermissions) {
		return fmt.Errorf("identity lacks required permissions")
	}

	return nil
}

// GetIdentityByID retrieves an identity by ID
func (s *TerraformAuthService) GetIdentityByID(ctx context.Context, id string) (*model.TerraformIdentity, error) {
	identity, err := s.identityRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find identity: %w", err)
	}

	terraformIdentity, ok := identity.(*model.TerraformIdentity)
	if !ok {
		return nil, fmt.Errorf("identity is not a Terraform identity")
	}

	return terraformIdentity, nil
}

// RevokeIdentity revokes an identity
func (s *TerraformAuthService) RevokeIdentity(ctx context.Context, id string) error {
	identity, err := s.GetIdentityByID(ctx, id)
	if err != nil {
		return err
	}

	// Clear access token to effectively revoke
	identity.SetAccessToken("", nil)

	return s.identityRepo.Save(ctx, identity)
}