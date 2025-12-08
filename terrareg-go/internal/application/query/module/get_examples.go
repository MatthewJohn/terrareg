package module

import (
	"context"
	"fmt"
	"path/filepath"

	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
)

// GetExamplesQuery retrieves examples for a module version
type GetExamplesQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	moduleParser       service.ModuleParser
	infraConfig        *infraConfig.InfrastructureConfig
}

// NewGetExamplesQuery creates a new query
func NewGetExamplesQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleParser service.ModuleParser,
	infraConfig *infraConfig.InfrastructureConfig,
) *GetExamplesQuery {
	return &GetExamplesQuery{
		moduleProviderRepo: moduleProviderRepo,
		moduleParser:       moduleParser,
		infraConfig:        infraConfig,
	}
}

// ExampleInfo represents information about an example
type ExampleInfo struct {
	Path string `json:"path"`
}

// Execute retrieves examples for a module version
func (q *GetExamplesQuery) Execute(ctx context.Context, namespace, module, provider, version string) ([]ExampleInfo, error) {
	// Find the module provider
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx, namespace, module, provider,
	)
	if err != nil {
		return nil, fmt.Errorf("module provider not found: %w", err)
	}

	// Find the version
	moduleVersion, err := moduleProvider.GetVersion(version)
	if err != nil {
		return nil, fmt.Errorf("module version not found: %w", err)
	}

	// Check if version is published
	if !moduleVersion.IsPublished() {
		return nil, fmt.Errorf("module version is not published")
	}

	// Get the module directory
	moduleDir := filepath.Join(q.infraConfig.DataDirectory, "modules", namespace, module, provider, version)

	// Use the parser to detect examples
	examplePaths, err := q.moduleParser.DetectExamples(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to detect examples: %w", err)
	}

	// Convert to ExampleInfo
	result := make([]ExampleInfo, len(examplePaths))
	for i, path := range examplePaths {
		result[i] = ExampleInfo{Path: path}
	}

	return result, nil
}
