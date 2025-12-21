package auth

import (
	"context"
)

// OpenidConnectAuthContext implements AuthContext for OpenID Connect-based authentication
// It holds the authentication state and permission logic derived from OIDC claims
type OpenidConnectAuthContext struct {
	BaseAuthContext
	sub         string
	claims      map[string]interface{}
	username    string
	email       string
	name        string
	picture     string
	userGroups  []*UserGroup
	permissions map[string]string
	isAdmin     bool
}

// NewOpenidConnectAuthContext creates a new OpenID Connect auth context
func NewOpenidConnectAuthContext(ctx context.Context, sub string, claims map[string]interface{}) *OpenidConnectAuthContext {
	return &OpenidConnectAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		sub:            sub,
		claims:         claims,
		userGroups:     make([]*UserGroup, 0),
		permissions:    make(map[string]string),
		isAdmin:        false,
	}
}

// ExtractUserDetails extracts user details from OIDC claims
func (o *OpenidConnectAuthContext) ExtractUserDetails() {
	// Extract username (preferred claim order)
	if username, ok := o.claims["preferred_username"].(string); ok && username != "" {
		o.username = username
	} else if username, ok := o.claims["username"].(string); ok && username != "" {
		o.username = username
	} else if username, ok := o.claims["sub"].(string); ok && username != "" {
		o.username = username
	} else {
		o.username = o.sub
	}

	// Extract email
	if email, ok := o.claims["email"].(string); ok {
		o.email = email
	}

	// Extract name
	if name, ok := o.claims["name"].(string); ok {
		o.name = name
	}

	// Extract picture
	if picture, ok := o.claims["picture"].(string); ok {
		o.picture = picture
	}

	// Check for admin status
	if isAdmin, ok := o.claims["admin"].(bool); ok && isAdmin {
		o.isAdmin = true
	} else if isAdminStr, ok := o.claims["admin"].(string); ok && (isAdminStr == "true" || isAdminStr == "1") {
		o.isAdmin = true
	}

	// Check groups claim for admin status
	if groups, ok := o.claims["groups"].([]interface{}); ok {
		for _, group := range groups {
			if groupStr, ok := group.(string); ok && (groupStr == "admin" || groupStr == "administrators") {
				o.isAdmin = true
				break
			}
		}
	}
}

// AddUserGroup adds a user group derived from OIDC claims
func (o *OpenidConnectAuthContext) AddUserGroup(group *UserGroup) {
	o.userGroups = append(o.userGroups, group)

	// Update admin status if any group has admin rights
	if group.SiteAdmin {
		o.isAdmin = true
	}
}

// SetPermission sets a namespace permission derived from OIDC claims
func (o *OpenidConnectAuthContext) SetPermission(namespace, permission string) {
	if o.permissions == nil {
		o.permissions = make(map[string]string)
	}
	o.permissions[namespace] = permission
}

// GetProviderType returns the authentication method type
func (o *OpenidConnectAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodOpenIDConnect
}

// GetUsername returns the username extracted from OIDC claims
func (o *OpenidConnectAuthContext) GetUsername() string {
	return o.username
}

// IsAuthenticated returns true if OIDC authentication was successful
func (o *OpenidConnectAuthContext) IsAuthenticated() bool {
	return o.sub != "" && o.username != ""
}

// IsAdmin returns true if the user is an admin based on OIDC claims
func (o *OpenidConnectAuthContext) IsAdmin() bool {
	return o.isAdmin
}

// IsBuiltInAdmin returns false for OpenID Connect-based users
func (o *OpenidConnectAuthContext) IsBuiltInAdmin() bool {
	return false
}

// IsEnabled returns true if the OpenID Connect authentication is valid
func (o *OpenidConnectAuthContext) IsEnabled() bool {
	return o.IsAuthenticated()
}

// RequiresCSRF returns true for OpenID Connect-based authentication (uses sessions)
func (o *OpenidConnectAuthContext) RequiresCSRF() bool {
	return true
}

// CheckAuthState returns true if the OIDC context is in a valid state
func (o *OpenidConnectAuthContext) CheckAuthState() bool {
	return o.IsAuthenticated()
}

