package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
)

// TerraformOidcAuthMethod handles Terraform CLI OIDC authentication
// Matches Python's TerraformOidcAuthMethod exactly
type TerraformOidcAuthMethod struct {
	auth.BaseAuthMethod
	idp             TerraformIDP
	isAuthenticated bool
	validation      *TerraformTokenValidation
	userinfo        *TerraformUserinfoResponse
}

// TerraformIDP interface for Terraform identity provider
type TerraformIDP interface {
	IsEnabled() bool
	HandleUserinfoRequest(data []byte, headers map[string]string) (*TerraformUserinfoResponse, error)
	ValidateAccessToken(token string) (*TerraformTokenValidation, error)
}

// TerraformUserinfoResponse represents userinfo response
type TerraformUserinfoResponse struct {
	Subject  string                 `json:"sub"`
	Name     string                 `json:"name"`
	Username string                 `json:"preferred_username"`
	Email    string                 `json:"email"`
	Groups   []string               `json:"groups,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TerraformTokenValidation represents token validation result
type TerraformTokenValidation struct {
	Valid    bool   `json:"valid"`
	Subject  string `json:"sub"`
	Username string `json:"username"`
	Expiry   int64  `json:"exp"`
}

// NewTerraformOidcAuthMethod creates a new Terraform OIDC auth method
func NewTerraformOidcAuthMethod(idp TerraformIDP) *TerraformOidcAuthMethod {
	return &TerraformOidcAuthMethod{
		idp: idp,
	}
}

// CheckAuthState checks if the current request has valid Terraform OIDC authentication
// Matches Python's implementation - validates Authorization header
func (t *TerraformOidcAuthMethod) CheckAuthState() bool {
	return t.isAuthenticated
}

// Authenticate authenticates a request using Terraform OIDC (implements AuthMethod interface)
func (t *TerraformOidcAuthMethod) Authenticate(ctx context.Context, headers map[string]string, cookies map[string]string) error {
	if !t.IsEnabled() {
		return terraformOidcErr("IDP not enabled")
	}

	// Extract Authorization header
	authHeader, exists := headers["Authorization"]
	if !exists {
		return terraformOidcErr("missing authorization header")
	}

	// Remove "Bearer " prefix
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return terraformOidcErr("invalid authorization header format")
	}

	// Validate access token
	validation, err := t.idp.ValidateAccessToken(token)
	if err != nil {
		return terraformOidcErr("token validation failed: " + err.Error())
	}

	if !validation.Valid {
		return terraformOidcErr("invalid access token")
	}

	// Store validation result
	t.validation = validation
	t.isAuthenticated = true

	// Optional: Handle userinfo request if needed
	// In a full implementation, you might call userinfo endpoint
	// t.userinfo, err = t.idp.HandleUserinfoRequest(nil, headers)

	return nil
}

// AuthenticateRequest authenticates a Terraform request
func (t *TerraformOidcAuthMethod) AuthenticateRequest(ctx context.Context, headers map[string]string, data []byte) (*model.TerraformAuthResponse, error) {
	if !t.IsEnabled() {
		return &model.TerraformAuthResponse{
			Valid:    false,
			Username: "",
			Metadata: map[string]interface{}{"error": "IDP not enabled"},
		}, nil
	}

	// Extract Authorization header
	authHeader := headers["Authorization"]
	if authHeader == "" {
		return &model.TerraformAuthResponse{
			Valid:    false,
			Username: "",
			Metadata: map[string]interface{}{"error": "Missing Authorization header"},
		}, nil
	}

	// Remove "Bearer " prefix
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return &model.TerraformAuthResponse{
			Valid:    false,
			Username: "",
			Metadata: map[string]interface{}{"error": "Invalid Authorization header format"},
		}, nil
	}

	// Validate access token
	validation, err := t.idp.ValidateAccessToken(token)
	if err != nil {
		return &model.TerraformAuthResponse{
			Valid:    false,
			Username: "",
			Metadata: map[string]interface{}{"error": fmt.Sprintf("Token validation failed: %s", err.Error())},
		}, nil
	}

	if !validation.Valid {
		return &model.TerraformAuthResponse{
			Valid:    false,
			Username: "",
			Metadata: map[string]interface{}{"error": "Invalid access token"},
		}, nil
	}

	// Handle userinfo request
	userinfo, err := t.idp.HandleUserinfoRequest(data, headers)
	if err != nil {
		return &model.TerraformAuthResponse{
			Valid:    false,
			Username: validation.Username, // Use token validation username as fallback
			Metadata: map[string]interface{}{"error": fmt.Sprintf("Userinfo request failed: %s", err.Error())},
		}, nil
	}

	// Create response
	response := &model.TerraformAuthResponse{
		Valid:             true,
		SubjectIdentifier: userinfo.Subject,
		Username:          userinfo.Username,
		Permissions:       []string{}, // Terraform CLI typically gets all permissions
		Metadata:          userinfo.Metadata,
	}

	return response, nil
}

// IsBuiltInAdmin checks if this is the built-in admin user
func (t *TerraformOidcAuthMethod) IsBuiltInAdmin() bool {
	return false
}

// IsAdmin checks if the authenticated user has admin privileges
func (t *TerraformOidcAuthMethod) IsAdmin() bool {
	// Terraform CLI users are typically not admin users
	// Admin operations should use other auth methods
	return false
}

// IsAuthenticated checks if the authentication method is authenticated
func (t *TerraformOidcAuthMethod) IsAuthenticated() bool {
	return t.isAuthenticated
}

// IsEnabled checks if this authentication method is enabled
func (t *TerraformOidcAuthMethod) IsEnabled() bool {
	return t.idp != nil && t.idp.IsEnabled()
}

// RequiresCSRF checks if CSRF protection is required
func (t *TerraformOidcAuthMethod) RequiresCSRF() bool {
	// Terraform CLI doesn't use web forms, so no CSRF required
	return false
}

// CanPublishModuleVersion checks if user can publish modules
func (t *TerraformOidcAuthMethod) CanPublishModuleVersion(namespace string) bool {
	// Terraform CLI typically has read-only access
	// Publishing should be done via other auth methods
	return false
}

// CanUploadModuleVersion checks if user can upload modules
func (t *TerraformOidcAuthMethod) CanUploadModuleVersion(namespace string) bool {
	// Terraform CLI typically has read-only access
	// Uploading should be done via other auth methods
	return false
}

// CheckNamespaceAccess checks if user has access to a namespace
func (t *TerraformOidcAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	// Terraform CLI typically gets read access to all namespaces
	// This can be overridden by checking user groups in the implementation
	return permissionType == "READ"
}

// GetAllNamespacePermissions returns all namespace permissions
func (t *TerraformOidcAuthMethod) GetAllNamespacePermissions() map[string]string {
	// Terraform CLI typically gets read access to all namespaces
	// In practice, this would be populated from user group memberships
	return make(map[string]string) // Will be populated by AuthFactory
}

// GetUsername returns the authenticated username
func (t *TerraformOidcAuthMethod) GetUsername() string {
	// Return username from validation or userinfo, fallback to generic name
	if t.validation != nil && t.validation.Username != "" {
		return t.validation.Username
	}
	if t.userinfo != nil && t.userinfo.Username != "" {
		return t.userinfo.Username
	}
	return "Terraform CLI User"
}

// GetUserGroupNames returns the user groups the user belongs to
func (t *TerraformOidcAuthMethod) GetUserGroupNames() []string {
	// Return groups from userinfo if available
	if t.userinfo != nil && len(t.userinfo.Groups) > 0 {
		return t.userinfo.Groups
	}
	return []string{}
}

// CanAccessReadAPI checks if user can access the read API
func (t *TerraformOidcAuthMethod) CanAccessReadAPI() bool {
	// Terraform CLI can access the read API for module discovery
	return true
}

// CanAccessTerraformAPI checks if user can access Terraform-specific APIs
func (t *TerraformOidcAuthMethod) CanAccessTerraformAPI() bool {
	// This auth method is specifically for Terraform API access
	return true
}

// GetTerraformAuthToken returns the Terraform auth token
func (t *TerraformOidcAuthMethod) GetTerraformAuthToken() string {
	// Terraform OIDC doesn't use traditional tokens
	// It validates the Authorization header directly
	return ""
}

// GetProviderType returns the provider type
func (t *TerraformOidcAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodTerraformOIDC
}

// GetProviderData returns provider-specific data
func (t *TerraformOidcAuthMethod) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"provider": "terraform_oidc",
		"type":     "oidc",
		"cli":      true,
	}
}

// MockTerraformIDP is a mock implementation for testing
type MockTerraformIDP struct {
	enabled bool
}

// NewMockTerraformIDP creates a mock Terraform IDP for testing
func NewMockTerraformIDP(enabled bool) *MockTerraformIDP {
	return &MockTerraformIDP{
		enabled: enabled,
	}
}

// IsEnabled returns whether the IDP is enabled
func (m *MockTerraformIDP) IsEnabled() bool {
	return m.enabled
}

// HandleUserinfoRequest handles a userinfo request (mock)
func (m *MockTerraformIDP) HandleUserinfoRequest(data []byte, headers map[string]string) (*TerraformUserinfoResponse, error) {
	return &TerraformUserinfoResponse{
		Subject:  "terraform-cli-user",
		Name:     "Terraform CLI User",
		Username: "terraform-cli",
		Email:    "terraform@cli.local",
		Groups:   []string{},
		Metadata: map[string]interface{}{
			"mock": true,
		},
	}, nil
}

// ValidateAccessToken validates an access token (mock)
func (m *MockTerraformIDP) ValidateAccessToken(token string) (*TerraformTokenValidation, error) {
	if token == "invalid-token" {
		return &TerraformTokenValidation{
			Valid:    false,
			Subject:  "",
			Username: "",
		}, nil
	}

	return &TerraformTokenValidation{
		Valid:    true,
		Subject:  "terraform-cli-user",
		Username: "terraform-cli",
		Expiry:   0, // No expiry for mock
	}, nil
}

// terraformOidcErr creates a formatted error message for Terraform OIDC authentication
func terraformOidcErr(message string) error {
	return &TerraformOidcError{Message: message}
}

// TerraformOidcError represents a Terraform OIDC authentication error
type TerraformOidcError struct {
	Message string
}

func (e *TerraformOidcError) Error() string {
	return "Terraform OIDC authentication failed: " + e.Message
}
