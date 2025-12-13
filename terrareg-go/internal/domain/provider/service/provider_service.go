package service

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// ProviderService handles business logic for providers
type ProviderService struct {
	providerRepo repository.ProviderRepository
}

// NewProviderService creates a new provider service
func NewProviderService(providerRepo repository.ProviderRepository) *ProviderService {
	return &ProviderService{
		providerRepo: providerRepo,
	}
}

// CreateProviderRequest represents a request to create a provider
type CreateProviderRequest struct {
	NamespaceID           int
	Name                  string
	Description           *string
	Tier                  string
	CategoryID            *int
	RepositoryID          *int
	UseProviderSourceAuth bool
}

// CreateProvider creates a new provider
func (s *ProviderService) CreateProvider(ctx context.Context, req CreateProviderRequest) (*provider.Provider, error) {
	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("provider name is required")
	}
	if req.NamespaceID <= 0 {
		return nil, fmt.Errorf("valid namespace ID is required")
	}

	// Check if provider already exists
	existing, err := s.providerRepo.FindByNamespaceAndName(ctx, "", req.Name) // TODO: Get namespace name from ID
	if err == nil && existing != nil {
		return nil, fmt.Errorf("provider %s already exists", req.Name)
	}

	// Create new provider
	newProvider := provider.NewProvider(
		req.NamespaceID,
		req.Name,
		req.Description,
		req.Tier,
		req.CategoryID,
		req.RepositoryID,
		req.UseProviderSourceAuth,
	)

	// Save to repository
	if err := s.providerRepo.Save(ctx, newProvider); err != nil {
		return nil, fmt.Errorf("failed to save provider: %w", err)
	}

	return newProvider, nil
}

// UpdateProviderRequest represents a request to update a provider
type UpdateProviderRequest struct {
	Description           *string
	Tier                  string
	RepositoryID          *int
	UseProviderSourceAuth *bool
}

// UpdateProvider updates an existing provider
func (s *ProviderService) UpdateProvider(ctx context.Context, providerID int, req UpdateProviderRequest) (*provider.Provider, error) {
	// Find existing provider
	existingProvider, err := s.providerRepo.FindByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}
	if existingProvider == nil {
		return nil, fmt.Errorf("provider not found")
	}

	// Update fields
	if req.Description != nil {
		existingProvider.SetDescription(req.Description)
	}
	if req.Tier != "" {
		existingProvider.SetTier(req.Tier)
	}
	if req.RepositoryID != nil {
		existingProvider.SetRepositoryID(req.RepositoryID)
	}
	if req.UseProviderSourceAuth != nil {
		existingProvider.SetUseProviderSourceAuth(*req.UseProviderSourceAuth)
	}

	// Save changes
	if err := s.providerRepo.Save(ctx, existingProvider); err != nil {
		return nil, fmt.Errorf("failed to update provider: %w", err)
	}

	return existingProvider, nil
}

// DeleteProvider deletes a provider
func (s *ProviderService) DeleteProvider(ctx context.Context, providerID int) error {
	// Find provider
	existingProvider, err := s.providerRepo.FindByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to find provider: %w", err)
	}
	if existingProvider == nil {
		return fmt.Errorf("provider not found")
	}

	// Check if provider has versions - if so, cannot delete
	versions, err := s.providerRepo.FindVersionsByProvider(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to check provider versions: %w", err)
	}
	if len(versions) > 0 {
		return fmt.Errorf("cannot delete provider with existing versions")
	}

	// TODO: Implement delete in repository interface
	// For now, this is a placeholder
	return nil
}

// GetProvider retrieves a provider by ID
func (s *ProviderService) GetProvider(ctx context.Context, providerID int) (*provider.Provider, error) {
	provider, err := s.providerRepo.FindByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}
	return provider, nil
}

// GetProviderByName retrieves a provider by namespace and name
func (s *ProviderService) GetProviderByName(ctx context.Context, namespace, name string) (*provider.Provider, error) {
	provider, err := s.providerRepo.FindByNamespaceAndName(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}
	return provider, nil
}

// ListProviders retrieves a paginated list of providers
func (s *ProviderService) ListProviders(ctx context.Context, offset, limit int) ([]*provider.Provider, int, error) {
	providers, total, err := s.providerRepo.FindAll(ctx, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list providers: %w", err)
	}
	return providers, total, nil
}

// SearchProviders searches for providers
func (s *ProviderService) SearchProviders(ctx context.Context, query string, offset, limit int) ([]*provider.Provider, int, error) {
	providers, total, err := s.providerRepo.Search(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search providers: %w", err)
	}
	return providers, total, nil
}

// PublishVersionRequest represents a request to publish a provider version
type PublishVersionRequest struct {
	Version          string
	GitTag           *string
	ProtocolVersions []string
	IsBeta           bool
}

// PublishVersion publishes a new version of a provider
func (s *ProviderService) PublishVersion(ctx context.Context, providerID int, req PublishVersionRequest) (*provider.ProviderVersion, error) {
	// Find provider
	providerEntity, err := s.providerRepo.FindByID(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}
	if providerEntity == nil {
		return nil, fmt.Errorf("provider not found")
	}

	// Check if version already exists
	existingVersion, err := s.providerRepo.FindVersionByProviderAndVersion(ctx, providerID, req.Version)
	if err == nil && existingVersion != nil {
		return nil, fmt.Errorf("version %s already exists", req.Version)
	}

	// Create version using aggregate
	newVersion, err := providerEntity.PublishVersion(req.Version, req.ProtocolVersions, req.IsBeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Set git tag if provided
	if req.GitTag != nil {
		newVersion.SetGitTag(req.GitTag)
	}

	// Save version
	if err := s.providerRepo.SaveVersion(ctx, newVersion); err != nil {
		return nil, fmt.Errorf("failed to save version: %w", err)
	}

	// Set as latest version if this is the first version or explicitly requested
	// TODO: Add logic to determine if this should be latest
	if err := s.providerRepo.SetLatestVersion(ctx, providerID, newVersion.ID()); err != nil {
		// Non-critical error, log but don't fail
		fmt.Printf("Warning: failed to set latest version: %v\n", err)
	}

	return newVersion, nil
}

// GetVersions retrieves all versions for a provider
func (s *ProviderService) GetVersions(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error) {
	versions, err := s.providerRepo.FindVersionsByProvider(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	return versions, nil
}

// GetVersion retrieves a specific version of a provider
func (s *ProviderService) GetVersion(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error) {
	versionEntity, err := s.providerRepo.FindVersionByProviderAndVersion(ctx, providerID, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}
	return versionEntity, nil
}

// DeleteVersion deletes a provider version
func (s *ProviderService) DeleteVersion(ctx context.Context, providerID, versionID int) error {
	// Find provider
	providerEntity, err := s.providerRepo.FindByID(ctx, providerID)
	if err != nil {
		return fmt.Errorf("failed to find provider: %w", err)
	}
	if providerEntity == nil {
		return fmt.Errorf("provider not found")
	}

	// Check if version is the latest
	if providerEntity.LatestVersionID() != nil && *providerEntity.LatestVersionID() == versionID {
		return fmt.Errorf("cannot delete latest version")
	}

	// Delete from repository
	if err := s.providerRepo.DeleteVersion(ctx, versionID); err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}

	return nil
}