package setup

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
)

// GetInitialSetupQuery handles retrieval of initial setup status
type GetInitialSetupQuery struct {
	namespaceRepo      repository.NamespaceRepository
	moduleProviderRepo repository.ModuleProviderRepository
	moduleVersionRepo  repository.ModuleVersionRepository
	urlService         *service.URLService
	config             *config.Config
}

// NewGetInitialSetupQuery creates a new GetInitialSetupQuery
func NewGetInitialSetupQuery(
	namespaceRepo repository.NamespaceRepository,
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
	urlService *service.URLService,
	config *config.Config,
) *GetInitialSetupQuery {
	return &GetInitialSetupQuery{
		namespaceRepo:      namespaceRepo,
		moduleProviderRepo: moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
		urlService:         urlService,
		config:             config,
	}
}

// InitialSetupResponse contains the initial setup status response
type InitialSetupResponse struct {
	NamespaceCreated        bool    `json:"namespace_created"`
	ModuleCreated           bool    `json:"module_created"`
	VersionIndexed          bool    `json:"version_indexed"`
	VersionPublished        bool    `json:"version_published"`
	ModuleConfiguredWithGit bool    `json:"module_configured_with_git"`
	ModuleViewURL           *string `json:"module_view_url,omitempty"`
	ModuleUploadEndpoint    *string `json:"module_upload_endpoint,omitempty"`
	ModulePublishEndpoint   *string `json:"module_publish_endpoint,omitempty"`
}

// Execute retrieves the initial setup status
func (q *GetInitialSetupQuery) Execute(ctx context.Context) (*InitialSetupResponse, error) {
	response := &InitialSetupResponse{}

	// 1. Check for namespaces (get first namespace if exists)
	namespaces, err := q.namespaceRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check for namespaces: %w", err)
	}

	if len(namespaces) == 0 {
		// No namespaces exist
		return response, nil
	}
	response.NamespaceCreated = true

	// 2. Check for module providers in first namespace
	firstNamespace := namespaces[0]
	moduleProviders, err := q.moduleProviderRepo.FindByNamespace(ctx, firstNamespace.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to check for module providers: %w", err)
	}

	if len(moduleProviders) == 0 {
		// No module providers exist
		return response, nil
	}
	response.ModuleCreated = true

	// Use first module provider for remaining checks
	firstModuleProvider := moduleProviders[0]

	// 3. Check for versions (including beta and unpublished)
	versions, err := q.moduleVersionRepo.FindByModuleProvider(ctx, firstModuleProvider.ID(), true, true)
	if err != nil {
		return nil, fmt.Errorf("failed to check for versions: %w", err)
	}

	if len(versions) == 0 {
		// No versions exist
		return q.setURLs(response, firstModuleProvider), nil
	}
	response.VersionIndexed = true

	// 4. Check if any version is published
	for _, version := range versions {
		if version.IsPublished() {
			response.VersionPublished = true
			break
		}
	}

	// 5. Check git configuration
	if firstModuleProvider.GetGitCloneURL() != nil {
		response.ModuleConfiguredWithGit = true
	}

	return q.setURLs(response, firstModuleProvider), nil
}

// setURLs sets the URL fields in the response based on the module provider and configuration
func (q *GetInitialSetupQuery) setURLs(response *InitialSetupResponse, moduleProvider *model.ModuleProvider) *InitialSetupResponse {
	// Set module view URL
	viewURL := q.urlService.BuildURL(moduleProvider.GetViewURL(), nil)
	response.ModuleViewURL = &viewURL

	// Set upload endpoint if module hosting is allowed
	if q.config.AllowModuleHosting {
		uploadEndpoint := q.urlService.BuildURL(moduleProvider.GetUploadEndpoint(), nil)
		response.ModuleUploadEndpoint = &uploadEndpoint
	}

	// Set publish endpoint
	publishEndpoint := q.urlService.BuildURL(moduleProvider.GetPublishEndpoint(), nil)
	response.ModulePublishEndpoint = &publishEndpoint

	return response
}
