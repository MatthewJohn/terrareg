package model

import (
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// TerraformIdentityType represents the type of Terraform identity
type TerraformIdentityType int

const (
	// TerraformIdentityTypeOIDC for OIDC-based authentication
	TerraformIdentityTypeOIDC TerraformIdentityType = iota
	// TerraformIdentityTypeAnalytics for analytics tokens
	TerraformIdentityTypeAnalytics
	// TerraformIdentityTypeInternalExtraction for internal operations
	TerraformIdentityTypeInternalExtraction
	// TerraformIdentityTypeDeployment for deployment operations
	TerraformIdentityTypeDeployment
)

// TerraformIdentity represents a Terraform user/service identity
type TerraformIdentity struct {
	id           string
	identityType TerraformIdentityType
	subject      string
	accessToken  string
	expiresAt    *time.Time
	permissions  []string
	metadata     map[string]string
}

// NewTerraformIdentity creates a new Terraform identity
func NewTerraformIdentity(id, subject string, identityType TerraformIdentityType) (*TerraformIdentity, error) {
	if id == "" {
		return nil, shared.ErrInvalidInput
	}
	if subject == "" {
		return nil, shared.ErrInvalidInput
	}

	return &TerraformIdentity{
		id:           id,
		identityType: identityType,
		subject:      subject,
		permissions:  make([]string, 0),
		metadata:     make(map[string]string),
	}, nil
}

// ReconstructTerraformIdentity reconstructs a Terraform identity from persistence
func ReconstructTerraformIdentity(id, subject string, identityType TerraformIdentityType, accessToken string, expiresAt *time.Time, permissions []string, metadata map[string]string) *TerraformIdentity {
	if permissions == nil {
		permissions = make([]string, 0)
	}
	if metadata == nil {
		metadata = make(map[string]string)
	}

	return &TerraformIdentity{
		id:           id,
		identityType: identityType,
		subject:      subject,
		accessToken:  accessToken,
		expiresAt:    expiresAt,
		permissions:  permissions,
		metadata:     metadata,
	}
}

// Getters

func (ti *TerraformIdentity) ID() string {
	return ti.id
}

func (ti *TerraformIdentity) IdentityType() TerraformIdentityType {
	return ti.identityType
}

func (ti *TerraformIdentity) Subject() string {
	return ti.subject
}

func (ti *TerraformIdentity) AccessToken() string {
	return ti.accessToken
}

func (ti *TerraformIdentity) ExpiresAt() *time.Time {
	return ti.expiresAt
}

func (ti *TerraformIdentity) Permissions() []string {
	return ti.permissions
}

func (ti *TerraformIdentity) Metadata() map[string]string {
	return ti.metadata
}

// Domain Methods

// SetAccessToken sets the access token and optionally expiration
func (ti *TerraformIdentity) SetAccessToken(token string, expiresAt *time.Time) {
	ti.accessToken = token
	ti.expiresAt = expiresAt
}

// AddPermission adds a permission to the identity
func (ti *TerraformIdentity) AddPermission(permission string) {
	for _, p := range ti.permissions {
		if p == permission {
			return // Already has permission
		}
	}
	ti.permissions = append(ti.permissions, permission)
}

// HasPermission checks if the identity has a specific permission
func (ti *TerraformIdentity) HasPermission(permission string) bool {
	for _, p := range ti.permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if the identity has all specified permissions
func (ti *TerraformIdentity) HasAllPermissions(permissions []string) bool {
	for _, permission := range permissions {
		if !ti.HasPermission(permission) {
			return false
		}
	}
	return true
}

// SetMetadata sets metadata key-value pair
func (ti *TerraformIdentity) SetMetadata(key, value string) {
	ti.metadata[key] = value
}

// GetMetadata gets metadata value by key
func (ti *TerraformIdentity) GetMetadata(key string) string {
	return ti.metadata[key]
}

// IsExpired checks if the identity has expired
func (ti *TerraformIdentity) IsExpired() bool {
	if ti.expiresAt == nil {
		return false // No expiration set
	}
	return time.Now().After(*ti.expiresAt)
}

// IsValid checks if the identity is valid (not expired)
func (ti *TerraformIdentity) IsValid() bool {
	return !ti.IsExpired()
}

// String representation of the identity type
func (t TerraformIdentityType) String() string {
	switch t {
	case TerraformIdentityTypeOIDC:
		return "oidc"
	case TerraformIdentityTypeAnalytics:
		return "analytics"
	case TerraformIdentityTypeInternalExtraction:
		return "internal-extraction"
	case TerraformIdentityTypeDeployment:
		return "deployment"
	default:
		return "unknown"
	}
}

// TerraformIdentityTypeFromString creates TerraformIdentityType from string
func TerraformIdentityTypeFromString(s string) TerraformIdentityType {
	switch s {
	case "oidc":
		return TerraformIdentityTypeOIDC
	case "analytics":
		return TerraformIdentityTypeAnalytics
	case "internal-extraction":
		return TerraformIdentityTypeInternalExtraction
	case "deployment":
		return TerraformIdentityTypeDeployment
	default:
		return TerraformIdentityTypeOIDC // Default
	}
}