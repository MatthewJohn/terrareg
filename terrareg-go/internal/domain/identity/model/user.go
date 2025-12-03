package model

import (
	"fmt"
	"regexp"
	"time"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// User represents a user in the system
type User struct {
	id                   string
	username             string
	displayName          string
	email                string
	authMethod           AuthMethod
	authProviderID       string
	externalID           string
	accessToken          string
	refreshToken         string
	accessTokenExpiry    *time.Time
	active               bool
	createdAt            time.Time
	lastLoginAt          *time.Time
	permissions          []Permission
}

// AuthMethod represents the authentication method used
type AuthMethod int

const (
	AuthMethodNone AuthMethod = iota
	AuthMethodSAML
	AuthMethodOIDC
	AuthMethodGitHub
	AuthMethodAPIKey
	AuthMethodTerraform
)

func (am AuthMethod) String() string {
	switch am {
	case AuthMethodSAML:
		return "SAML"
	case AuthMethodOIDC:
		return "OIDC"
	case AuthMethodGitHub:
		return "GITHUB"
	case AuthMethodAPIKey:
		return "API_KEY"
	case AuthMethodTerraform:
		return "TERRAFORM"
	default:
		return "NONE"
	}
}

// Permission represents a user's permission
type Permission struct {
	id           int
	resourceType ResourceType
	resourceID   string
	action       Action
	grantedAt    time.Time
	grantedBy    *User
}

// ResourceType represents the type of resource
type ResourceType string

const (
	ResourceTypeNamespace ResourceType = "namespace"
	ResourceTypeModule    ResourceType = "module"
	ResourceTypeProvider  ResourceType = "provider"
)

// Action represents the allowed action
type Action string

const (
	ActionRead   Action = "read"
	ActionWrite  Action = "write"
	ActionAdmin  Action = "admin"
)

// NewUser creates a new user
func NewUser(username, displayName, email string, authMethod AuthMethod) (*User, error) {
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	return &User{
		id:          shared.GenerateID(),
		username:    username,
		displayName: displayName,
		email:       email,
		authMethod:  authMethod,
		active:      true,
		createdAt:   time.Now(),
		permissions: make([]Permission, 0),
	}, nil
}

// Permission constants for Terraform operations
var (
	PermissionReadModules     = Permission{resourceType: ResourceTypeModule, resourceID: "*", action: ActionRead}
	PermissionReadProviders   = Permission{resourceType: ResourceTypeProvider, resourceID: "*", action: ActionRead}
	PermissionPublishModules  = Permission{resourceType: ResourceTypeModule, resourceID: "*", action: ActionWrite}
	PermissionPublishProviders = Permission{resourceType: ResourceTypeProvider, resourceID: "*", action: ActionWrite}
	PermissionReadAnalytics   = Permission{resourceType: ResourceTypeNamespace, resourceID: "*", action: ActionRead}
)

// Business methods

// Authenticate updates user's authentication tokens
func (u *User) Authenticate(accessToken, refreshToken string, expiry *time.Time) error {
	if accessToken == "" {
		return ErrAccessTokenRequired
	}

	u.accessToken = accessToken
	u.refreshToken = refreshToken
	u.accessTokenExpiry = expiry
	now := time.Now()
	u.lastLoginAt = &now

	return nil
}

// IsTokenExpired checks if the access token is expired
func (u *User) IsTokenExpired() bool {
	if u.accessTokenExpiry == nil {
		return false
	}
	return time.Now().After(*u.accessTokenExpiry)
}

// AddPermission adds a permission to the user
func (u *User) AddPermission(resourceType ResourceType, resourceID string, action Action, grantedBy *User) error {
	// Check if permission already exists
	for _, perm := range u.permissions {
		if perm.resourceType == resourceType && perm.resourceID == resourceID && perm.action == action {
			return ErrPermissionAlreadyExists
		}
	}

	permission := Permission{
		id:           shared.GenerateIntID(),
		resourceType: resourceType,
		resourceID:   resourceID,
		action:       action,
		grantedAt:    time.Now(),
		grantedBy:    grantedBy,
	}

	u.permissions = append(u.permissions, permission)
	return nil
}

// AddPermissionSimple adds a permission to the user without requiring grantedBy (for service layer)
func (u *User) AddPermissionSimple(permission Permission) {
	// Check if permission already exists
	for _, perm := range u.permissions {
		if perm.resourceType == permission.resourceType &&
		   perm.resourceID == permission.resourceID &&
		   perm.action == permission.action {
			return // Already exists
		}
	}

	// Create new permission with generated ID
	newPermission := Permission{
		id:           shared.GenerateIntID(),
		resourceType: permission.resourceType,
		resourceID:   permission.resourceID,
		action:       permission.action,
		grantedAt:    time.Now(),
		grantedBy:    nil, // No explicit grantor for service-added permissions
	}

	u.permissions = append(u.permissions, newPermission)
}

// HasPermission checks if user has a specific permission
func (u *User) HasPermission(resourceType ResourceType, resourceID string, action Action) bool {
	for _, perm := range u.permissions {
		if perm.resourceType == resourceType && perm.resourceID == resourceID {
			if perm.action == action || perm.action == ActionAdmin {
				return true
			}
		}
	}
	return false
}

// Deactivate deactivates the user
func (u *User) Deactivate() error {
	if !u.active {
		return ErrUserAlreadyDeactivated
	}
	u.active = false
	return nil
}

// UpdateProfile updates user's profile information
func (u *User) UpdateProfile(displayName, email string) error {
	if email != "" {
		if err := ValidateEmail(email); err != nil {
			return err
		}
		u.email = email
	}
	if displayName != "" {
		u.displayName = displayName
	}
	return nil
}

// SetExternalID sets the external ID for the user
func (u *User) SetExternalID(externalID string) {
	u.externalID = externalID
}

// SetAuthProviderID sets the auth provider ID for the user
func (u *User) SetAuthProviderID(authProviderID string) {
	u.authProviderID = authProviderID
}

// SetIPAddress sets the IP address for the user (for logging)
func (u *User) SetIPAddress(ipAddress string) {
	// This would typically be stored in a separate audit log
}

// SetUserAgent sets the user agent for the user (for logging)
func (u *User) SetUserAgent(userAgent string) {
	// This would typically be stored in a separate audit log
}

// Getters
func (u *User) ID() string                { return u.id }
func (u *User) Username() string          { return u.username }
func (u *User) DisplayName() string        { return u.displayName }
func (u *User) Email() string              { return u.email }
func (u *User) AuthMethod() AuthMethod     { return u.authMethod }
func (u *User) AuthProviderID() string    { return u.authProviderID }
func (u *User) ExternalID() string         { return u.externalID }
func (u *User) AccessToken() string        { return u.accessToken }
func (u *User) RefreshToken() string       { return u.refreshToken }
func (u *User) AccessTokenExpiry() *time.Time { return u.accessTokenExpiry }
func (u *User) Active() bool              { return u.active }
func (u *User) CreatedAt() time.Time       { return u.createdAt }
func (u *User) LastLoginAt() *time.Time   { return u.lastLoginAt }
func (u *User) Permissions() []Permission { return u.permissions }

// Permission getters
func (p Permission) ID() int           { return p.id }
func (p Permission) ResourceType() ResourceType { return p.resourceType }
func (p Permission) ResourceID() string { return p.resourceID }
func (p Permission) Action() Action     { return p.action }
func (p Permission) GrantedAt() time.Time { return p.grantedAt }
func (p Permission) GrantedBy() *User   { return p.grantedBy }

// ReconstructUser creates a user from database data (for repository use)
func ReconstructUser(
	id, username, displayName, email string,
	authMethod AuthMethod,
	authProviderID, externalID, accessToken, refreshToken string,
	tokenExpiry, lastLoginAt *time.Time,
	createdAt time.Time,
	active bool,
) (*User, error) {
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}
	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	return &User{
		id:                   id,
		username:             username,
		displayName:          displayName,
		email:                email,
		authMethod:           authMethod,
		authProviderID:       authProviderID,
		externalID:           externalID,
		accessToken:          accessToken,
		refreshToken:         refreshToken,
		accessTokenExpiry:    tokenExpiry,
		active:               active,
		createdAt:            createdAt,
		lastLoginAt:          lastLoginAt,
		permissions:          make([]Permission, 0),
	}, nil
}

// SetAccessToken sets the access token and optionally expiry
func (u *User) SetAccessToken(accessToken string) {
	u.accessToken = accessToken
}

// AddPermissionFromService adds a permission using the service layer pattern
func (u *User) AddPermissionFromService(permission Permission) {
	u.permissions = append(u.permissions, permission)
}