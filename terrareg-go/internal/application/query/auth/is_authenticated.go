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
	// Get domain auth context from context set by SessionMiddleware
	// Returns domain auth.AuthContext interface - never nil (returns NotAuthenticatedAuthContext if not set)
	authCtx := middleware.GetAuthenticationContext(ctx)

	return &dto.IsAuthenticatedResponse{
		Authenticated:        authCtx.IsAuthenticated(),
		ReadAccess:           authCtx.CanAccessReadAPI(),
		SiteAdmin:            authCtx.IsAdmin(),
		NamespacePermissions: authCtx.GetAllNamespacePermissions(),
	}, nil
}
