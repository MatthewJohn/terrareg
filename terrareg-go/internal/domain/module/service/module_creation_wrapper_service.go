package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	moduleModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// ModuleCreationWrapperService handles module creation with prepare/extract/publish pattern
// This matches the Python module_create_extraction_wrapper context manager
type ModuleCreationWrapperService struct {
	moduleVersionRepo  repository.ModuleVersionRepository
	moduleProviderRepo repository.ModuleProviderRepository
	savepointHelper    *transaction.SavepointHelper
	domainConfig       *configModel.DomainConfig
}

// NewModuleCreationWrapperService creates a new module creation wrapper service
func NewModuleCreationWrapperService(
	moduleVersionRepo repository.ModuleVersionRepository,
	moduleProviderRepo repository.ModuleProviderRepository,
	savepointHelper *transaction.SavepointHelper,
	domainConfig *configModel.DomainConfig,
) *ModuleCreationWrapperService {
	return &ModuleCreationWrapperService{
		moduleVersionRepo:  moduleVersionRepo,
		moduleProviderRepo: moduleProviderRepo,
		savepointHelper:    savepointHelper,
		domainConfig:       domainConfig,
	}
}

// PrepareModuleRequest represents a request to prepare a module version
type PrepareModuleRequest struct {
	Namespace        string
	ModuleName       string
	Provider         string
	Version          string
	GitTag           *string
	SourceGitTag     *string
	CommitSHA        *string // Git commit SHA for git-based imports
	ModuleProviderID *int
}

// PrepareModuleResult represents the result of module preparation
type PrepareModuleResult struct {
	ModuleVersion *moduleModel.ModuleVersion
	ShouldPublish bool
}

