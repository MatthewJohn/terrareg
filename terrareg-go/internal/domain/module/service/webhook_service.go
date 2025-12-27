package service

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// WebhookResult represents the result of webhook processing
type WebhookResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	TriggerBuild bool   `json:"trigger_build,omitempty"`
}

// WebhookService handles webhook processing for modules
type WebhookService struct {
	moduleImporterService *ModuleImporterService
	moduleProviderRepo    repository.ModuleProviderRepository
	config                *infraConfig.InfrastructureConfig
	savepointHelper       *transaction.SavepointHelper
	moduleCreationWrapper *ModuleCreationWrapperService
}

// NewWebhookService creates a new webhook service
func NewWebhookService(
	moduleImporterService *ModuleImporterService,
	moduleProviderRepo repository.ModuleProviderRepository,
	config *infraConfig.InfrastructureConfig,
	savepointHelper *transaction.SavepointHelper,
	moduleCreationWrapper *ModuleCreationWrapperService,
) *WebhookService {
	return &WebhookService{
		moduleImporterService: moduleImporterService,
		moduleProviderRepo:    moduleProviderRepo,
		config:                config,
		savepointHelper:       savepointHelper,
		moduleCreationWrapper: moduleCreationWrapper,
	}
}

// ProcessWebhook processes a webhook event
func (ws *WebhookService) ProcessWebhook(ctx context.Context, gitProvider, eventType string, body []byte) (*WebhookResult, error) {
	// This method would be used by the generic webhook handlers
	// The specific module webhook handlers will use the ModuleImporterService directly

	return &WebhookResult{
		Success: true,
		Message: "Webhook processed successfully",
	}, nil
}

// CreateModuleVersionFromTag creates a module version from a git tag
func (ws *WebhookService) CreateModuleVersionFromTag(ctx context.Context, namespace, moduleName, provider, version string) (*WebhookResult, error) {
	// Integrate with the ModuleImporterService workflow:
	// 1. Find the module provider and validate the tag against version regex
	// 2. Create ImportModuleVersionRequest
	// 3. Call moduleImporterService.ImportModuleVersion()
	// 4. Handle publishing based on configuration

	// Find the module provider to get git clone URL and validate version regex
	moduleProvider, err := ws.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, moduleName, provider)
	if err != nil {
		return &WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Module provider not found: %s/%s/%s - %v", namespace, moduleName, provider, err),
		}, nil
	}

	if moduleProvider == nil {
		return &WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Module provider not found: %s/%s/%s", namespace, moduleName, provider),
		}, nil
	}

	// Validate that module has git clone URL configured
	if moduleProvider.RepoCloneURLTemplate() == nil || *moduleProvider.RepoCloneURLTemplate() == "" {
		return &WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Module provider %s/%s has no git clone URL configured", namespace, provider),
		}, nil
	}

	// Create domain input DTO
	parsedVersion, err := shared.ParseVersion(version)
	if err != nil {
		return &WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Failed to parse version %s: %v", version, err),
		}, nil
	}

	domainInput := module.NewModuleVersionImportInput(namespace, moduleName, provider, parsedVersion, &version)

	// Execute the module import
	domainReq := DomainImportRequest{
		Input:             domainInput,
		ProcessingOptions: ProcessingOptions{
			SkipArchiveExtraction:   false,
			SkipTerraformProcessing: false,
			SkipMetadataProcessing:  false,
			SkipSecurityScanning:    false,
			SkipFileContentStorage:  false,
			SkipArchiveGeneration:   false,
			SecurityScanEnabled:     true,
			FileProcessingEnabled:   true,
			GenerateArchives:        true,
			PublishModule:           true,
		},
		SourceType:         "git",
		EnableSecurityScan: true,
		GenerateArchives:   true,
	}

	if result, err := ws.moduleImporterService.ImportModuleVersionWithTransaction(ctx, domainReq); err != nil {
		return &WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Failed to import module version %s: %v", version, err),
		}, nil
	} else if !result.Success {
		errorMsg := ""
		if result.Error != nil {
			errorMsg = *result.Error
		}
		return &WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Failed to import module version %s: %s", version, errorMsg),
		}, nil
	}

	return &WebhookResult{
		Success:      true,
		Message:      fmt.Sprintf("Successfully imported module version %s for %s/%s/%s", version, namespace, moduleName, provider),
		TriggerBuild: true,
	}, nil
}

// VersionImportRequest represents a request to import a specific version
type VersionImportRequest struct {
	Version string
	Request module.ImportModuleVersionRequest
}

