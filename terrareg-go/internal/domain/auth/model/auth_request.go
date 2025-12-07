package model

import (
	"context"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"time"
)

// AuthenticationRequest represents a request to authenticate
type AuthenticationRequest struct {
	// Request identification
	RequestID string          `json:"request_id"`
	Timestamp time.Time       `json:"timestamp"`
	Context   context.Context `json:"-"`

	// Authentication method
	AuthMethodType auth.AuthMethodType `json:"auth_method_type"`

	// Request data
	Headers     map[string]string      `json:"headers"`
	FormData    map[string]string      `json:"form_data"`
	JSONData    map[string]interface{} `json:"json_data"`
	QueryParams map[string]string      `json:"query_params"`

	// Session information
	SessionID *string `json:"session_id,omitempty"`
	ClientIP  string  `json:"client_ip"`
	UserAgent string  `json:"user_agent"`

	// Target resource (for permission checking)
	ResourceType *string `json:"resource_type,omitempty"`
	ResourceID   *string `json:"resource_id,omitempty"`
	Namespace    *string `json:"namespace,omitempty"`
	Action       *string `json:"action,omitempty"`
}

// AuthenticationResponse represents the result of an authentication attempt
type AuthenticationResponse struct {
	// Request identification
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`

	// Authentication result
	Success    bool                `json:"success"`
	AuthMethod auth.AuthMethodType `json:"auth_method"`
	Username   string              `json:"username,omitempty"`
	SessionID  *string             `json:"session_id,omitempty"`

	// Permission information
	IsAdmin            bool              `json:"is_admin"`
	UserGroups         []string          `json:"user_groups,omitempty"`
	Permissions        map[string]string `json:"permissions,omitempty"` // namespace -> permission_type
	CanPublish         bool              `json:"can_publish"`
	CanUpload          bool              `json:"can_upload"`
	CanAccessAPI       bool              `json:"can_access_api"`
	CanAccessTerraform bool              `json:"can_access_terraform"`

	// Tokens
	AuthToken      *string    `json:"auth_token,omitempty"`
	TerraformToken *string    `json:"terraform_token,omitempty"`
	TokenExpiry    *time.Time `json:"token_expiry,omitempty"`

	// Error information
	ErrorCode    *string `json:"error_code,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SessionRequest represents a request to create or manage a session
type SessionRequest struct {
	AuthMethod   auth.AuthMethodType    `json:"auth_method"`
	ProviderData map[string]interface{} `json:"provider_data"`
	TTL          *time.Duration         `json:"ttl,omitempty"`
	ClientIP     string                 `json:"client_ip"`
	UserAgent    string                 `json:"user_agent"`
	RememberMe   bool                   `json:"remember_me"`
}

// SessionResponse represents session information
type SessionResponse struct {
	SessionID      string                 `json:"session_id"`
	Expiry         time.Time              `json:"expiry"`
	AuthMethod     auth.AuthMethodType    `json:"auth_method"`
	ProviderData   map[string]interface{} `json:"provider_data"`
	IsActive       bool                   `json:"is_active"`
	CreatedAt      time.Time              `json:"created_at"`
	LastAccessedAt *time.Time             `json:"last_accessed_at,omitempty"`
}

// PermissionCheckRequest represents a permission checking request
type PermissionCheckRequest struct {
	UserGroupID    *int                `json:"user_group_id,omitempty"`
	NamespaceID    *int                `json:"namespace_id,omitempty"`
	NamespaceName  *string             `json:"namespace_name,omitempty"`
	PermissionType auth.PermissionType `json:"permission_type"`
	ResourceType   *string             `json:"resource_type,omitempty"`
	Action         *string             `json:"action,omitempty"`
}

// PermissionCheckResponse represents the result of a permission check
type PermissionCheckResponse struct {
	Allowed        bool                 `json:"allowed"`
	PermissionType *auth.PermissionType `json:"permission_type,omitempty"`
	UserGroupID    *int                 `json:"user_group_id,omitempty"`
	NamespaceID    *int                 `json:"namespace_id,omitempty"`
	NamespaceName  *string              `json:"namespace_name,omitempty"`
	Reason         *string              `json:"reason,omitempty"`
}

// TerraformAuthRequest represents a Terraform CLI authentication request
type TerraformAuthRequest struct {
	AuthorizationHeader string                 `json:"authorization_header"`
	ClientID            string                 `json:"client_id"`
	RequestType         string                 `json:"request_type"` // userinfo, token, etc.
	Scopes              []string               `json:"scopes,omitempty"`
	SubjectIdentifier   string                 `json:"subject_identifier"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// TerraformAuthResponse represents a Terraform authentication response
type TerraformAuthResponse struct {
	Valid             bool                   `json:"valid"`
	SubjectIdentifier string                 `json:"subject_identifier"`
	Username          string                 `json:"username"`
	Permissions       []string               `json:"permissions,omitempty"`
	ExpirationTime    *time.Time             `json:"expiration_time,omitempty"`
	TokenType         string                 `json:"token_type,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// APIKeyAuthRequest represents an API key authentication request
type APIKeyAuthRequest struct {
	APIKey    string  `json:"api_key"`
	KeyType   string  `json:"key_type"` // admin, upload, publish, analytics
	ClientIP  string  `json:"client_ip"`
	UserAgent string  `json:"user_agent"`
	Resource  *string `json:"resource,omitempty"` // Optional resource being accessed
}

// APIKeyAuthResponse represents an API key authentication response
type APIKeyAuthResponse struct {
	Valid       bool                   `json:"valid"`
	KeyType     string                 `json:"key_type"`
	Username    string                 `json:"username,omitempty"`
	UserGroups  []string               `json:"user_groups,omitempty"`
	Permissions map[string]string      `json:"permissions,omitempty"`
	ExpiryTime  *time.Time             `json:"expiry_time,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Helper functions for creating requests and responses

// NewAuthenticationRequest creates a new authentication request
func NewAuthenticationRequest(authMethod auth.AuthMethodType, headers, formData, queryParams map[string]string) *AuthenticationRequest {
	return &AuthenticationRequest{
		RequestID:      shared.GenerateID(),
		Timestamp:      time.Now(),
		AuthMethodType: authMethod,
		Headers:        headers,
		FormData:       formData,
		QueryParams:    queryParams,
		JSONData:       make(map[string]interface{}),
		Context:        context.Background(),
	}
}

// NewAuthenticationResponse creates a new authentication response
func NewAuthenticationResponse(requestID string, success bool, authMethod auth.AuthMethodType) *AuthenticationResponse {
	return &AuthenticationResponse{
		RequestID:   requestID,
		Timestamp:   time.Now(),
		Success:     success,
		AuthMethod:  authMethod,
		Permissions: make(map[string]string),
		Metadata:    make(map[string]interface{}),
	}
}

// NewSessionRequest creates a new session request
func NewSessionRequest(authMethod auth.AuthMethodType, providerData map[string]interface{}) *SessionRequest {
	return &SessionRequest{
		AuthMethod:   authMethod,
		ProviderData: providerData,
		TTL:          nil, // Use default
		RememberMe:   false,
	}
}

// NewTerraformAuthRequest creates a Terraform authentication request
func NewTerraformAuthRequest(authHeader, subjectIdentifier string) *TerraformAuthRequest {
	return &TerraformAuthRequest{
		AuthorizationHeader: authHeader,
		SubjectIdentifier:   subjectIdentifier,
		RequestType:         "userinfo",
		Scopes:              []string{},
		Metadata:            make(map[string]interface{}),
	}
}