// PrepareModule creates the database row for the module version and returns whether it should be published
// This matches the Python prepare_module() method with reindex mode support
func (s *ModuleCreationWrapperService) PrepareModule(ctx context.Context, req PrepareModuleRequest) (*PrepareModuleResult, error) {
	// Validate ModuleProviderID is provided
	if req.ModuleProviderID == nil {
		return nil, fmt.Errorf("module provider ID is required for module version creation")
	}

	var result *PrepareModuleResult
	err := s.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Find the module provider to associate with this module version
		moduleProvider, err := s.moduleProviderRepo.FindByID(ctx, *req.ModuleProviderID)
		if err != nil {
			return fmt.Errorf("failed to find module provider with ID %d: %w", *req.ModuleProviderID, err)
		}

		// Check if version already exists and handle reindexing logic
		reindexMode := s.domainConfig.ModuleVersionReindexMode
		var existingVersion *moduleModel.ModuleVersion

		// Look for existing version in the module provider
		for _, v := range moduleProvider.GetAllVersions() {
			if v.Version().String() == req.Version {
				existingVersion = v
				break
			}
		}

		// Handle based on reindex mode
		if existingVersion != nil {
			switch reindexMode {
			case configModel.ModuleVersionReindexModeProhibit:
				return fmt.Errorf("version %s already exists and reindex mode is prohibit", req.Version)
			case configModel.ModuleVersionReindexModeLegacy:
				// In legacy mode, we need to delete the existing version and create a new one
				logger := zerolog.Ctx(ctx)
				logger.Debug().
					Str("version", req.Version).
					Str("reindex_mode", "legacy").
					Int("existing_version_id", existingVersion.ID()).
					Msg("Deleting existing module version for reindexing")

				// Delete existing version (this will cascade delete related data)
				if err := s.moduleVersionRepo.Delete(ctx, existingVersion.ID()); err != nil {
					return fmt.Errorf("failed to delete existing module version %s: %w", req.Version, err)
				}

				// CRITICAL FIX: Explicitly clear the ID from the deleted version's in-memory representation
				// to prevent the new version from inheriting the old ID
				logger.Debug().
					Int("deleted_version_id", existingVersion.ID()).
					Str("deleted_version", existingVersion.Version().String()).
					Msg("Clearing ID from deleted version to prevent inheritance")

				// Clear the ID from the deleted version to ensure new versions start with ID=0
				existingVersion.ResetID()

				// Remove the existing version from the module provider's in-memory list to avoid conflicts
				// Filter out the existing version and set the updated list
				updatedVersions := make([]*moduleModel.ModuleVersion, 0)
				for _, v := range moduleProvider.GetAllVersions() {
					if v.ID() != existingVersion.ID() { // Note: This should be 0 after ResetID
						updatedVersions = append(updatedVersions, v)
					}
				}
				moduleProvider.SetVersions(updatedVersions)
			case configModel.ModuleVersionReindexModeAutoPublish:
				// In auto-publish mode, preserve published state
				logger := zerolog.Ctx(ctx)
				logger.Debug().
					Str("version", req.Version).
					Str("reindex_mode", "auto-publish").
					Int("existing_version_id", existingVersion.ID()).
					Msg("Deleting existing module version for reindexing (preserving published state)")

				// Delete existing version (this will cascade delete related data)
				if err := s.moduleVersionRepo.Delete(ctx, existingVersion.ID()); err != nil {
					return fmt.Errorf("failed to delete existing module version %s: %w", req.Version, err)
				}

				// CRITICAL FIX: Explicitly clear the ID from the deleted version's in-memory representation
				logger.Debug().
					Int("deleted_version_id", existingVersion.ID()).
					Str("deleted_version", existingVersion.Version().String()).
					Msg("Clearing ID from deleted version to prevent inheritance (auto-publish mode)")

				// Clear the ID from the deleted version to ensure new versions start with ID=0
				existingVersion.ResetID()

				// Remove the existing version from the module provider's in-memory list to avoid conflicts
				// Filter out the existing version and set the updated list
				updatedVersions := make([]*moduleModel.ModuleVersion, 0)
				for _, v := range moduleProvider.GetAllVersions() {
					if v.ID() != existingVersion.ID() { // Note: This should be 0 after ResetID
						updatedVersions = append(updatedVersions, v)
					}
				}
				moduleProvider.SetVersions(updatedVersions)
			}
		}

		// Create module version entity with module provider association
		moduleVersion, err := moduleModel.NewModuleVersion(req.Version, nil, false)
		if err != nil {
			return fmt.Errorf("failed to create module version: %w", err)
		}

		// Add the version to the module provider aggregate (this sets the parent relationship)
		if err := moduleProvider.AddVersion(moduleVersion); err != nil {
			return fmt.Errorf("failed to add version to module provider: %w", err)
		}

		// Debug: Verify the module provider relationship was established and confirm new record creation
		logger := zerolog.Ctx(ctx)
		logger.Info().
			Int("module_provider_id", moduleProvider.ID()).
			Int("module_version_id_before_save", moduleVersion.ID()).
			Bool("is_new_record", moduleVersion.ID() == 0).
			Str("module_version_module_provider", func() string {
				if moduleVersion.ModuleProvider() != nil {
					return fmt.Sprintf("id=%d", moduleVersion.ModuleProvider().ID())
				}
				return "nil"
			}()).
			Msg("CRITICAL: New module version created - should have ID=0 for CREATE operation")

		// Save the module version - this creates the database row with proper module_provider_id
		savedModuleVersion, err := s.moduleVersionRepo.Save(ctx, moduleVersion)
		if err != nil {
			return fmt.Errorf("failed to create module version: %w", err)
		}

		// Verify the module version got a valid ID from the database
		if savedModuleVersion.ID() == 0 {
			return fmt.Errorf("module version was not assigned a valid ID from database")
		}

		// Set git commit SHA if provided
		if req.CommitSHA != nil {
			savedModuleVersion.SetGitInfo(req.CommitSHA, nil, false)
			// Save again to persist the SHA
			savedModuleVersion, err = s.moduleVersionRepo.Save(ctx, savedModuleVersion)
			if err != nil {
				return fmt.Errorf("failed to update module version with git SHA: %w", err)
			}
		}

		// CRITICAL SUCCESS: Log that a new record was created with a new ID
		logger.Info().
			Int("old_version_id", func() int {
				if existingVersion != nil {
					return existingVersion.ID() // Should be 0 after ResetID
				}
				return 0
			}()).
			Int("new_version_id", savedModuleVersion.ID()).
			Str("version", savedModuleVersion.Version().String()).
			Bool("is_new_id", func() bool {
				if existingVersion != nil {
					return savedModuleVersion.ID() != existingVersion.ID()
				}
				return true
			}()).
			Msg("CRITICAL SUCCESS: New module version record created in database")

		// Determine if module should be published based on reindex mode and configuration
		shouldPublish := s.shouldPublishModuleWithReindexMode(savedModuleVersion, reindexMode, existingVersion != nil)

		result = &PrepareModuleResult{
			ModuleVersion: savedModuleVersion,
			ShouldPublish: shouldPublish,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to prepare module: %w", err)
	}

	return result, nil
}

