package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
)

// OpenIDConnectAuthMethod implements authentication for OpenID Connect (OIDC)
type OpenIDConnectAuthMethod struct {
	auth.BaseAuthMethod
	oidcConfig     *OIDCConfig
	userGroupRepo  repository.UserGroupRepository
	httpClient     *http.Client
	isAuthenticated bool
	isAdmin         bool
	username        string
	email           string
	subject         string
	userPermissions map[string]string
	userGroups      []*auth.UserGroup
}

// OIDCConfig represents OpenID Connect configuration
type OIDCConfig struct {
	IssuerURL      string
	ClientID       string
	ClientSecret   string
	RedirectURI    string
	Scopes         []string
	ResponseType   string
	GrantType      string
	TokenEndpoint  string
	UserInfoEndpoint string
	JWKSURL        string
	AuthorizationEndpoint string
	ClaimMapping   map[string]string
	GroupClaim     string
	AdminGroups    []string
}

// OIDCTokenResponse represents OIDC token response
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope"`
}

// OIDCUserInfo represents OIDC user information
type OIDCUserInfo struct {
	Sub               string            `json:"sub"`
	Name              string            `json:"name"`
	Email             string            `json:"email"`
	EmailVerified     bool              `json:"email_verified"`
	PreferredUsername string            `json:"preferred_username"`
	GivenName         string            `json:"given_name"`
	FamilyName        string            `json:"family_name"`
	Groups            []string          `json:"groups,omitempty"`
	RawClaims         map[string]interface{} `json:"-"`
}