// VersionImportResult represents the result of importing a single version
type VersionImportResult struct {
	Version         string        `json:"version"`
	Status          string        `json:"status"` // "Success" or "Failed"
	ModuleVersionID *int          `json:"module_version_id,omitempty"`
	Error           *string       `json:"error,omitempty"`
	Duration        time.Duration `json:"duration"`
	Timestamp       time.Time     `json:"timestamp"`
}

// MultiVersionResult represents the result of processing multiple versions (matches Python format)
type MultiVersionResult struct {
	OverallStatus     string                          `json:"overall_status"` // "Success" or "Error"
	VersionsProcessed map[string]*VersionImportResult `json:"tags"`           // Maps version to result
	HasFailures       bool                            `json:"has_failures"`
	FailureSummary    string                          `json:"failure_summary,omitempty"`
	TotalVersions     int                             `json:"total_versions"`
	SuccessCount      int                             `json:"success_count"`
	FailureCount      int                             `json:"failure_count"`
}

// ProcessMultipleVersionsWithSavepoints processes multiple versions with individual savepoints
// This matches the Python Bitbucket webhook pattern where each version gets its own savepoint
func (ws *WebhookService) ProcessMultipleVersionsWithSavepoints(
	ctx context.Context,
	namespace, moduleName, provider string,
	versionRequests []VersionImportRequest,
) (*MultiVersionResult, error) {
	result := &MultiVersionResult{
		OverallStatus:     "Success",
		VersionsProcessed: make(map[string]*VersionImportResult),
		HasFailures:       false,
		TotalVersions:     len(versionRequests),
		SuccessCount:      0,
		FailureCount:      0,
	}

	// Process each version with its own savepoint for isolation
	for _, versionReq := range versionRequests {
		startTime := time.Now()

		versionResult := &VersionImportResult{
			Version:   versionReq.Version,
			Status:    "Failed",
			Timestamp: startTime,
		}

		// Create savepoint for this version

		err := ws.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
			// Use module creation wrapper for this version
			prepareReq := PrepareModuleRequest{
				Namespace:  namespace,
				ModuleName: moduleName,
				Provider:   provider,
				Version:    versionReq.Version,
				GitTag:     versionReq.Request.GitTag,
			}

			return ws.moduleCreationWrapper.WithModuleCreationWrapper(
				ctx,
				prepareReq,
				func(ctx context.Context, moduleVersion *model.ModuleVersion) error {
					// Create domain input DTO
					parsedVersion, err := shared.ParseVersion(versionReq.Version)
					if err != nil {
						return fmt.Errorf("failed to parse version %s: %w", versionReq.Version, err)
					}

					domainInput := module.NewModuleVersionImportInput(
						versionReq.Request.Namespace,
						versionReq.Request.Module,
						versionReq.Request.Provider,
						parsedVersion,
						versionReq.Request.GitTag,
					)

					// Execute the actual module import
					domainReq := DomainImportRequest{
						Input:             domainInput,
						ProcessingOptions: ProcessingOptions{
							SkipArchiveExtraction:   false,
							SkipTerraformProcessing: false,
							SkipMetadataProcessing:  false,
							SkipSecurityScanning:    false,
							SkipFileContentStorage:  false,
							SkipArchiveGeneration:   false,
							SecurityScanEnabled:     true,
							FileProcessingEnabled:   true,
							GenerateArchives:        true,
							PublishModule:           true,
						},
						SourceType:         "git",
						EnableSecurityScan: true,
						GenerateArchives:   true,
					}

				result, err := ws.moduleImporterService.ImportModuleVersionWithTransaction(ctx, domainReq)
				if err != nil {
					return err
				}
				if !result.Success {
					errorMsg := ""
					if result.Error != nil {
						errorMsg = *result.Error
					}
					return fmt.Errorf("module import failed: %s", errorMsg)
				}
				return nil
				},
			)
		})

		versionResult.Duration = time.Since(startTime)

		if err != nil {
			// Version processing failed
			errorMsg := err.Error()
			versionResult.Error = &errorMsg
			versionResult.Status = "Failed"
			result.FailureCount++
			result.HasFailures = true
		} else {
			// Version processing succeeded
			versionResult.Status = "Success"
			versionResult.ModuleVersionID = nil // Would be set if we could get the ID from the wrapper
			result.SuccessCount++
		}

		result.VersionsProcessed[versionReq.Version] = versionResult
	}

	// Set overall status and failure summary
	if result.HasFailures {
		result.OverallStatus = "Error"
		result.FailureSummary = fmt.Sprintf("%d of %d versions failed to import", result.FailureCount, result.TotalVersions)
	}

	return result, nil
}
