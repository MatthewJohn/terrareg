package module

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/terrareg/terrareg/internal/config"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
	"github.com/terrareg/terrareg/internal/domain/module/service"
)

// GetSubmodulesQuery retrieves submodules for a module version
type GetSubmodulesQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	moduleParser       service.ModuleParser
	config             *config.Config
}

// NewGetSubmodulesQuery creates a new query
func NewGetSubmodulesQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleParser service.ModuleParser,
	config *config.Config,
) *GetSubmodulesQuery {
	return &GetSubmodulesQuery{
		moduleProviderRepo: moduleProviderRepo,
		moduleParser:       moduleParser,
		config:             config,
	}
}

// SubmoduleInfo represents information about a submodule
type SubmoduleInfo struct {
	Path string `json:"path"`
}

// Execute retrieves submodules for a module version
func (q *GetSubmodulesQuery) Execute(ctx context.Context, namespace, module, provider, version string) ([]SubmoduleInfo, error) {
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
	moduleDir := filepath.Join(q.config.DataDirectory, "modules", namespace, module, provider, version)

	// Use the parser to detect submodules
	submodulePaths, err := q.moduleParser.DetectSubmodules(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to detect submodules: %w", err)
	}

	// Convert to SubmoduleInfo
	result := make([]SubmoduleInfo, len(submodulePaths))
	for i, path := range submodulePaths {
		result[i] = SubmoduleInfo{Path: path}
	}

	return result, nil
}
