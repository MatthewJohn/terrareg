package auth

import (
	"context"
)

// SamlAuthContext implements AuthContext for SAML-based authentication
// It holds the authentication state and permission logic derived from SAML attributes
type SamlAuthContext struct {
	BaseAuthContext
	nameID      string
	attributes  map[string][]string
	username    string
	email       string
	firstName   string
	lastName    string
	userGroups  []*UserGroup
	permissions map[string]string
	isAdmin     bool
}

// NewSamlAuthContext creates a new SAML auth context
func NewSamlAuthContext(ctx context.Context, nameID string, attributes map[string][]string) *SamlAuthContext {
	return &SamlAuthContext{
		BaseAuthContext: BaseAuthContext{ctx: ctx},
		nameID:          nameID,
		attributes:      attributes,
		userGroups:      make([]*UserGroup, 0),
		permissions:     make(map[string]string),
		isAdmin:         false,
	}
}

// ExtractUserDetails extracts user details from SAML attributes
func (s *SamlAuthContext) ExtractUserDetails() {
	// Extract username (common attributes)
	if usernames, exists := s.attributes["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"]; exists && len(usernames) > 0 {
		s.username = usernames[0]
	} else if usernames, exists := s.attributes["username"]; exists && len(usernames) > 0 {
		s.username = usernames[0]
	} else if usernames, exists := s.attributes["uid"]; exists && len(usernames) > 0 {
		s.username = usernames[0]
	} else {
		s.username = s.nameID
	}

	// Extract email
	if emails, exists := s.attributes["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"]; exists && len(emails) > 0 {
		s.email = emails[0]
	} else if emails, exists := s.attributes["email"]; exists && len(emails) > 0 {
		s.email = emails[0]
	} else if emails, exists := s.attributes["mail"]; exists && len(emails) > 0 {
		s.email = emails[0]
	}

	// Extract first name
	if firstNames, exists := s.attributes["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"]; exists && len(firstNames) > 0 {
		s.firstName = firstNames[0]
	} else if firstNames, exists := s.attributes["firstname"]; exists && len(firstNames) > 0 {
		s.firstName = firstNames[0]
	}

	// Extract last name
	if lastNames, exists := s.attributes["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"]; exists && len(lastNames) > 0 {
		s.lastName = lastNames[0]
	} else if lastNames, exists := s.attributes["lastname"]; exists && len(lastNames) > 0 {
		s.lastName = lastNames[0]
	}

	// Check for admin status
	if adminValues, exists := s.attributes["admin"]; exists && len(adminValues) > 0 {
		s.isAdmin = adminValues[0] == "true" || adminValues[0] == "1"
	} else if adminValues, exists := s.attributes["is_admin"]; exists && len(adminValues) > 0 {
		s.isAdmin = adminValues[0] == "true" || adminValues[0] == "1"
	}
}

// AddUserGroup adds a user group derived from SAML attributes
func (s *SamlAuthContext) AddUserGroup(group *UserGroup) {
	s.userGroups = append(s.userGroups, group)

	// Update admin status if any group has admin rights
	if group.SiteAdmin {
		s.isAdmin = true
	}
}

// SetPermission sets a namespace permission derived from SAML attributes
func (s *SamlAuthContext) SetPermission(namespace, permission string) {
	if s.permissions == nil {
		s.permissions = make(map[string]string)
	}
	s.permissions[namespace] = permission
}

// GetProviderType returns the authentication method type
func (s *SamlAuthContext) GetProviderType() AuthMethodType {
	return AuthMethodSAML
}

// GetUsername returns the username extracted from SAML attributes
func (s *SamlAuthContext) GetUsername() string {
	return s.username
}

// IsAuthenticated returns true if SAML authentication was successful
func (s *SamlAuthContext) IsAuthenticated() bool {
	return s.nameID != "" && s.username != ""
}

// IsAdmin returns true if the user is an admin based on SAML attributes
func (s *SamlAuthContext) IsAdmin() bool {
	return s.isAdmin
}

// IsBuiltInAdmin returns false for SAML-based users
func (s *SamlAuthContext) IsBuiltInAdmin() bool {
	return false
}

// IsEnabled returns true if the SAML authentication is valid
func (s *SamlAuthContext) IsEnabled() bool {
	return s.IsAuthenticated()
}

// RequiresCSRF returns true for SAML-based authentication (uses sessions)
func (s *SamlAuthContext) RequiresCSRF() bool {
	return true
}

// CheckAuthState returns true if the SAML context is in a valid state
func (s *SamlAuthContext) CheckAuthState() bool {
	return s.IsAuthenticated()
}

