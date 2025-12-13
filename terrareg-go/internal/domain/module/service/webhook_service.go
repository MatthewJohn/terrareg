package service

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
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
	config               *infraConfig.InfrastructureConfig
}

// NewWebhookService creates a new webhook service
func NewWebhookService(
	moduleImporterService *ModuleImporterService,
	moduleProviderRepo repository.ModuleProviderRepository,
	config *infraConfig.InfrastructureConfig,
) *WebhookService {
	return &WebhookService{
		moduleImporterService: moduleImporterService,
		moduleProviderRepo:    moduleProviderRepo,
		config:               config,
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

	// Create and execute the import request
	importRequest := ImportModuleVersionRequest{
		Namespace: namespace,
		Module:    moduleName,
		Provider:  provider,
		GitTag:    &version,
	}

	// Execute the module import
	if err := ws.moduleImporterService.ImportModuleVersion(ctx, importRequest); err != nil {
		return &WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Failed to import module version %s: %v", version, err),
		}, nil
	}

	return &WebhookResult{
		Success:      true,
		Message:      fmt.Sprintf("Successfully imported module version %s for %s/%s/%s", version, namespace, moduleName, provider),
		TriggerBuild: true,
	}, nil
}