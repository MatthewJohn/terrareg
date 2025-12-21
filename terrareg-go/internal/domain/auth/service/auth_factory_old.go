package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	domainAuthModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/rs/zerolog"
)

// TerraformIDPServiceAdapter wraps the domain service to implement the infrastructure interface
type TerraformIDPServiceAdapter struct {
	service interface{}
}

// ValidateToken implements the infrastructure interface
func (a *TerraformIDPServiceAdapter) ValidateToken(ctx context.Context, token string) (interface{}, error) {
	// Use type assertion to call the domain service method
	if validator, ok := a.service.(interface {
		ValidateToken(ctx context.Context, token string) (interface{}, error)
	}); ok {
		return validator.ValidateToken(ctx, token)
	}
	return nil, fmt.Errorf("service does not implement ValidateToken")
}

// AuthFactory implements the factory pattern for authentication methods
// Matches Python's AuthFactory behavior with priority-ordered discovery
// Uses immutable auth methods internally
type AuthFactory struct {
	immutableFactory *ImmutableAuthFactory
	mutex            sync.RWMutex
}

// NewAuthFactory creates a new authentication factory
func NewAuthFactory(
	sessionRepo repository.SessionRepository,
	userGroupRepo repository.UserGroupRepository,
	config *infraConfig.InfrastructureConfig,
	terraformIDPService interface{},
	logger *zerolog.Logger,
) *AuthFactory {
	// Create the immutable factory internally
	immutableFactory := NewImmutableAuthFactory(
		sessionRepo,
		userGroupRepo,
		config,
		logger,
	)

	factory := &AuthFactory{
		immutableFactory: immutableFactory,
	}

	return factory
}

// AuthenticateRequest authenticates an HTTP request
func (af *AuthFactory) AuthenticateRequest(ctx context.Context, headers, formData, queryParams map[string]string) (*domainAuthModel.AuthenticationResponse, error) {
	// Delegate to the immutable factory
	return af.immutableFactory.AuthenticateRequest(ctx, headers, formData, queryParams)
}

// GetCurrentAuthMethod returns the current auth method
// Note: This is maintained for compatibility but should be deprecated in favor of per-request auth
func (af *AuthFactory) GetCurrentAuthMethod() auth.AuthMethod {
	// With immutable auth methods, there's no "current" auth method
	// Return a not authenticated method as a fallback
	return &NotAuthenticatedAuthMethod{}
}

// GetCurrentAuthContext returns the current auth context
// Note: This is maintained for compatibility but should be deprecated
func (af *AuthFactory) GetCurrentAuthContext() *auth.AuthContext {
	// With immutable auth methods, there's no stored auth context
	// Return an empty auth context as a fallback
	return auth.NewAuthContext(&NotAuthenticatedAuthMethod{})
}

// CheckNamespacePermission checks if the current auth context has permission for a namespace
func (af *AuthFactory) CheckNamespacePermission(namespace, permissionType string) bool {
	// With immutable auth methods, this needs to be called with a specific auth context
	// For backward compatibility, return false
	return false
}

// CanPublishModuleVersion checks if current user can publish to a namespace
func (af *AuthFactory) CanPublishModuleVersion(namespace string) bool {
	// With immutable auth methods, this needs to be called with a specific auth context
	// For backward compatibility, return false
	return false
}

// CanUploadModuleVersion checks if current user can upload to a namespace
func (af *AuthFactory) CanUploadModuleVersion(namespace string) bool {
	// With immutable auth methods, this needs to be called with a specific auth context
	// For backward compatibility, return false
	return false
}

// CanAccessReadAPI checks if current user can access the read API
func (af *AuthFactory) CanAccessReadAPI() bool {
	// With immutable auth methods, this needs to be called with a specific auth context
	// For backward compatibility, return false
	return false
}

// CanAccessTerraformAPI checks if current user can access Terraform API
func (af *AuthFactory) CanAccessTerraformAPI() bool {
	// With immutable auth methods, this needs to be called with a specific auth context
	// For backward compatibility, return false
	return false
}

// InvalidateAuthentication clears the current authentication state
// Note: With immutable auth methods, there's no state to invalidate
func (af *AuthFactory) InvalidateAuthentication() {
	// No-op with immutable auth methods
}