package auth

import (
	"context"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// TerraformInternalExtractionAuthMethod implements authentication for internal Terraform extraction processes
type TerraformInternalExtractionAuthMethod struct {
	auth.BaseAuthMethod
	authToken      string
	isValid        bool
	extractionType string
	accessScope    []string
}

// NewTerraformInternalExtractionAuthMethod creates a new Terraform internal extraction authentication method
func NewTerraformInternalExtractionAuthMethod() *TerraformInternalExtractionAuthMethod {
	return &TerraformInternalExtractionAuthMethod{
		isValid: false,
	}
}

// GetProviderType returns the authentication provider type
func (t *TerraformInternalExtractionAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodTerraformInternalExtraction
}

// CheckAuthState validates the current authentication state
func (t *TerraformInternalExtractionAuthMethod) CheckAuthState() bool {
	return t.isValid
}

// IsBuiltInAdmin returns whether this is a built-in admin method
func (t *TerraformInternalExtractionAuthMethod) IsBuiltInAdmin() bool {
	return true // Internal extraction processes have full system access
}

// IsAuthenticated returns whether the current request is authenticated
func (t *TerraformInternalExtractionAuthMethod) IsAuthenticated() bool {
	return t.isValid
}

// IsAdmin returns whether the authenticated user has admin privileges
func (t *TerraformInternalExtractionAuthMethod) IsAdmin() bool {
	return true // Internal extraction processes have admin privileges
}

// IsEnabled returns whether this authentication method is enabled
func (t *TerraformInternalExtractionAuthMethod) IsEnabled() bool {
	return true
}

// RequiresCSRF returns whether this authentication method requires CSRF protection
func (t *TerraformInternalExtractionAuthMethod) RequiresCSRF() bool {
	return false // Internal service-to-service auth doesn't need CSRF
}

// CanPublishModuleVersion checks if the user can publish module versions to the given namespace
func (t *TerraformInternalExtractionAuthMethod) CanPublishModuleVersion(namespace string) bool {
	// Internal extraction processes have full access for module discovery and cataloging
	return t.isValid && t.hasAccessScope("modules.publish")
}

// CanUploadModuleVersion checks if the user can upload module versions to the given namespace
func (t *TerraformInternalExtractionAuthMethod) CanUploadModuleVersion(namespace string) bool {
	// Internal extraction processes can upload extracted modules
	return t.isValid && t.hasAccessScope("modules.upload")
}

// CheckNamespaceAccess checks if the user has the specified permission for a namespace
func (t *TerraformInternalExtractionAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	if !t.isValid {
		return false
	}

	// Internal extraction has admin privileges - full access to all namespaces
	switch auth.PermissionType(permissionType) {
	case auth.PermissionRead, auth.PermissionModify, auth.PermissionFull:
		return t.hasAccessScope("modules.read") || t.hasAccessScope("modules.full")
	default:
		return false
	}
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (t *TerraformInternalExtractionAuthMethod) GetAllNamespacePermissions() map[string]string {
	if !t.isValid {
		return map[string]string{}
	}

	// Internal extraction has full access to all namespaces
	return map[string]string{
		"*": string(auth.PermissionFull),
	}
}

// GetUsername returns the authenticated username
func (t *TerraformInternalExtractionAuthMethod) GetUsername() string {
	if t.extractionType != "" {
		return "Terraform Internal Extraction (" + t.extractionType + ")"
	}
	return "Terraform Internal Extraction"
}

// GetUserGroupNames returns the names of all user groups
func (t *TerraformInternalExtractionAuthMethod) GetUserGroupNames() []string {
	// Internal processes belong to system groups
	return []string{"system", "internal", "extraction"}
}

// CanAccessReadAPI returns whether the user can access read APIs
func (t *TerraformInternalExtractionAuthMethod) CanAccessReadAPI() bool {
	// Internal extraction can access all read APIs
	return t.isValid && t.hasAccessScope("api.read")
}

// CanAccessTerraformAPI returns whether the user can access Terraform APIs
func (t *TerraformInternalExtractionAuthMethod) CanAccessTerraformAPI() bool {
	// Internal extraction can access Terraform APIs
	return t.isValid && t.hasAccessScope("terraform.access")
}

// GetTerraformAuthToken returns the Terraform authentication token
func (t *TerraformInternalExtractionAuthMethod) GetTerraformAuthToken() string {
	return t.authToken
}

// GetProviderData returns provider-specific data
func (t *TerraformInternalExtractionAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["auth_token"] = t.authToken
	data["extraction_type"] = t.extractionType
	data["access_scope"] = t.accessScope
	data["internal"] = true
	return data
}

// Authenticate authenticates a request using internal extraction auth token
func (t *TerraformInternalExtractionAuthMethod) Authenticate(ctx context.Context, headers map[string]string, cookies map[string]string) error {
	// Look for auth token in Authorization header
	authHeader, exists := headers["Authorization"]
	if !exists {
		// Also check for X-Internal-Extraction-Token header
		authHeader = headers["X-Internal-Extraction-Token"]
	}

	if authHeader == "" {
		return terraformInternalExtractionErr("missing authorization header or X-Internal-Extraction-Token")
	}

	// Handle different header formats
	var authToken string
	if strings.HasPrefix(authHeader, "Bearer ") {
		authToken = strings.TrimSpace(authHeader[7:]) // Remove "Bearer " prefix
	} else if strings.HasPrefix(authHeader, "Internal ") {
		authToken = strings.TrimSpace(authHeader[9:]) // Remove "Internal " prefix
	} else {
		// Use raw header value if no prefix
		authToken = strings.TrimSpace(authHeader)
	}

	if authToken == "" {
		return terraformInternalExtractionErr("empty internal extraction auth token")
	}

	// Validate internal extraction auth token
	tokenInfo, err := t.validateInternalExtractionToken(authToken)
	if err != nil {
		return err
	}

	t.authToken = authToken
	t.isValid = true
	t.extractionType = tokenInfo.ExtractionType
	t.accessScope = tokenInfo.AccessScope

	return nil
}

// hasAccessScope checks if the token has the specified access scope
func (t *TerraformInternalExtractionAuthMethod) hasAccessScope(scope string) bool {
	if t.accessScope == nil {
		return false
	}

	// Check for full access
	for _, s := range t.accessScope {
		if s == "full" || s == "*" {
			return true
		}
	}

	// Check for specific scope
	for _, s := range t.accessScope {
		if s == scope {
			return true
		}
	}

	return false
}

// TerraformInternalExtractionTokenInfo represents information about an internal extraction token
type TerraformInternalExtractionTokenInfo struct {
	ExtractionType string
	AccessScope    []string
}

// validateInternalExtractionToken validates an internal extraction token and returns token information
// This is a placeholder implementation - in a real system, you would validate against a secure configuration
func (t *TerraformInternalExtractionAuthMethod) validateInternalExtractionToken(token string) (*TerraformInternalExtractionTokenInfo, error) {
	// Simple validation for demonstration
	// In production, this would check against a secure database or configuration
	validTokens := map[string]*TerraformInternalExtractionTokenInfo{
		"internal-extraction-full": {
			ExtractionType: "module-discovery",
			AccessScope:    []string{"full", "*"},
		},
		"internal-extraction-read": {
			ExtractionType: "module-indexing",
			AccessScope:    []string{"modules.read", "api.read", "terraform.access"},
		},
		"internal-extraction-upload": {
			ExtractionType: "module-uploading",
			AccessScope:    []string{"modules.read", "modules.upload", "api.read"},
		},
		"internal-extraction-catalog": {
			ExtractionType: "module-cataloging",
			AccessScope:    []string{"modules.read", "modules.publish", "api.read", "terraform.access"},
		},
		"internal-extraction-migration": {
			ExtractionType: "data-migration",
			AccessScope:    []string{"full", "*"}, // Migration processes need full access
		},
	}

	tokenInfo, exists := validTokens[token]
	if !exists {
		return nil, terraformInternalExtractionErr("invalid internal extraction auth token")
	}

	return tokenInfo, nil
}

// terraformInternalExtractionErr creates a formatted error message for Terraform internal extraction authentication
func terraformInternalExtractionErr(message string) error {
	return &TerraformInternalExtractionError{Message: message}
}

// TerraformInternalExtractionError represents a Terraform internal extraction authentication error
type TerraformInternalExtractionError struct {
	Message string
}

func (e *TerraformInternalExtractionError) Error() string {
	return "Terraform Internal Extraction authentication failed: " + e.Message
}