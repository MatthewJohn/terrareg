package provider_source

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// GetOrganizationsQuery retrieves organizations for a provider source
type GetOrganizationsQuery struct {
	providerSourceFactory *service.ProviderSourceFactory
	sessionRepo           repository.SessionRepository
}

// NewGetOrganizationsQuery creates a new GetOrganizationsQuery
func NewGetOrganizationsQuery(
	providerSourceFactory *service.ProviderSourceFactory,
	sessionRepo repository.SessionRepository,
) (*GetOrganizationsQuery, error) {
	if providerSourceFactory == nil {
		return nil, fmt.Errorf("providerSourceFactory cannot be nil")
	}
	if sessionRepo == nil {
		return nil, fmt.Errorf("sessionRepo cannot be nil")
	}

	return &GetOrganizationsQuery{
		providerSourceFactory: providerSourceFactory,
		sessionRepo:           sessionRepo,
	}, nil
}

// GetOrganizationsRequest represents a request to get organizations
type GetOrganizationsRequest struct {
	// ProviderSource is the name of the provider source (e.g., "github")
	ProviderSource string
	// SessionID is the user's session ID for authentication
	SessionID string
}

// Execute retrieves organizations for the authenticated user from a provider source
func (q *GetOrganizationsQuery) Execute(ctx context.Context, req GetOrganizationsRequest) ([]*model.Organization, error) {
	// Validate inputs
	if req.ProviderSource == "" {
		return nil, fmt.Errorf("provider_source is required")
	}
	if req.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	// Get provider source instance
	providerSource, err := q.providerSourceFactory.GetProviderSourceByName(ctx, req.ProviderSource)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider source: %w", err)
	}
	if providerSource == nil {
		return nil, shared.ErrNotFound
	}

	// Get organizations from provider source
	organizations, err := providerSource.GetUserOrganizationsList(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations from provider source: %w", err)
	}

	return organizations, nil
}
