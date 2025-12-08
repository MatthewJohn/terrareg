package presenter

import (
	"fmt"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduledto "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/terrareg"
)

// ModuleVersionPresenter converts module version domain models to DTOs
type ModuleVersionPresenter struct{}

// NewModuleVersionPresenter creates a new module version presenter
func NewModuleVersionPresenter() *ModuleVersionPresenter {
	return &ModuleVersionPresenter{}
}

// ToDTO converts a module version domain model to a DTO
func (p *ModuleVersionPresenter) ToDTO(mv *model.ModuleVersion, namespace, moduleName, provider string) moduledto.ModuleVersionResponse {
	// Build version ID in format: namespace/name/provider/version
	id := fmt.Sprintf("%s/%s/%s/%s", namespace, moduleName, provider, mv.Version().String())

	response := moduledto.ModuleVersionResponse{
		VersionBase: moduledto.VersionBase{
			ProviderBase: moduledto.ProviderBase{
				ID:        id,
				Namespace: namespace,
				Name:      moduleName,
				Provider:  provider,
				Verified:  false, // TODO: Get from module provider
				Trusted:   false, // TODO: Get from namespace service
			},
			Version:  mv.Version().String(),
			Internal: mv.IsInternal(),
		},
	}

	// Add optional fields
	if owner := mv.Owner(); owner != nil {
		response.Owner = owner
	}

	if desc := mv.Description(); desc != nil {
		response.Description = desc
	}

	if publishedAt := mv.PublishedAt(); publishedAt != nil {
		publishedAtStr := publishedAt.Format("2006-01-02T15:04:05Z")
		response.PublishedAt = &publishedAtStr
	}

	return response
}

// ToTerraregDTO converts a module version to TerraregModuleVersionResponse
func (p *ModuleVersionPresenter) ToTerraregDTO(mv *model.ModuleVersion, namespace, moduleName, provider string) moduledto.TerraregModuleVersionResponse {
	// Build version ID
	id := fmt.Sprintf("%s/%s/%s/%s", namespace, moduleName, provider, mv.Version().String())

	// Get module provider for additional details
	moduleProvider := mv.ModuleProvider()

	response := moduledto.TerraregModuleVersionResponse{
		TerraregVersionDetails: moduledto.TerraregVersionDetails{
			VersionDetails: moduledto.VersionDetails{
				VersionBase: moduledto.VersionBase{
					ProviderBase: moduledto.ProviderBase{
						ID:        id,
						Namespace: namespace,
						Name:      moduleName,
						Provider:  provider,
						Verified:  moduleProvider != nil && moduleProvider.IsVerified(),
						Trusted:   false, // TODO: Get from namespace service when implemented
					},
					Version:  mv.Version().String(),
					Internal: mv.IsInternal(),
				},
				// Add Terraform specs (this would need to be populated from the module version details)
				Root: moduledto.ModuleSpecs{
					Path:   "",
					Readme: "",
					Empty:  true,
					// Other fields would be populated from actual module data
				},
				Submodules: []moduledto.ModuleSpecs{},
				Providers:  []string{}, // Would be populated from module data
			},
			// Terrareg-specific fields
			Beta:             mv.IsBeta(),
			Published:        mv.IsPublished(),
			SecurityFailures: 0, // TODO: Implement security scanning
			// Other UI-specific fields would be populated here
		},
	}

	// Add optional fields
	if owner := mv.Owner(); owner != nil {
		response.Owner = owner
	}

	if desc := mv.Description(); desc != nil {
		response.Description = desc
	}

	if publishedAt := mv.PublishedAt(); publishedAt != nil {
		publishedAtStr := publishedAt.Format("2006-01-02T15:04:05Z")
		response.PublishedAt = &publishedAtStr
	}

	// Add module provider specific fields
	if moduleProvider != nil {
		// Get provider ID - using a field that exists
		moduleProviderID := moduleProvider.ID()
		response.ModuleProviderID = &moduleProviderID

		response.GitProviderID = moduleProvider.GitProviderID()
		response.GitTagFormat = moduleProvider.GitTagFormat()
		response.GitPath = moduleProvider.GitPath()
		response.ArchiveGitPath = moduleProvider.ArchiveGitPath()
		response.RepoBaseURLTemplate = moduleProvider.RepoBaseURLTemplate()
		response.RepoCloneURLTemplate = moduleProvider.RepoCloneURLTemplate()
		response.RepoBrowseURLTemplate = moduleProvider.RepoBrowseURLTemplate()
	}

	return response
}

