package model

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// AuthenticationTokenType represents the type of authentication token
type AuthenticationTokenType int

const (
	// AuthenticationTokenTypeAdmin represents admin tokens with full access
	AuthenticationTokenTypeAdmin AuthenticationTokenType = iota
	// AuthenticationTokenTypeUpload represents upload tokens for module uploads
	AuthenticationTokenTypeUpload
	// AuthenticationTokenTypePublish represents publish tokens for module publishing
	AuthenticationTokenTypePublish
)

// String returns the string representation of the token type
func (t AuthenticationTokenType) String() string {
	switch t {
	case AuthenticationTokenTypeAdmin:
		return "admin"
	case AuthenticationTokenTypeUpload:
		return "upload"
	case AuthenticationTokenTypePublish:
		return "publish"
	default:
		return "unknown"
	}
}

// ToAuthMethodType converts token type to auth method type
func (t AuthenticationTokenType) ToAuthMethodType() auth.AuthMethodType {
	switch t {
	case AuthenticationTokenTypeAdmin:
		return auth.AuthMethodAdminApiKey
	case AuthenticationTokenTypeUpload:
		return auth.AuthMethodUploadApiKey
	case AuthenticationTokenTypePublish:
		return auth.AuthMethodPublishApiKey
	default:
		return auth.AuthMethodNotAuthenticated
	}
}

// AuthenticationTokenTypeFromString converts string to token type
func AuthenticationTokenTypeFromString(s string) (AuthenticationTokenType, error) {
	switch strings.ToLower(s) {
	case "admin":
		return AuthenticationTokenTypeAdmin, nil
	case "upload":
		return AuthenticationTokenTypeUpload, nil
	case "publish":
		return AuthenticationTokenTypePublish, nil
	default:
		return AuthenticationTokenTypeAdmin, fmt.Errorf("%w: %s", ErrInvalidTokenType, s)
	}
}

// AuthenticationToken represents an authentication token for API access
type AuthenticationToken struct {
	id          int
	tokenType   AuthenticationTokenType
	tokenValue  string
	namespace   *model.Namespace // Only for publish tokens
	description string
	createdAt   time.Time
	expiresAt   *time.Time
	isActive    bool
	createdBy   string
}

// NewAuthenticationToken creates a new authentication token
func NewAuthenticationToken(
	tokenType AuthenticationTokenType,
	description string,
	namespace *model.Namespace,
	expiresAt *time.Time,
	createdBy string,
) (*AuthenticationToken, error) {
	// Validate token type
	if tokenType != AuthenticationTokenTypeAdmin &&
	   tokenType != AuthenticationTokenTypeUpload &&
	   tokenType != AuthenticationTokenTypePublish {
		return nil, ErrInvalidTokenType
	}

	// Validate description
	if strings.TrimSpace(description) == "" {
		return nil, ErrDescriptionRequired
	}

	if len(description) > 255 {
		return nil, ErrDescriptionTooLong
	}

	// For publish tokens, namespace is required
	if tokenType == AuthenticationTokenTypePublish && namespace == nil {
		return nil, fmt.Errorf("%w: namespace is required for publish tokens", shared.ErrInvalidNamespace)
	}

	// For non-publish tokens, namespace should not be set
	if tokenType != AuthenticationTokenTypePublish && namespace != nil {
		return nil, fmt.Errorf("%w: namespace is only allowed for publish tokens", shared.ErrInvalidNamespace)
	}

	// Validate created by
	if strings.TrimSpace(createdBy) == "" {
		return nil, ErrCreatedByRequired
	}

	// Generate secure token value
	tokenValue, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Validate expiration if provided
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		return nil, ErrExpirationInPast
	}

	now := time.Now()
	return &AuthenticationToken{
		id:          shared.GenerateIntID(),
		tokenType:   tokenType,
		tokenValue:  tokenValue,
		namespace:   namespace,
		description: description,
		createdAt:   now,
		expiresAt:   expiresAt,
		isActive:    true,
		createdBy:   createdBy,
	}, nil
}

