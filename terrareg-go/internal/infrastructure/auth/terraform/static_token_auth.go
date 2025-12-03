package terraform

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/service"
)

// TerraformStaticTokenType represents the type of Terraform static token
type TerraformStaticTokenType int

const (
	// TerraformAnalyticsToken for analytics authentication
	TerraformAnalyticsToken TerraformStaticTokenType = iota
	// TerraformInternalExtractionToken for internal extraction operations
	TerraformInternalExtractionToken
	// TerraformDeploymentToken for deployment operations
	TerraformDeploymentToken
)

// TerraformStaticTokenAuth handles Terraform static token authentication
type TerraformStaticTokenAuth struct {
	config TerraformStaticTokenConfig
}

// TerraformStaticTokenConfig holds static token configuration
type TerraformStaticTokenConfig struct {
	AnalyticsTokens         []string
	InternalExtractionToken string
	DeploymentTokens        []string
}

// NewTerraformStaticTokenAuth creates a new static token authenticator
func NewTerraformStaticTokenAuth(config TerraformStaticTokenConfig) *TerraformStaticTokenAuth {
	return &TerraformStaticTokenAuth{
		config: config,
	}
}

// Authenticate validates a Terraform static token and returns authentication result
func (auth *TerraformStaticTokenAuth) Authenticate(ctx context.Context, token string) (*service.AuthResult, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}

	// Check analytics tokens
	for _, validToken := range auth.config.AnalyticsTokens {
		if token == validToken {
			return &service.AuthResult{
				UserID:         "terraform-analytics",
				Username:       "Terraform analytics token",
				Email:          "",
				DisplayName:    "Terraform analytics token",
				ExternalID:     "terraform-analytics",
				AuthProviderID: "terraform-static-token",
				Roles:          []string{"terraform-analytics"},
				Permissions:    []string{"read:analytics"},
			}, nil
		}
	}

	// Check internal extraction token
	if auth.config.InternalExtractionToken != "" && token == auth.config.InternalExtractionToken {
		return &service.AuthResult{
			UserID:         "terraform-internal-extraction",
			Username:       "Terraform internal extraction",
			Email:          "",
			DisplayName:    "Terraform internal extraction",
			ExternalID:     "terraform-internal-extraction",
			AuthProviderID: "terraform-static-token",
			Roles:          []string{"terraform-internal-extraction"},
			Permissions:    []string{"read:modules", "read:providers", "no-analytics"},
		}, nil
	}

	// Check deployment tokens
	for i, validToken := range auth.config.DeploymentTokens {
		if token == validToken {
			return &service.AuthResult{
				UserID:         fmt.Sprintf("terraform-deployment-%d", i),
				Username:       "Terraform deployment token",
				Email:          "",
				DisplayName:    "Terraform deployment token",
				ExternalID:     fmt.Sprintf("terraform-deployment-%d", i),
				AuthProviderID: "terraform-static-token",
				Roles:          []string{"terraform-deployment"},
				Permissions:    []string{"read:modules", "read:providers", "write:modules", "write:providers"},
			}, nil
		}
	}

	return nil, errors.New("invalid Terraform token")
}

// ValidateToken checks if a token is valid for a specific operation
func (auth *TerraformStaticTokenAuth) ValidateToken(ctx context.Context, token string, requiredPermissions []string) error {
	if token == "" {
		return errors.New("empty token")
	}

	// Check analytics tokens
	for _, validToken := range auth.config.AnalyticsTokens {
		if token == validToken {
			return auth.validatePermissions([]string{"read:analytics"}, requiredPermissions)
		}
	}

	// Check internal extraction token
	if auth.config.InternalExtractionToken != "" && token == auth.config.InternalExtractionToken {
		return auth.validatePermissions([]string{"read:modules", "read:providers", "no-analytics"}, requiredPermissions)
	}

	// Check deployment tokens
	for _, validToken := range auth.config.DeploymentTokens {
		if token == validToken {
			return auth.validatePermissions([]string{"read:modules", "read:providers", "write:modules", "write:providers"}, requiredPermissions)
		}
	}

	return errors.New("invalid Terraform token")
}