// CanPublishModuleVersion checks if the user can publish to a namespace
func (o *OpenidConnectAuthContext) CanPublishModuleVersion(namespace string) bool {
	if o.IsAdmin() {
		return true
	}

	// Check OIDC claim-based permissions
	if namespacePerms, ok := o.claims["namespace_permissions"].(map[string]interface{}); ok {
		if permission, exists := namespacePerms[namespace]; exists {
			if permStr, ok := permission.(string); ok {
				return permStr == "FULL" || permStr == "PUBLISH"
			}
		}
	}

	// Check namespace permissions
	if permission, exists := o.permissions[namespace]; exists {
		return permission == "FULL" || permission == "PUBLISH"
	}

	// Check group permissions
	for _, group := range o.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// CanUploadModuleVersion checks if the user can upload to a namespace
func (o *OpenidConnectAuthContext) CanUploadModuleVersion(namespace string) bool {
	if o.IsAdmin() {
		return true
	}

	// Check OIDC claim-based permissions
	if namespacePerms, ok := o.claims["namespace_permissions"].(map[string]interface{}); ok {
		if permission, exists := namespacePerms[namespace]; exists {
			if permStr, ok := permission.(string); ok {
				return permStr == "FULL" || permStr == "PUBLISH" || permStr == "UPLOAD"
			}
		}
	}

	// Check namespace permissions
	if permission, exists := o.permissions[namespace]; exists {
		return permission == "FULL" || permission == "PUBLISH" || permission == "UPLOAD"
	}

	// Check group permissions
	for _, group := range o.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// CheckNamespaceAccess checks if the user has access to a namespace
func (o *OpenidConnectAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	if o.IsAdmin() {
		return true
	}

	// Check OIDC claim-based permissions
	if namespacePerms, ok := o.claims["namespace_permissions"].(map[string]interface{}); ok {
		if permission, exists := namespacePerms[namespace]; exists {
			if permStr, ok := permission.(string); ok {
				return o.hasPermissionHierarchy(permStr, permissionType)
			}
		}
	}

	// Check namespace permissions
	storedPermission, exists := o.permissions[namespace]
	if exists {
		return o.hasPermissionHierarchy(storedPermission, permissionType)
	}

	// Check group permissions
	for _, group := range o.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (o *OpenidConnectAuthContext) GetAllNamespacePermissions() map[string]string {
	result := make(map[string]string)

	// Add direct namespace permissions
	for k, v := range o.permissions {
		result[k] = v
	}

	// Add permissions from OIDC claims
	if namespacePerms, ok := o.claims["namespace_permissions"].(map[string]interface{}); ok {
		for k, v := range namespacePerms {
			if permStr, ok := v.(string); ok {
				result[k] = permStr
			}
		}
	}

	// Add group permissions
	for _, group := range o.userGroups {
		if group.SiteAdmin {
			// Site admins get full access to all namespaces
			result["*"] = "FULL"
			break
		}
	}

	return result
}

// GetUserGroupNames returns the names of all user groups (including from OIDC claims)
func (o *OpenidConnectAuthContext) GetUserGroupNames() []string {
	names := make([]string, len(o.userGroups))
	for i, group := range o.userGroups {
		names[i] = group.Name
	}

	// Add groups from OIDC claims
	if groups, ok := o.claims["groups"].([]interface{}); ok {
		for _, group := range groups {
			if groupStr, ok := group.(string); ok {
				names = append(names, groupStr)
			}
		}
	} else if groups, ok := o.claims["groups"].(string); ok {
		// Single group as string
		names = append(names, groups)
	}

	return names
}

// CanAccessReadAPI returns true if the user can access the read API
func (o *OpenidConnectAuthContext) CanAccessReadAPI() bool {
	return o.IsAuthenticated()
}

// CanAccessTerraformAPI returns true if the user can access the Terraform API
func (o *OpenidConnectAuthContext) CanAccessTerraformAPI() bool {
	return o.IsAuthenticated()
}

// GetTerraformAuthToken returns empty string for OpenID Connect-based auth
func (o *OpenidConnectAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns provider-specific data for the OpenID Connect authentication
func (o *OpenidConnectAuthContext) GetProviderData() map[string]interface{} {
	data := map[string]interface{}{
		"sub":       o.sub,
		"username":  o.username,
		"email":     o.email,
		"name":      o.name,
		"is_admin":  o.isAdmin,
		"auth_method": string(AuthMethodOpenIDConnect),
	}

	if o.picture != "" {
		data["picture"] = o.picture
	}

	// Add all claims (excluding sensitive ones)
	for k, v := range o.claims {
		// Skip sensitive claims
		if k != "access_token" && k != "refresh_token" && k != "id_token" {
			data[k] = v
		}
	}

	return data
}

// hasPermissionHierarchy checks if the stored permission meets or exceeds the required permission
func (o *OpenidConnectAuthContext) hasPermissionHierarchy(stored, required string) bool {
	switch required {
	case "READ":
		return stored == "READ" || stored == "MODIFY" || stored == "FULL" || stored == "UPLOAD" || stored == "PUBLISH"
	case "MODIFY":
		return stored == "MODIFY" || stored == "FULL" || stored == "UPLOAD" || stored == "PUBLISH"
	case "UPLOAD":
		return stored == "UPLOAD" || stored == "FULL" || stored == "PUBLISH"
	case "PUBLISH":
		return stored == "PUBLISH" || stored == "FULL"
	case "FULL":
		return stored == "FULL"
	default:
		return false
	}
}

// comparePermissionLevel compares two permission levels, returns 1 if a > b, -1 if a < b, 0 if equal
func (o *OpenidConnectAuthContext) comparePermissionLevel(a, b string) int {
	levelA := o.getPermissionLevel(a)
	levelB := o.getPermissionLevel(b)

	if levelA > levelB {
		return 1
	} else if levelA < levelB {
		return -1
	}
	return 0
}

// getPermissionLevel returns a numeric value for permission comparison
func (o *OpenidConnectAuthContext) getPermissionLevel(permission string) int {
	switch permission {
	case "READ":
		return 1
	case "MODIFY":
		return 2
	case "UPLOAD":
		return 3
	case "PUBLISH":
		return 4
	case "FULL":
		return 5
	default:
		return 0
	}
}