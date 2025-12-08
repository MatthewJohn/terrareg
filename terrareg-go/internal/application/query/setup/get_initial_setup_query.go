package setup

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
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
	domainConfig       *configModel.DomainConfig
}

// NewGetInitialSetupQuery creates a new GetInitialSetupQuery
func NewGetInitialSetupQuery(
	namespaceRepo repository.NamespaceRepository,
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
	urlService *service.URLService,
	domainConfig *configModel.DomainConfig,
) *GetInitialSetupQuery {
	return &GetInitialSetupQuery{
		namespaceRepo:      namespaceRepo,
		moduleProviderRepo: moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
		urlService:         urlService,
		domainConfig:       domainConfig,
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

	// 2. Check for module providers in any namespace
	var firstModuleProvider *model.ModuleProvider
	var foundNamespace string

	// Check each namespace until we find one with module providers
	for _, ns := range namespaces {
		log.Info().Str("namespace", ns.Name()).Msg("InitialSetup: checking for module providers")

		moduleProviders, err := q.moduleProviderRepo.FindByNamespace(ctx, ns.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to check for module providers: %w", err)
		}

		log.Info().Int("count", len(moduleProviders)).Msg("InitialSetup: found module providers")

		if len(moduleProviders) > 0 {
			// Found module providers
			response.ModuleCreated = true
			log.Info().Msg("InitialSetup: set ModuleCreated = true")
			firstModuleProvider = moduleProviders[0]
			foundNamespace = ns.Name()
			break
		}
	}

	// If no module providers found in any namespace
	if firstModuleProvider == nil {
		log.Info().Msg("InitialSetup: no module providers found in any namespace")
		return response, nil
	}

	log.Info().
		Str("namespace", foundNamespace).
		Int("provider_id", firstModuleProvider.ID()).
		Str("module", firstModuleProvider.Module()).
		Str("provider", firstModuleProvider.Provider()).
		Msg("InitialSetup: checking versions for provider")

	// 3. Check for versions (including beta and unpublished)
	versions, err := q.moduleVersionRepo.FindByModuleProvider(ctx, firstModuleProvider.ID(), true, true)
	if err != nil {
		return nil, fmt.Errorf("failed to check for versions: %w", err)
	}

	log.Info().Int("version_count", len(versions)).Msg("InitialSetup: found versions")

	if len(versions) > 0 {
		response.VersionIndexed = true
		log.Info().Msg("InitialSetup: set VersionIndexed = true")
	}

	// 4. Check if any version is published
	if len(versions) > 0 {
		for _, version := range versions {
			if version.IsPublished() {
				response.VersionPublished = true
				log.Info().Msg("InitialSetup: set VersionPublished = true")
				break
			}
		}
	}

	// 5. Check git configuration
	if firstModuleProvider.GetGitCloneURL() != nil {
		response.ModuleConfiguredWithGit = true
		log.Info().Msg("InitialSetup: set ModuleConfiguredWithGit = true")
	}

	log.Info().
		Bool("namespace_created", response.NamespaceCreated).
		Bool("module_created", response.ModuleCreated).
		Bool("version_indexed", response.VersionIndexed).
		Bool("version_published", response.VersionPublished).
		Msg("InitialSetup: final response before setURLs")

	return q.setURLs(response, firstModuleProvider), nil
}

// setURLs sets the URL fields in the response based on the module provider and configuration
func (q *GetInitialSetupQuery) setURLs(response *InitialSetupResponse, moduleProvider *model.ModuleProvider) *InitialSetupResponse {
	log.Info().
		Bool("module_created_in", response.ModuleCreated).
		Msg("InitialSetup: setURLs entry")

	// Set module view URL
	viewURL := q.urlService.BuildURL(moduleProvider.GetViewURL(), nil)
	response.ModuleViewURL = &viewURL

	// Set upload endpoint if module hosting is allowed
	if q.domainConfig.AllowModuleHosting != configModel.ModuleHostingModeDisallow {
		uploadEndpoint := q.urlService.BuildURL(moduleProvider.GetUploadEndpoint(), nil)
		response.ModuleUploadEndpoint = &uploadEndpoint
	}

	// Set publish endpoint
	publishEndpoint := q.urlService.BuildURL(moduleProvider.GetPublishEndpoint(), nil)
	response.ModulePublishEndpoint = &publishEndpoint

	log.Info().
		Bool("module_created_out", response.ModuleCreated).
		Msg("InitialSetup: setURLs returning")

	return response
}