// ReconstructAuthenticationToken reconstructs an authentication token from persistence
func ReconstructAuthenticationToken(
	id int,
	tokenType AuthenticationTokenType,
	tokenValue string,
	namespace *model.Namespace,
	description string,
	createdAt time.Time,
	expiresAt *time.Time,
	isActive bool,
	createdBy string,
) *AuthenticationToken {
	return &AuthenticationToken{
		id:          id,
		tokenType:   tokenType,
		tokenValue:  tokenValue,
		namespace:   namespace,
		description: description,
		createdAt:   createdAt,
		expiresAt:   expiresAt,
		isActive:    isActive,
		createdBy:   createdBy,
	}
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Business methods

// Validate checks if the token is valid for authentication
func (t *AuthenticationToken) Validate() error {
	if !t.isActive {
		return ErrTokenInactive
	}

	if t.IsExpired() {
		return ErrTokenExpired
	}

	return nil
}

// IsExpired checks if the token has expired
func (t *AuthenticationToken) IsExpired() bool {
	if t.expiresAt == nil {
		return false
	}
	return time.Now().After(*t.expiresAt)
}

// CanAccessNamespace checks if the token can access a specific namespace
func (t *AuthenticationToken) CanAccessNamespace(namespace string) bool {
	// Admin tokens can access all namespaces
	if t.tokenType == AuthenticationTokenTypeAdmin {
		return true
	}

	// Upload tokens don't have namespace restrictions
	if t.tokenType == AuthenticationTokenTypeUpload {
		return true
	}

	// Publish tokens can only access their specific namespace
	if t.tokenType == AuthenticationTokenTypePublish && t.namespace != nil {
		return t.namespace.Name() == namespace
	}

	return false
}

// Revoke deactivates the token
func (t *AuthenticationToken) Revoke() error {
	if !t.isActive {
		return ErrTokenAlreadyRevoked
	}

	t.isActive = false
	return nil
}

// Getters

func (t *AuthenticationToken) ID() int                        { return t.id }
func (t *AuthenticationToken) TokenType() AuthenticationTokenType { return t.tokenType }
func (t *AuthenticationToken) TokenValue() string              { return t.tokenValue }
func (t *AuthenticationToken) Namespace() *model.Namespace     { return t.namespace }
func (t *AuthenticationToken) Description() string             { return t.description }
func (t *AuthenticationToken) CreatedAt() time.Time            { return t.createdAt }
func (t *AuthenticationToken) ExpiresAt() *time.Time           { return t.expiresAt }
func (t *AuthenticationToken) IsActive() bool                  { return t.isActive }
func (t *AuthenticationToken) CreatedBy() string               { return t.createdBy }

// HasNamespace returns true if the token has an associated namespace
func (t *AuthenticationToken) HasNamespace() bool {
	return t.namespace != nil
}

// IsAdmin returns true if this is an admin token
func (t *AuthenticationToken) IsAdmin() bool {
	return t.tokenType == AuthenticationTokenTypeAdmin
}

// IsUpload returns true if this is an upload token
func (t *AuthenticationToken) IsUpload() bool {
	return t.tokenType == AuthenticationTokenTypeUpload
}

// IsPublish returns true if this is a publish token
func (t *AuthenticationToken) IsPublish() bool {
	return t.tokenType == AuthenticationTokenTypePublish
}

// String returns a string representation (without the actual token value)
func (t *AuthenticationToken) String() string {
	var namespaceStr string
	if t.namespace != nil {
		namespaceStr = t.namespace.Name()
	}
	return fmt.Sprintf("AuthenticationToken{id=%d, type=%s, namespace=%s, active=%t}",
		t.id, t.tokenType.String(), namespaceStr, t.isActive)
}

// GetAuthMethod returns the corresponding auth method for this token
func (t *AuthenticationToken) GetAuthMethod() auth.AuthMethodType {
	return t.tokenType.ToAuthMethodType()
}

// GetDisplayName returns a display name for the token
func (t *AuthenticationToken) GetDisplayName() string {
	if t.description != "" {
		return t.description
	}
	return fmt.Sprintf("%s Token", strings.Title(t.tokenType.String()))
}

// TimeToExpiration returns the duration until expiration
func (t *AuthenticationToken) TimeToExpiration() time.Duration {
	if t.expiresAt == nil {
		return time.Duration(0) // No expiration
	}
	return time.Until(*t.expiresAt)
}