package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

// IsAuthenticatedQuery handles getting current authentication status
type IsAuthenticatedQuery struct{}

// NewIsAuthenticatedQuery creates a new is authenticated query
func NewIsAuthenticatedQuery() *IsAuthenticatedQuery {
	return &IsAuthenticatedQuery{}
}

// Execute returns the current authentication status
func (q *IsAuthenticatedQuery) Execute(ctx context.Context) (*dto.IsAuthenticatedResponse, error) {
	// Get session data from context set by SessionMiddleware
	sessionData := middleware.GetSessionData(ctx)

	// If no session data, user is not authenticated
	if sessionData == nil {
		return &dto.IsAuthenticatedResponse{
			Authenticated:        false,
			ReadAccess:          false,
			SiteAdmin:           false,
			NamespacePermissions: make(map[string]string),
		}, nil
	}

	// Return authentication status from session data
	permissions := sessionData.Permissions
	if permissions == nil {
		permissions = make(map[string]string)
	}

	return &dto.IsAuthenticatedResponse{
		Authenticated:        true, // Session exists, so authenticated
		ReadAccess:          true, // Authenticated users have read access
		SiteAdmin:           sessionData.IsAdmin,
		NamespacePermissions: permissions,
	}, nil
}