// GetTokenType returns the type of token
func (auth *TerraformStaticTokenAuth) GetTokenType(ctx context.Context, token string) (TerraformStaticTokenType, error) {
	if token == "" {
		return -1, errors.New("empty token")
	}

	// Check analytics tokens
	for _, validToken := range auth.config.AnalyticsTokens {
		if token == validToken {
			return TerraformAnalyticsToken, nil
		}
	}

	// Check internal extraction token
	if auth.config.InternalExtractionToken != "" && token == auth.config.InternalExtractionToken {
		return TerraformInternalExtractionToken, nil
	}

	// Check deployment tokens
	for _, validToken := range auth.config.DeploymentTokens {
		if token == validToken {
			return TerraformDeploymentToken, nil
		}
	}

	return -1, errors.New("invalid Terraform token")
}

// validatePermissions checks if the token has all required permissions
func (auth *TerraformStaticTokenAuth) validatePermissions(tokenPermissions, requiredPermissions []string) error {
	for _, required := range requiredPermissions {
		found := false
		for _, has := range tokenPermissions {
			if has == required {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("token lacks required permission: %s", required)
		}
	}
	return nil
}

// ExtractTokenFromHeader extracts token from Authorization header
func ExtractTokenFromHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	// Remove "Bearer " prefix if present
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return authHeader
}

// TerraformAuthMethod represents the type of Terraform authentication
type TerraformAuthMethod int

const (
	// TerraformOIDCMethod for OIDC-based authentication
	TerraformOIDCMethod TerraformAuthMethod = iota
	// TerraformStaticTokenMethod for static token authentication
	TerraformStaticTokenMethod
)

// TerraformAuthenticator handles all Terraform authentication methods
type TerraformAuthenticator struct {
	idp           *TerraformIDP
	staticToken   *TerraformStaticTokenAuth
	authMethods   map[TerraformAuthMethod]bool
}

// NewTerraformAuthenticator creates a new Terraform authenticator
func NewTerraformAuthenticator(idp *TerraformIDP, staticToken *TerraformStaticTokenAuth) *TerraformAuthenticator {
	return &TerraformAuthenticator{
		idp:         idp,
		staticToken: staticToken,
		authMethods: make(map[TerraformAuthMethod]bool),
	}
}

// EnableAuthMethod enables a specific authentication method
func (auth *TerraformAuthenticator) EnableAuthMethod(method TerraformAuthMethod) {
	auth.authMethods[method] = true
}

// DisableAuthMethod disables a specific authentication method
func (auth *TerraformAuthenticator) DisableAuthMethod(method TerraformAuthMethod) {
	auth.authMethods[method] = false
}

// IsMethodEnabled checks if an authentication method is enabled
func (auth *TerraformAuthenticator) IsMethodEnabled(method TerraformAuthMethod) bool {
	return auth.authMethods[method]
}

// Authenticate attempts authentication using enabled methods
func (auth *TerraformAuthenticator) Authenticate(ctx context.Context, token string, method TerraformAuthMethod) (*service.AuthResult, error) {
	if !auth.IsMethodEnabled(method) {
		return nil, fmt.Errorf("authentication method %v is not enabled", method)
	}

	switch method {
	case TerraformOIDCMethod:
		if auth.idp == nil {
			return nil, errors.New("OIDC authentication is not configured")
		}
		// Handle OIDC authentication (token should be a JWT)
		userInfo, err := auth.idp.HandleUserInfoRequest(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("OIDC authentication failed: %w", err)
		}
		return &service.AuthResult{
			UserID:         userInfo.Subject,
			Username:       userInfo.Subject,
			Email:          "",
			DisplayName:    userInfo.Subject,
			ExternalID:     userInfo.Subject,
			AuthProviderID: "terraform-oidc",
		}, nil

	case TerraformStaticTokenMethod:
		if auth.staticToken == nil {
			return nil, errors.New("static token authentication is not configured")
		}
		return auth.staticToken.Authenticate(ctx, token)

	default:
		return nil, fmt.Errorf("unknown authentication method: %v", method)
	}
}

// GetEnabledMethods returns list of enabled authentication methods
func (auth *TerraformAuthenticator) GetEnabledMethods() []TerraformAuthMethod {
	methods := make([]TerraformAuthMethod, 0)
	for method, enabled := range auth.authMethods {
		if enabled {
			methods = append(methods, method)
		}
	}
	return methods
}