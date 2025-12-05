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
	// Get the auth method instance from context (set by middleware)
	authMethod, hasAuth := middleware.GetAuthMethodInstanceFromContext(ctx)

	// If no auth method in context, return unauthenticated
	if !hasAuth || authMethod == nil {
		return &dto.IsAuthenticatedResponse{
			Authenticated:        false,
			ReadAccess:          false,
			SiteAdmin:           false,
			NamespacePermissions: make(map[string]string),
		}, nil
	}

	// Get authentication status from the auth method instance
	return &dto.IsAuthenticatedResponse{
		Authenticated:        authMethod.IsAuthenticated(),
		ReadAccess:          authMethod.CanAccessReadAPI(),
		SiteAdmin:           authMethod.IsAdmin(),
		NamespacePermissions: authMethod.GetAllNamespacePermissions(),
	}, nil
}