// CompleteModule completes the module creation by publishing if necessary
// This should be called after successful extraction
func (s *ModuleCreationWrapperService) CompleteModule(ctx context.Context, moduleVersion *moduleModel.ModuleVersion) error {
	// Publish the module version using the domain method
	if err := moduleVersion.Publish(); err != nil {
		return fmt.Errorf("failed to publish module version: %w", err)
	}

	// Save the updated module version
	_, err := s.moduleVersionRepo.Save(ctx, moduleVersion)
	if err != nil {
		return fmt.Errorf("failed to save published module version: %w", err)
	}

	return nil
}

// WithModuleCreationWrapper provides a context manager pattern similar to Python
// It prepares the module, executes the extraction function, and completes or rolls back
// This method now uses the unified transaction API for automatic transaction management
func (s *ModuleCreationWrapperService) WithModuleCreationWrapper(
	ctx context.Context,
	req PrepareModuleRequest,
	extractionFunc func(ctx context.Context, moduleVersion *moduleModel.ModuleVersion) error,
) error {
	// Wrap the entire operation in a single transaction
	return s.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Prepare the module within the transaction
		prepareResult, err := s.PrepareModule(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to prepare module: %w", err)
		}

		// Execute the extraction function
		if err := extractionFunc(ctx, prepareResult.ModuleVersion); err != nil {
			return fmt.Errorf("extraction failed: %w", err)
		}

		// Complete the module (publish if necessary)
		if prepareResult.ShouldPublish {
			if err := s.CompleteModule(ctx, prepareResult.ModuleVersion); err != nil {
				return fmt.Errorf("failed to complete module: %w", err)
			}
		}

		return nil
	})
}

// shouldPublishModule determines if a module should be published based on configuration
// This matches the logic in Python's prepare_module()
func (s *ModuleCreationWrapperService) shouldPublishModule(moduleVersion *moduleModel.ModuleVersion) bool {
	// TODO: Implement the logic to determine if module should be published
	// For now, return false - this should be based on:
	// - Whether the module is replacing a previously published module
	// - Whether auto-publish is enabled
	// - Other configuration options
	return false
}

// shouldPublishModuleWithReindexMode determines publishing logic based on reindex mode
func (s *ModuleCreationWrapperService) shouldPublishModuleWithReindexMode(
	moduleVersion *moduleModel.ModuleVersion,
	reindexMode configModel.ModuleVersionReindexMode,
	hadExistingVersion bool,
) bool {
	switch reindexMode {
	case configModel.ModuleVersionReindexModeLegacy:
		// In legacy mode, always start unpublished (unless auto-publish is enabled)
		return s.domainConfig.AutoPublishModuleVersions
	case configModel.ModuleVersionReindexModeAutoPublish:
		// In auto-publish mode, preserve the published state if there was an existing version
		if hadExistingVersion {
			// For now, we'll need to check the previous version's published state
			// This could be improved by caching the previous state before deletion
			return s.domainConfig.AutoPublishModuleVersions
		}
		return false
	case configModel.ModuleVersionReindexModeProhibit:
		// This mode shouldn't reach here as we would have returned an error earlier
		return false
	default:
		return false
	}
}
