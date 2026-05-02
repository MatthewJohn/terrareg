package service

import (
	"context"
)

// authContextKeyType is the type for context keys to avoid collisions
type authContextKeyType string

const (
	// This must match the key used in internal/interfaces/http/middleware/auth.go
	authContextKey authContextKeyType = "auth_middleware_context"
)

// authContext represents the minimal auth context we need for audit logging
// This mirrors the structure in middleware/model/auth_context.go
type authContext struct {
	Username        string
	IsAuthenticated bool
}

// getUsernameFromContext extracts the username from context for audit logging
// Returns "Built-in admin" for system operations or when no user is authenticated
func getUsernameFromContext(ctx context.Context) string {
	if ctx != nil {
		// Try the typed key first
		if authCtx, ok := ctx.Value(authContextKey).(*authContext); ok && authCtx.IsAuthenticated && authCtx.Username != "" {
			return authCtx.Username
		}
		// Also try with string key for compatibility
		if authCtx, ok := ctx.Value("auth_middleware_context").(*authContext); ok && authCtx.IsAuthenticated && authCtx.Username != "" {
			return authCtx.Username
		}
	}
	// Default to built-in admin for system operations
	return "Built-in admin"
}