// OIDCDiscoveryDocument represents OIDC discovery document
type OIDCDiscoveryDocument struct {
	Issuer                           string   `json:"issuer"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	TokenEndpoint                    string   `json:"token_endpoint"`
	UserinfoEndpoint                 string   `json:"userinfo_endpoint"`
	JwksUri                          string   `json:"jwks_uri"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	GrantTypesSupported              []string `json:"grant_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                  []string `json:"scopes_supported"`
}

// NewOpenIDConnectAuthMethod creates a new OpenID Connect authentication method
func NewOpenIDConnectAuthMethod(oidcConfig *OIDCConfig, userGroupRepo repository.UserGroupRepository) *OpenIDConnectAuthMethod {
	return &OpenIDConnectAuthMethod{
		oidcConfig:      oidcConfig,
		userGroupRepo:   userGroupRepo,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
		userPermissions: make(map[string]string),
		userGroups:      make([]*auth.UserGroup, 0),
	}
}

// GetProviderType returns the authentication provider type
func (o *OpenIDConnectAuthMethod) GetProviderType() auth.AuthMethodType {
	return auth.AuthMethodOpenIDConnect
}

// CheckAuthState validates the current authentication state
func (o *OpenIDConnectAuthMethod) CheckAuthState() bool {
	return o.isAuthenticated
}

// IsBuiltInAdmin returns whether this is a built-in admin method
func (o *OpenIDConnectAuthMethod) IsBuiltInAdmin() bool {
	return false // OpenID Connect is not built-in admin
}

// IsAuthenticated returns whether the current request is authenticated
func (o *OpenIDConnectAuthMethod) IsAuthenticated() bool {
	return o.isAuthenticated
}

// IsAdmin returns whether the authenticated user has admin privileges
func (o *OpenIDConnectAuthMethod) IsAdmin() bool {
	return o.isAdmin
}

// IsEnabled returns whether this authentication method is enabled
func (o *OpenIDConnectAuthMethod) IsEnabled() bool {
	return o.oidcConfig != nil
}

// RequiresCSRF returns whether this authentication method requires CSRF protection
func (o *OpenIDConnectAuthMethod) RequiresCSRF() bool {
	return true // OpenID Connect requires CSRF protection
}

// CanPublishModuleVersion checks if the user can publish module versions to the given namespace
func (o *OpenIDConnectAuthMethod) CanPublishModuleVersion(namespace string) bool {
	if o.isAdmin {
		return true
	}
	return o.CheckNamespaceAccess("FULL", namespace)
}

// CanUploadModuleVersion checks if the user can upload module versions to the given namespace
func (o *OpenIDConnectAuthMethod) CanUploadModuleVersion(namespace string) bool {
	if o.isAdmin {
		return true
	}
	return o.CheckNamespaceAccess("FULL", namespace) || o.CheckNamespaceAccess("MODIFY", namespace)
}

// CheckNamespaceAccess checks if the user has the specified permission for a namespace
func (o *OpenIDConnectAuthMethod) CheckNamespaceAccess(permissionType, namespace string) bool {
	if o.isAdmin {
		return true
	}

	storedPermission, exists := o.userPermissions[namespace]
	if !exists {
		return false
	}

	// Check permission hierarchy
	switch auth.PermissionType(permissionType) {
	case auth.PermissionRead:
		return storedPermission == string(auth.PermissionRead) ||
			storedPermission == string(auth.PermissionModify) ||
			storedPermission == string(auth.PermissionFull)
	case auth.PermissionModify:
		return storedPermission == string(auth.PermissionModify) ||
			storedPermission == string(auth.PermissionFull)
	case auth.PermissionFull:
		return storedPermission == string(auth.PermissionFull)
	default:
		return false
	}
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (o *OpenIDConnectAuthMethod) GetAllNamespacePermissions() map[string]string {
	return o.userPermissions
}

// GetUsername returns the authenticated username
func (o *OpenIDConnectAuthMethod) GetUsername() string {
	return o.username
}

// GetUserGroupNames returns the names of all user groups
func (o *OpenIDConnectAuthMethod) GetUserGroupNames() []string {
	names := make([]string, len(o.userGroups))
	for i, group := range o.userGroups {
		names[i] = group.GetName()
	}
	return names
}

// CanAccessReadAPI returns whether the user can access read APIs
func (o *OpenIDConnectAuthMethod) CanAccessReadAPI() bool {
	return o.isAuthenticated
}

// CanAccessTerraformAPI returns whether the user can access Terraform APIs
func (o *OpenIDConnectAuthMethod) CanAccessTerraformAPI() bool {
	return o.isAdmin
}

// GetTerraformAuthToken returns the Terraform authentication token
func (o *OpenIDConnectAuthMethod) GetTerraformAuthToken() string {
	// OpenID Connect doesn't typically provide Terraform tokens
	return ""
}

// GetProviderData returns provider-specific data
func (o *OpenIDConnectAuthMethod) GetProviderData() map[string]interface{} {
	data := make(map[string]interface{})
	data["username"] = o.username
	data["email"] = o.email
	data["subject"] = o.subject
	data["groups"] = o.GetUserGroupNames()
	return data
}

// Authenticate authenticates a request using OpenID Connect
func (o *OpenIDConnectAuthMethod) Authenticate(ctx context.Context, headers map[string]string, cookies map[string]string) error {
	// Look for authorization code in headers or cookies
	authCode, exists := headers["X-OIDC-Auth-Code"]
	if !exists {
		// Try to get from session cookies
		sessionID, exists := cookies["session_id"]
		if !exists {
			return oidcErr("missing OIDC authorization code")
		}

		// In a real implementation, you would look up the session and extract OIDC data
		// For now, we'll simulate OIDC authentication with a placeholder
		return o.simulateOIDCAuthentication(ctx)
	}

	// Exchange authorization code for tokens
	tokenResponse, err := o.exchangeCodeForTokens(ctx, authCode)
	if err != nil {
		return err
	}

	// Get user info from OIDC provider
	userInfo, err := o.getUserInfo(ctx, tokenResponse.AccessToken)
	if err != nil {
		return err
	}

	// Extract and validate ID token to get subject
	subject, err := o.extractSubjectFromIDToken(tokenResponse.IDToken)
	if err != nil {
		return err
	}

	o.username = userInfo.Name
	if o.username == "" {
		o.username = userInfo.PreferredUsername
	}
	if o.username == "" {
		o.username = userInfo.Email
	}
	o.email = userInfo.Email
	o.subject = subject
	o.isAuthenticated = true

	// Process user groups from OIDC claims
	err = o.processUserGroups(ctx, userInfo.Groups)
	if err != nil {
		return err
	}

	return nil
}

// exchangeCodeForTokens exchanges authorization code for access tokens
func (o *OpenIDConnectAuthMethod) exchangeCodeForTokens(ctx context.Context, authCode string) (*OIDCTokenResponse, error) {
	// This is a placeholder implementation
	// In production, you would make an HTTP POST request to the token endpoint
	// with client authentication and the authorization code

	// Mock token response for demonstration
	return &OIDCTokenResponse{
		AccessToken:  "mock-access-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "mock-refresh-token",
		IDToken:      "mock-id-token",
		Scope:        "openid profile email groups",
	}, nil
}

// getUserInfo retrieves user information from OIDC userinfo endpoint
func (o *OpenIDConnectAuthMethod) getUserInfo(ctx context.Context, accessToken string) (*OIDCUserInfo, error) {
	// This is a placeholder implementation
	// In production, you would make an HTTP GET request to the userinfo endpoint
	// with the access token in the Authorization header

	// Mock user info for demonstration
	return &OIDCUserInfo{
		Sub:               "1234567890",
		Name:              "OIDC User",
		Email:             "oidc-user@example.com",
		EmailVerified:     true,
		PreferredUsername: "oidc-user",
		Groups:            []string{"developers", "terraform-users", "oidc-admins"},
	}, nil
}

// extractSubjectFromIDToken extracts subject claim from ID token
func (o *OpenIDConnectAuthMethod) extractSubjectFromIDToken(idToken string) (string, error) {
	// This is a placeholder implementation
	// In production, you would:
	// 1. Parse the JWT ID token
	// 2. Validate the signature using the JWKS endpoint
	// 3. Validate claims (iss, aud, exp, nbf, iat)
	// 4. Extract the 'sub' claim

	// For demonstration, return mock subject
	return "1234567890", nil
}

// simulateOIDCAuthentication simulates OIDC authentication for testing
func (o *OpenIDConnectAuthMethod) simulateOIDCAuthentication(ctx context.Context) error {
	// Mock OIDC user data for demonstration
	o.username = "oidc-user@example.com"
	o.email = "oidc-user@example.com"
	o.subject = "1234567890"
	o.isAuthenticated = true

	// Mock groups for demonstration
	mockGroups := []string{"developers", "terraform-users", "oidc-admins"}
	err := o.processUserGroups(ctx, mockGroups)
	if err != nil {
		return err
	}

	return nil
}

// processUserGroups processes OIDC groups and maps to local user groups
func (o *OpenIDConnectAuthMethod) processUserGroups(ctx context.Context, oidcGroups []string) error {
	// This is a placeholder implementation
	// In production, you would map OIDC groups to local user groups
	// using a mapping configuration or by matching group names

	// Check if user belongs to any admin groups
	for _, adminGroup := range o.oidcConfig.AdminGroups {
		for _, userGroup := range oidcGroups {
			if userGroup == adminGroup {
				o.isAdmin = true
				break
			}
		}
		if o.isAdmin {
			break
		}
	}

	// Create mock user groups for demonstration
	for _, groupName := range oidcGroups {
		userGroup := &auth.UserGroup{
			ID:          len(o.userGroups) + 1,
			Name:        groupName,
			SiteAdmin:   o.isAdmin,
			Description: fmt.Sprintf("OIDC group: %s", groupName),
		}
		o.userGroups = append(o.userGroups, userGroup)
	}

	// Get user permissions for groups
	o.userPermissions = o.getUserPermissions(ctx)

	return nil
}

// getUserPermissions gets the user's permissions across all namespaces
func (o *OpenIDConnectAuthMethod) getUserPermissions(ctx context.Context) map[string]string {
	permissions := make(map[string]string)

	if o.isAdmin {
		// Admin users get access to all namespaces - return empty map to signify admin
		return permissions
	}

	// Get namespace permissions for each group
	for _, group := range o.userGroups {
		// This is a placeholder - in production, you would query the repository
		// for group permissions and apply them
		groupPermissions, err := o.userGroupRepo.GetNamespacePermissions(ctx, group.GetID())
		if err != nil {
			continue
		}

		for _, perm := range groupPermissions {
			namespaceName := o.getNamespaceName(perm.GetNamespaceID())
			if namespaceName == "" {
				continue
			}

			// Use the highest permission level if multiple permissions exist
			current, exists := permissions[namespaceName]
			permType := string(perm.GetPermissionType())
			if !exists || o.isHigherPermission(permType, current) {
				permissions[namespaceName] = permType
			}
		}
	}

	return permissions
}

// getNamespaceName would get the namespace name from ID
// This is a placeholder - in a real implementation, you'd query the namespace repository
func (o *OpenIDConnectAuthMethod) getNamespaceName(namespaceID int) string {
	// Placeholder implementation
	return ""
}

// isHigherPermission checks if permission1 is higher level than permission2
func (o *OpenIDConnectAuthMethod) isHigherPermission(perm1, perm2 string) bool {
	permLevels := map[string]int{
		"READ":   1,
		"MODIFY": 2,
		"FULL":   3,
	}

	level1, exists1 := permLevels[perm1]
	level2, exists2 := permLevels[perm2]

	if !exists1 {
		return false
	}
	if !exists2 {
		return true
	}

	return level1 > level2
}

// DiscoverOIDCEndpoints discovers OIDC endpoints from issuer URL
func (o *OpenIDConnectAuthMethod) DiscoverOIDCEndpoints(ctx context.Context) (*OIDCDiscoveryDocument, error) {
	if o.oidcConfig.IssuerURL == "" {
		return nil, oidcErr("issuer URL not configured")
	}

	discoveryURL := fmt.Sprintf("%s/.well-known/openid-configuration", o.oidcConfig.IssuerURL)

	req, err := http.NewRequestWithContext(ctx, "GET", discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery endpoint returned status: %d", resp.StatusCode)
	}

	var discoveryDoc OIDCDiscoveryDocument
	err = json.NewDecoder(resp.Body).Decode(&discoveryDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to decode discovery document: %w", err)
	}

	return &discoveryDoc, nil
}

// oidcErr creates a formatted error message for OIDC authentication
func oidcErr(message string) error {
	return &OIDCError{Message: message}
}

// OIDCError represents an OpenID Connect authentication error
type OIDCError struct {
	Message string
}

func (e *OIDCError) Error() string {
	return "OpenID Connect authentication failed: " + e.Message
}