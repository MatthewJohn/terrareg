package module

import (
	"context"
	"errors"
	"fmt"

	configmodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	moduleModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// GetModuleDownloadQuery handles retrieving download information for a module version
type GetModuleDownloadQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	domainConfig       *configmodel.DomainConfig
	gitURLBuilder      *service.GitURLBuilderService
}

// NewGetModuleDownloadQuery creates a new get module download query
func NewGetModuleDownloadQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	domainConfig *configmodel.DomainConfig,
	gitURLBuilder *service.GitURLBuilderService,
) *GetModuleDownloadQuery {
	return &GetModuleDownloadQuery{
		moduleProviderRepo: moduleProviderRepo,
		domainConfig:       domainConfig,
		gitURLBuilder:      gitURLBuilder,
	}
}

// DownloadInfo represents the download information for a module version
type DownloadInfo struct {
	ModuleProvider *moduleModel.ModuleProvider
	Version        *moduleModel.ModuleVersion

	// Git URL information
	GitURL      string // Git clone URL (if available)
	BuiltInURL  string // Built-in hosting URL (if available)
	HostingMode configmodel.ModuleHostingMode
	GitPath     string // Git path for subdirectory
	GitRef      string // Git reference (SHA or tag)
}

// Execute executes the query
// If version is empty string, returns the latest version
func (q *GetModuleDownloadQuery) Execute(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName, version types.ModuleVersion) (*DownloadInfo, error) {
	// First get the module provider
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, module, provider)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, fmt.Errorf("module provider %s/%s/%s not found", namespace, module, provider)
		}
		return nil, fmt.Errorf("failed to get module provider: %w", err)
	}

	var moduleVersion *moduleModel.ModuleVersion

	// If version is specified, get that specific version
	if string(version) != "" {
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

	// Build download info with git URL and built-in URL
	downloadInfo := &DownloadInfo{
		ModuleProvider: moduleProvider,
		Version:        moduleVersion,
		HostingMode:    q.domainConfig.AllowModuleHosting,
		GitPath:        "", // Will be populated by GetSourceDownloadURL
		GitRef:         "", // Will be populated by GetSourceDownloadURL
	}

	// Get git clone URL (if available)
	gitCloneURL := moduleVersion.GetGitCloneURL(ctx, q.domainConfig, q.gitURLBuilder)
	downloadInfo.GitURL = gitCloneURL

	// Get built-in hosting URL (this would typically be generated in the handler)
	// For now, we'll leave it empty and let the handler build it
	downloadInfo.BuiltInURL = ""

	// Populate git path and ref if available
	if moduleVersion.GitPath() != nil {
		downloadInfo.GitPath = *moduleVersion.GitPath()
	}
	if moduleVersion.GitSHA() != nil && *moduleVersion.GitSHA() != "" {
		downloadInfo.GitRef = *moduleVersion.GitSHA()
	}

	return downloadInfo, nil
}
