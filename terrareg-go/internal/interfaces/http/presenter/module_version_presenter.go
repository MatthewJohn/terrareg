package presenter

import (
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduledto "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/module"
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
