package provider

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	providerService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/service"
	providerSourceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ProviderExtractionOrchestratorInterface defines the interface for provider extraction orchestration
type ProviderExtractionOrchestratorInterface interface {
	ExtractProviderVersion(
		ctx context.Context,
		req providerService.ExtractProviderVersionRequest,
		repository *sqldb.RepositoryDB,
	) error
}

// ExtractProviderVersionCommand handles provider version extraction
type ExtractProviderVersionCommand struct {
	orchestrator         ProviderExtractionOrchestratorInterface
	providerRepo         providerRepo.ProviderRepository
	providerSourceFactory *providerSourceService.ProviderSourceFactory
	logger               zerolog.Logger
}

// ExtractProviderVersionRequest contains the parameters for extraction (external API)
type ExtractProviderVersionRequest struct {
	ProviderID   int    `json:"provider_id"`
	Version      string `json:"version"`
	GitTag       string `json:"git_tag"`
	ProviderName string `json:"provider_name"`
	Namespace    string `json:"namespace"`
}

// NewExtractProviderVersionCommand creates a new ExtractProviderVersionCommand
func NewExtractProviderVersionCommand(
	orchestrator ProviderExtractionOrchestratorInterface,
	providerRepo providerRepo.ProviderRepository,
	providerSourceFactory *providerSourceService.ProviderSourceFactory,
	logger zerolog.Logger,
) *ExtractProviderVersionCommand {
	return &ExtractProviderVersionCommand{
		orchestrator:          orchestrator,
		providerRepo:          providerRepo,
		providerSourceFactory: providerSourceFactory,
		logger:                logger,
	}
}

// Execute performs provider version extraction
func (c *ExtractProviderVersionCommand) Execute(ctx context.Context, req ExtractProviderVersionRequest) error {
	c.logger.Info().
		Int("provider_id", req.ProviderID).
		Str("provider", req.ProviderName).
		Str("namespace", req.Namespace).
		Str("version", req.Version).
		Msg("Executing provider version extraction")

	// Validate provider exists
	providerEntity, err := c.providerRepo.FindByID(ctx, req.ProviderID)
	if err != nil {
		return fmt.Errorf("failed to find provider: %w", err)
	}
	if providerEntity == nil {
		return fmt.Errorf("provider not found: %d", req.ProviderID)
	}

	// TODO: Get repository from provider and convert to RepositoryDB
	// For now, we'll create a minimal RepositoryDB structure
	// This will need to be implemented based on how providers are associated with repositories
	repositoryDB := &sqldb.RepositoryDB{
		ID:       0, // TODO: Get actual repository ID from provider
		Name:     &req.ProviderName,
		CloneURL: nil, // TODO: Get actual clone URL
	}

	// Build extraction request
	extractionReq := providerService.ExtractProviderVersionRequest{
		ProviderID:   req.ProviderID,
		Version:      req.Version,
		GitTag:       req.GitTag,
		ProviderName: req.ProviderName,
		Namespace:    req.Namespace,
	}

	// Execute extraction
	if err := c.orchestrator.ExtractProviderVersion(ctx, extractionReq, repositoryDB); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	c.logger.Info().
		Int("provider_id", req.ProviderID).
		Str("version", req.Version).
		Msg("Provider version extraction completed successfully")

	return nil
}
