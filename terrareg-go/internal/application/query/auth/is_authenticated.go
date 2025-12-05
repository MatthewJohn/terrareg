package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// IsAuthenticatedQuery handles getting current authentication status
type IsAuthenticatedQuery struct {
	authFactory *service.AuthFactory
}

// NewIsAuthenticatedQuery creates a new is authenticated query
func NewIsAuthenticatedQuery(authFactory *service.AuthFactory) *IsAuthenticatedQuery {
	return &IsAuthenticatedQuery{
		authFactory: authFactory,
	}
}

// Execute returns the current authentication status
func (q *IsAuthenticatedQuery) Execute(ctx context.Context) (*dto.IsAuthenticatedResponse, error) {
	authMethod := q.authFactory.GetCurrentAuthMethod()

	return &dto.IsAuthenticatedResponse{
		Authenticated:        authMethod.IsAuthenticated(),
		ReadAccess:          authMethod.CanAccessReadAPI(),
		SiteAdmin:           authMethod.IsAdmin(),
		NamespacePermissions: authMethod.GetAllNamespacePermissions(),
	}, nil
}