// CanPublishModuleVersion checks if the user can publish to a namespace
func (s *SamlAuthContext) CanPublishModuleVersion(namespace string) bool {
	if s.IsAdmin() {
		return true
	}

	// Check SAML attribute-based permissions
	if namespacePerms, exists := s.attributes["namespace_permissions"]; exists {
		for _, perm := range namespacePerms {
			if s.parseNamespacePermission(perm, namespace, []string{"FULL", "PUBLISH"}) {
				return true
			}
		}
	}

	// Check namespace permissions
	if permission, exists := s.permissions[namespace]; exists {
		return permission == "FULL" || permission == "PUBLISH"
	}

	// Check group permissions
	for _, group := range s.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// CanUploadModuleVersion checks if the user can upload to a namespace
func (s *SamlAuthContext) CanUploadModuleVersion(namespace string) bool {
	if s.IsAdmin() {
		return true
	}

	// Check SAML attribute-based permissions
	if namespacePerms, exists := s.attributes["namespace_permissions"]; exists {
		for _, perm := range namespacePerms {
			if s.parseNamespacePermission(perm, namespace, []string{"FULL", "PUBLISH", "UPLOAD"}) {
				return true
			}
		}
	}

	// Check namespace permissions
	if permission, exists := s.permissions[namespace]; exists {
		return permission == "FULL" || permission == "PUBLISH" || permission == "UPLOAD"
	}

	// Check group permissions
	for _, group := range s.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// CheckNamespaceAccess checks if the user has access to a namespace
func (s *SamlAuthContext) CheckNamespaceAccess(permissionType, namespace string) bool {
	if s.IsAdmin() {
		return true
	}

	// Check SAML attribute-based permissions
	if namespacePerms, exists := s.attributes["namespace_permissions"]; exists {
		for _, perm := range namespacePerms {
			if s.parseNamespacePermission(perm, namespace, []string{permissionType}) {
				return true
			}
		}
	}

	// Check namespace permissions
	storedPermission, exists := s.permissions[namespace]
	if exists {
		return s.hasPermissionHierarchy(storedPermission, permissionType)
	}

	// Check group permissions
	for _, group := range s.userGroups {
		if group.SiteAdmin {
			return true
		}
	}

	return false
}

// GetAllNamespacePermissions returns all namespace permissions for the user
func (s *SamlAuthContext) GetAllNamespacePermissions() map[string]string {
	result := make(map[string]string)

	// Add direct namespace permissions
	for k, v := range s.permissions {
		result[k] = v
	}

	// Add permissions from SAML attributes
	if namespacePerms, exists := s.attributes["namespace_permissions"]; exists {
		for _, perm := range namespacePerms {
			if namespace, permission := s.parseNamespacePermissionFull(perm); namespace != "" && permission != "" {
				result[namespace] = permission
			}
		}
	}

	// Add group permissions
	for _, group := range s.userGroups {
		if group.SiteAdmin {
			// Site admins get full access to all namespaces
			result["*"] = "FULL"
			break
		}
	}

	return result
}

// GetUserGroupNames returns the names of all user groups (including from SAML attributes)
func (s *SamlAuthContext) GetUserGroupNames() []string {
	names := make([]string, len(s.userGroups))
	for i, group := range s.userGroups {
		names[i] = group.Name
	}

	// Add groups from SAML attributes
	if groups, exists := s.attributes["groups"]; exists {
		names = append(names, groups...)
	} else if groups, exists := s.attributes["member"]; exists {
		names = append(names, groups...)
	}

	return names
}

// CanAccessReadAPI returns true if the user can access the read API
func (s *SamlAuthContext) CanAccessReadAPI() bool {
	return s.IsAuthenticated()
}

// CanAccessTerraformAPI returns true if the user can access the Terraform API
func (s *SamlAuthContext) CanAccessTerraformAPI() bool {
	return s.IsAuthenticated()
}

// GetTerraformAuthToken returns empty string for SAML-based auth
func (s *SamlAuthContext) GetTerraformAuthToken() string {
	return ""
}

// GetProviderData returns provider-specific data for the SAML authentication
func (s *SamlAuthContext) GetProviderData() map[string]interface{} {
	return map[string]interface{}{
		"name_id":     s.nameID,
		"username":    s.username,
		"email":       s.email,
		"first_name":  s.firstName,
		"last_name":   s.lastName,
		"attributes":  s.attributes,
		"is_admin":    s.isAdmin,
		"auth_method": string(AuthMethodSAML),
	}
}

// parseNamespacePermission parses a namespace permission string from SAML attributes
func (s *SamlAuthContext) parseNamespacePermission(perm, namespace string, allowedPermissions []string) bool {
	// Format: "namespace:permission" or "namespace:*"
	for i := 1; i < len(perm); i++ {
		if perm[i-1] == ':' {
			parts := []string{perm[:i-1], perm[i:]}

			if len(parts) != 2 {
				return false
			}

			if parts[0] != namespace && parts[0] != "*" {
				return false
			}

			if parts[1] == "*" {
				return true
			}

			for _, allowed := range allowedPermissions {
				if parts[1] == allowed {
					return true
				}
			}

			return false
		}
	}

	return false
}

// parseNamespacePermissionFull parses a namespace permission string and returns both parts
func (s *SamlAuthContext) parseNamespacePermissionFull(perm string) (string, string) {
	for i := 1; i < len(perm); i++ {
		if perm[i-1] == ':' {
			parts := []string{perm[:i-1], perm[i:]}

			if len(parts) != 2 {
				return "", ""
			}

			return parts[0], parts[1]
		}
	}

	return "", ""
}

// hasPermissionHierarchy checks if the stored permission meets or exceeds the required permission
func (s *SamlAuthContext) hasPermissionHierarchy(stored, required string) bool {
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
func (s *SamlAuthContext) comparePermissionLevel(a, b string) int {
	levelA := s.getPermissionLevel(a)
	levelB := s.getPermissionLevel(b)

	if levelA > levelB {
		return 1
	} else if levelA < levelB {
		return -1
	}
	return 0
}

// getPermissionLevel returns a numeric value for permission comparison
func (s *SamlAuthContext) getPermissionLevel(permission string) int {
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
