package provider_source

import (
	"context"
	"fmt"

	authRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	providerSourceModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// GetRepositoriesQuery retrieves repositories for a provider source
// Python reference: github_repositories.py::GithubRepositories
type GetRepositoriesQuery struct {
	providerSourceFactory *service.ProviderSourceFactory
	sessionRepo           authRepo.SessionRepository
	namespaceRepo         repository.NamespaceRepository
}

// NewGetRepositoriesQuery creates a new GetRepositoriesQuery
func NewGetRepositoriesQuery(
	providerSourceFactory *service.ProviderSourceFactory,
	sessionRepo authRepo.SessionRepository,
	namespaceRepo repository.NamespaceRepository,
) (*GetRepositoriesQuery, error) {
	if providerSourceFactory == nil {
		return nil, fmt.Errorf("providerSourceFactory cannot be nil")
	}
	if sessionRepo == nil {
		return nil, fmt.Errorf("sessionRepo cannot be nil")
	}
	if namespaceRepo == nil {
		return nil, fmt.Errorf("namespaceRepo cannot be nil")
	}

	return &GetRepositoriesQuery{
		providerSourceFactory: providerSourceFactory,
		sessionRepo:           sessionRepo,
		namespaceRepo:         namespaceRepo,
	}, nil
}

// GetRepositoriesRequest represents a request to get repositories
type GetRepositoriesRequest struct {
	// ProviderSource is the name of the provider source (e.g., "github")
	ProviderSource string
	// SessionID is the user's session ID for authentication
	SessionID string
	// IsAdmin indicates if the user is an admin (admins can see all repositories)
	IsAdmin bool
}

// Execute retrieves repositories for the authenticated user from a provider source
// Python reference: github_repositories.py::get()
func (q *GetRepositoriesQuery) Execute(ctx context.Context, req GetRepositoriesRequest) ([]*providerSourceModel.Repository, error) {
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

	// Get repositories from provider source
	repositories, err := providerSource.GetUserRepositories(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get repositories from provider source: %w", err)
	}

	return repositories, nil
}
