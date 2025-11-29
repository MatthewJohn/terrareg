package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/terrareg/terrareg/internal/domain/module/model"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
	"github.com/terrareg/terrareg/internal/domain/shared"
)

// GetModuleDownloadQuery handles retrieving download information for a module version
type GetModuleDownloadQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewGetModuleDownloadQuery creates a new get module download query
func NewGetModuleDownloadQuery(moduleProviderRepo repository.ModuleProviderRepository) *GetModuleDownloadQuery {
	return &GetModuleDownloadQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// DownloadInfo represents the download information for a module version
type DownloadInfo struct {
	ModuleProvider *model.ModuleProvider
	Version        *model.ModuleVersion
}

// Execute executes the query
// If version is empty string, returns the latest version
func (q *GetModuleDownloadQuery) Execute(ctx context.Context, namespace, module, provider, version string) (*DownloadInfo, error) {
	// First get the module provider
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, module, provider)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, fmt.Errorf("module provider %s/%s/%s not found", namespace, module, provider)
		}
		return nil, fmt.Errorf("failed to get module provider: %w", err)
	}

	var moduleVersion *model.ModuleVersion

	// If version is specified, get that specific version
	if version != "" {
		moduleVersion, err = moduleProvider.GetVersion(version)
		if err != nil || moduleVersion == nil {
			return nil, fmt.Errorf("version %s not found for %s/%s/%s", version, namespace, module, provider)
		}
	} else {
		// Get the latest version
		moduleVersion = moduleProvider.GetLatestVersion()
		if moduleVersion == nil {
			return nil, fmt.Errorf("no published versions found for %s/%s/%s", namespace, module, provider)
		}
	}

	// Check if version is published
	if !moduleVersion.IsPublished() {
		return nil, fmt.Errorf("version %s is not published", moduleVersion.Version().String())
	}

	return &DownloadInfo{
		ModuleProvider: moduleProvider,
		Version:        moduleVersion,
	}, nil
}
