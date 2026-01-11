package model

import (
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// AuthContext consolidates all authentication-related context values
type AuthContext struct {
	// Authentication details
	AuthMethod auth.AuthMethodType `json:"auth_method"`
	Username   string              `json:"username"`
	IsAdmin    bool                `json:"is_admin"`
	SessionID  string              `json:"session_id,omitempty"`

	// Authorization details
	Permissions map[string]string `json:"permissions,omitempty"`

	// Request metadata
	IsAuthenticated bool `json:"is_authenticated"`
}

// HasPermission checks if the user has the specified permission for a namespace
func (ac *AuthContext) HasPermission(permissionType, namespace string) bool {
	if !ac.IsAuthenticated {
		return false
	}

	// Admins have all permissions
	if ac.IsAdmin {
		return true
	}

	// Check specific namespace permission
	storedPermission, exists := ac.Permissions[namespace]
	if !exists {
		return false
	}

	// Check permission hierarchy: FULL > MODIFY > READ
	switch permissionType {
	case "READ":
		return storedPermission == "READ" || storedPermission == "MODIFY" || storedPermission == "FULL"
	case "MODIFY":
		return storedPermission == "MODIFY" || storedPermission == "FULL"
	case "FULL":
		return storedPermission == "FULL"
	default:
		return false
	}
}

// GetHighestPermission returns the highest permission level for any namespace
func (ac *AuthContext) GetHighestPermission() string {
	highest := ""
	for _, permission := range ac.Permissions {
		switch permission {
		case "FULL":
			return "FULL"
		case "MODIFY":
			if highest != "FULL" {
				highest = "MODIFY"
			}
		case "READ":
			if highest == "" {
				highest = "READ"
			}
		}
	}
	return highest
}

// CanAccessNamespace checks if the user can access a namespace with any permission
func (ac *AuthContext) CanAccessNamespace(namespace string) bool {
	if !ac.IsAuthenticated {
		return false
	}

	// Admins can access all namespaces
	if ac.IsAdmin {
		return true
	}

	_, exists := ac.Permissions[namespace]
	return exists
}
