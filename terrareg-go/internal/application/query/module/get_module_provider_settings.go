package module

import (
	"context"
	"fmt"

	"github.com/terrareg/terrareg/internal/domain/module/model"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
)

// GetModuleProviderSettingsQuery handles retrieving module provider settings
type GetModuleProviderSettingsQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewGetModuleProviderSettingsQuery creates a new query
func NewGetModuleProviderSettingsQuery(moduleProviderRepo repository.ModuleProviderRepository) *GetModuleProviderSettingsQuery {
	return &GetModuleProviderSettingsQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// ModuleProviderSettings represents detailed module provider settings
type ModuleProviderSettings struct {
	ModuleProvider        *model.ModuleProvider
	GitProviderID         *int
	RepoBaseURLTemplate   *string
	RepoCloneURLTemplate  *string
	RepoBrowseURLTemplate *string
	GitTagFormat          *string
	GitPath               *string
	ArchiveGitPath        bool
	Verified              bool
}

// Execute retrieves the module provider settings
func (q *GetModuleProviderSettingsQuery) Execute(ctx context.Context, namespace, module, provider string) (*ModuleProviderSettings, error) {
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, module, provider)
	if err != nil {
		return nil, fmt.Errorf("module provider not found: %w", err)
	}

	return &ModuleProviderSettings{
		ModuleProvider:        moduleProvider,
		GitProviderID:         moduleProvider.GitProviderID(),
		RepoBaseURLTemplate:   moduleProvider.RepoBaseURLTemplate(),
		RepoCloneURLTemplate:  moduleProvider.RepoCloneURLTemplate(),
		RepoBrowseURLTemplate: moduleProvider.RepoBrowseURLTemplate(),
		GitTagFormat:          moduleProvider.GitTagFormat(),
		GitPath:               moduleProvider.GitPath(),
		ArchiveGitPath:        moduleProvider.ArchiveGitPath(),
		Verified:              moduleProvider.IsVerified(),
	}, nil
}
