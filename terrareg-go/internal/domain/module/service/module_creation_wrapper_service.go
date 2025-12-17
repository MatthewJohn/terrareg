package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// ModuleCreationWrapperService handles module creation with prepare/extract/publish pattern
// This matches the Python module_create_extraction_wrapper context manager
type ModuleCreationWrapperService struct {
	moduleVersionRepo repository.ModuleVersionRepository
	savepointHelper   *transaction.SavepointHelper
}

// NewModuleCreationWrapperService creates a new module creation wrapper service
func NewModuleCreationWrapperService(
	moduleVersionRepo repository.ModuleVersionRepository,
	savepointHelper *transaction.SavepointHelper,
) *ModuleCreationWrapperService {
	return &ModuleCreationWrapperService{
		moduleVersionRepo: moduleVersionRepo,
		savepointHelper:   savepointHelper,
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
	ModuleProviderID *int
}

// PrepareModuleResult represents the result of module preparation
type PrepareModuleResult struct {
	ModuleVersion *model.ModuleVersion
	ShouldPublish bool
}

// PrepareModule creates the database row for the module version and returns whether it should be published
// This matches the Python prepare_module() method
func (s *ModuleCreationWrapperService) PrepareModule(ctx context.Context, req PrepareModuleRequest) (*PrepareModuleResult, error) {
	// Create a savepoint for the preparation

	var result *PrepareModuleResult
	err := s.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Create module version entity
		moduleVersion, err := model.NewModuleVersion(req.Version, nil, false)
		if err != nil {
			return fmt.Errorf("failed to create module version: %w", err)
		}

		// Save the module version - this creates the database row
		if err := s.moduleVersionRepo.Save(ctx, moduleVersion); err != nil {
			return fmt.Errorf("failed to create module version: %w", err)
		}

		// Determine if module should be published
		shouldPublish := s.shouldPublishModule(moduleVersion)

		result = &PrepareModuleResult{
			ModuleVersion: moduleVersion,
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
func (s *ModuleCreationWrapperService) CompleteModule(ctx context.Context, moduleVersion *model.ModuleVersion) error {
	// Publish the module version using the domain method
	if err := moduleVersion.Publish(); err != nil {
		return fmt.Errorf("failed to publish module version: %w", err)
	}

	// Save the updated module version
	if err := s.moduleVersionRepo.Save(ctx, moduleVersion); err != nil {
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
	extractionFunc func(ctx context.Context, moduleVersion *model.ModuleVersion) error,
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
func (s *ModuleCreationWrapperService) shouldPublishModule(moduleVersion *model.ModuleVersion) bool {
	// TODO: Implement the logic to determine if module should be published
	// For now, return false - this should be based on:
	// - Whether the module is replacing a previously published module
	// - Whether auto-publish is enabled
	// - Other configuration options
	return false
}