// ToTerraregProviderDetailsDTO converts a module version to TerraregModuleProviderDetailsResponse.
// This replicates Python's ModuleVersion.get_terrareg_api_details() method.
func (p *ModuleVersionPresenter) ToTerraregProviderDetailsDTO(
	mv *model.ModuleVersion,
	namespace, moduleName, provider string,
	requestDomain string,
) *terrareg.TerraregModuleProviderDetailsResponse {
	// Build provider ID (without version)
	providerID := fmt.Sprintf("%s/%s/%s", namespace, moduleName, provider)

	// Get module provider for additional details
	moduleProvider := mv.ModuleProvider()

	// Start with base response structure
	response := &terrareg.TerraregModuleProviderDetailsResponse{
		// Base provider info (from ModuleProvider.get_terrareg_api_details)
		ID:        providerID,
		Namespace: namespace,
		Name:      moduleName,
		Provider:  provider,
		Verified:  moduleProvider != nil && moduleProvider.IsVerified(),
		Trusted:   false, // TODO: Get from namespace service when implemented

		// Module version details (from ModuleVersion.get_api_details)
		Owner:      mv.Owner(),
		Version:    mv.Version().String(),
		Description: mv.Description(),
		Internal:   mv.IsInternal(),

		// UI-specific terrareg fields
		Beta:      mv.IsBeta(),
		Published: mv.IsPublished(),
	}

	// Add module provider metadata
	if moduleProvider != nil {
		if moduleProviderID := moduleProvider.ID(); moduleProviderID > 0 {
			moduleProviderIDStr := fmt.Sprintf("%d", moduleProviderID)
			response.ModuleProviderID = &moduleProviderIDStr
		}
		response.GitProviderID = moduleProvider.GitProviderID()
		response.GitTagFormat = moduleProvider.GitTagFormat()
		response.GitPath = moduleProvider.GitPath()
		response.ArchiveGitPath = moduleProvider.ArchiveGitPath()
		response.RepoBaseURLTemplate = moduleProvider.RepoBaseURLTemplate()
		response.RepoCloneURLTemplate = moduleProvider.RepoCloneURLTemplate()
		response.RepoBrowseURLTemplate = moduleProvider.RepoBrowseURLTemplate()

		// TODO: Get versions list (limited list, not all versions)
		// Will implement when GetVersionsList() method is added to domain model
		response.Versions = []string{}
	}

	// TODO: Add source URL when Source() method is implemented in domain model
	// response.Source = mv.Source()

	// Add published date
	if publishedAt := mv.PublishedAt(); publishedAt != nil {
		publishedAtStr := publishedAt.Format("2006-01-02T15:04:05Z")
		response.PublishedAt = &publishedAtStr
	}

	// Add downloads count
	// TODO: Implement GetDownloads() method in domain model
	// response.Downloads = mv.GetDownloads()

	// Extract analytics token from namespace if present (namespace__token format)
	if strings.Contains(namespace, "__") {
		parts := strings.Split(namespace, "__")
		if len(parts) > 1 {
			response.AnalyticsToken = &parts[1]
		}
	}

	// TODO: Populate remaining fields with actual domain method implementations
	// These are placeholders until domain methods are implemented

	// Terraform module specifications (would come from domain methods)
	response.Root = terrareg.TerraregModuleSpecs{
		Path:   "",
		Readme: "",
		Empty:  true,
		// TODO: Populate from mv.GetRootModuleSpecs()
	}
	response.Submodules = []terrareg.TerraregModuleSpecs{} // TODO: Populate from mv.GetSubmodules()
	response.Providers = []string{} // TODO: Populate from module provider

	// UI-specific fields (would come from domain methods)
	response.UsageExample = nil // TODO: Generate from mv.GetUsageExample(requestDomain)
	response.PublishedAtDisplay = nil // TODO: Format from mv.GetPublishedAtDisplay()
	response.DisplaySourceURL = nil // TODO: Get from mv.GetDisplaySourceURL(requestDomain)
	response.GraphURL = nil // TODO: Generate from mv.GetGraphURL()

	// Security scanning (would come from domain methods)
	response.SecurityFailures = 0 // TODO: Get from mv.GetSecurityFailures()
	response.SecurityResults = []terrareg.TerraregSecurityResult{} // TODO: Get from mv.GetSecurityResults()

	// Configuration (would come from domain methods)
	response.AdditionalTabFiles = map[string]string{} // TODO: Get from mv.GetAdditionalTabFiles()
	response.CustomLinks = []terrareg.TerraregCustomLink{} // TODO: Get from mv.GetCustomLinks()
	response.TerraformExampleVersionString = nil // TODO: Get from mv.GetTerraformExampleVersionString()
	response.TerraformExampleVersionComment = []string{} // TODO: Get from mv.GetTerraformExampleVersionComment()
	response.TerraformVersionConstraint = nil // TODO: Get from module version constraints
	response.ModuleExtractionUpToDate = false // TODO: Get from mv.GetModuleExtractionUpToDate()

	return response
}
