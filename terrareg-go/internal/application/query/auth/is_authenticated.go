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
	// Get authentication context from context set by SessionMiddleware
	authCtx := middleware.GetAuthenticationContext(ctx)

	// If no authentication context, user is not authenticated
	if authCtx == nil {
		return &dto.IsAuthenticatedResponse{
			Authenticated:        false,
			ReadAccess:           false,
			SiteAdmin:            false,
			NamespacePermissions: make(map[string]string),
		}, nil
	}

	// Return authentication status from authentication context
	permissions := authCtx.Permissions
	if permissions == nil {
		permissions = make(map[string]string)
	}

	return &dto.IsAuthenticatedResponse{
		Authenticated:        authCtx.IsAuthenticated,
		ReadAccess:           authCtx.IsAuthenticated, // Authenticated users have read access
		SiteAdmin:            authCtx.IsAdmin,
		NamespacePermissions: permissions,
	}, nil
}
