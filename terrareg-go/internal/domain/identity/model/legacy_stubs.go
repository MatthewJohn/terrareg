package model

import (
	"context"
	"errors"
	"time"
)

// Temporary stub implementations for legacy user-based services
// These should be removed as part of the migration to group-based auth

// User represents a user (legacy - should be removed)
// TODO: Remove this as part of migration to group-based authentication
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthMethod represents authentication methods (legacy)
// TODO: Remove this as part of migration to auth.AuthMethodType
type AuthMethod string

// Permission represents permissions (legacy)
// TODO: Remove this as part of migration to auth.PermissionType
type Permission struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Action       string    `json:"action"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthResult represents authentication result (legacy)
type AuthResult struct {
	User      *User    `json:"user"`
	Token     string   `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// AuthProvider represents auth provider (legacy)
type AuthProvider interface {
	Authenticate(ctx context.Context, token string) (*AuthResult, error)
	GetName() string
	GetType() AuthMethod
}

// Session represents a user session (legacy - should be removed)
// TODO: Remove this as part of migration to auth domain sessions
type Session struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// UserGroup represents user group (legacy)
// TODO: Remove this as part of migration to auth domain user groups
type UserGroup struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResourceType represents resource types (legacy)
type ResourceType string

// Action represents actions (legacy)
type Action string

// IDPAccessToken represents IDP access token (legacy)
type IDPAccessToken struct {
	ID        int       `json:"id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Legacy error constants (should be replaced with auth domain errors)
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserInactive     = errors.New("user is inactive")
	ErrInvalidAPIKey    = errors.New("invalid API key")
	ErrAPIKeyNotFound   = errors.New("API key not found")
)