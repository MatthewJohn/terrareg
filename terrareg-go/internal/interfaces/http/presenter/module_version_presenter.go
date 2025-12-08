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
func (p *ModuleVersionPresenter) ToTerraregDTO(mv *model.ModuleVersion, namespace, moduleName, provider string) *moduledto.TerraregModuleVersionResponse {
	// Build version ID
	id := fmt.Sprintf("%s/%s/%s/%s", namespace, moduleName, provider, mv.Version().String())

	// Get module provider for additional details
	moduleProvider := mv.ModuleProvider()

	if moduleProvider == nil {
		return nil
	}

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
	// Get provider ID - using a field that exists
	response.ModuleProviderID = moduleProvider.FrontendID()

	response.GitProviderID = moduleProvider.GitProviderID()
	response.GitTagFormat = moduleProvider.GitTagFormat()
	response.GitPath = moduleProvider.GitPath()
	response.ArchiveGitPath = moduleProvider.ArchiveGitPath()
	response.RepoBaseURLTemplate = moduleProvider.RepoBaseURLTemplate()
	response.RepoCloneURLTemplate = moduleProvider.RepoCloneURLTemplate()
	response.RepoBrowseURLTemplate = moduleProvider.RepoBrowseURLTemplate()

	return &response
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
		Owner:       mv.Owner(),
		Version:     mv.Version().String(),
		Description: mv.Description(),
		Internal:    mv.IsInternal(),

		// UI-specific terrareg fields
		Beta:      mv.IsBeta(),
		Published: mv.IsPublished(),
	}

	// Add module provider metadata
	if moduleProvider != nil {
		response.ModuleProviderID = moduleProvider.FrontendID()
		response.GitProviderID = moduleProvider.GitProviderID()
		response.GitTagFormat = moduleProvider.GitTagFormat()
		response.GitPath = moduleProvider.GitPath()
		response.ArchiveGitPath = moduleProvider.ArchiveGitPath()
		response.RepoBaseURLTemplate = moduleProvider.RepoBaseURLTemplate()
		response.RepoCloneURLTemplate = moduleProvider.RepoCloneURLTemplate()
		response.RepoBrowseURLTemplate = moduleProvider.RepoBrowseURLTemplate()

		// Get versions list from module provider
		response.Versions = moduleProvider.GetVersionsList()
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
	response.Downloads = mv.GetDownloads()

	// Extract analytics token from namespace if present (namespace__token format)
	if strings.Contains(namespace, "__") {
		parts := strings.Split(namespace, "__")
		if len(parts) > 1 {
			response.AnalyticsToken = &parts[1]
		}
	}

	// TODO: Populate remaining fields with actual domain method implementations
	// These are placeholders until domain methods are implemented

	// Populate with real data from domain methods

	// Terraform module specifications
	rootSpecs := mv.GetRootModuleSpecs()
	response.Root = terrareg.TerraregModuleSpecs{
		Path:                 rootSpecs.Path,
		Readme:               rootSpecs.Readme,
		Empty:                rootSpecs.Empty,
		Inputs:               convertInputsToTerrareg(rootSpecs.Inputs),
		Outputs:              convertOutputsToTerrareg(rootSpecs.Outputs),
		Dependencies:         convertDependenciesToTerrareg(rootSpecs.Dependencies),
		ProviderDependencies: convertProviderDepsToTerrareg(rootSpecs.ProviderDependencies),
		Resources:            convertResourcesToTerrareg(rootSpecs.Resources),
		Modules:              convertModulesToTerrareg(rootSpecs.Modules),
	}

	// Convert submodules
	var submoduleSpecs []terrareg.TerraregModuleSpecs
	for _, subSpec := range mv.GetSubmodules() {
		submoduleSpecs = append(submoduleSpecs, terrareg.TerraregModuleSpecs{
			Path:                 subSpec.Path,
			Readme:               subSpec.Readme,
			Empty:                subSpec.Empty,
			Inputs:               convertInputsToTerrareg(subSpec.Inputs),
			Outputs:              convertOutputsToTerrareg(subSpec.Outputs),
			Dependencies:         convertDependenciesToTerrareg(subSpec.Dependencies),
			ProviderDependencies: convertProviderDepsToTerrareg(subSpec.ProviderDependencies),
			Resources:            convertResourcesToTerrareg(subSpec.Resources),
			Modules:              convertModulesToTerrareg(subSpec.Modules),
		})
	}
	response.Submodules = submoduleSpecs

	// Providers - TODO: Get unique providers from module version
	response.Providers = []string{}

	// UI-specific fields
	usageExample := mv.GetUsageExample(requestDomain)
	if usageExample != "" {
		response.UsageExample = &usageExample
	}

	publishedAtDisplay := mv.GetPublishedAtDisplay()
	if publishedAtDisplay != "" {
		response.PublishedAtDisplay = &publishedAtDisplay
	}

	displaySourceURL := mv.GetDisplaySourceURL(requestDomain)
	if displaySourceURL != "" {
		response.DisplaySourceURL = &displaySourceURL
	}

	graphURL := mv.GetGraphURL()
	if graphURL != "" {
		response.GraphURL = &graphURL
	}

	// Security scanning
	response.SecurityFailures = mv.GetSecurityFailures()
	securityResults := mv.GetSecurityResults()
	for _, result := range securityResults {
		response.SecurityResults = append(response.SecurityResults, terrareg.TerraregSecurityResult{
			RuleID:      result.RuleID,
			Severity:    result.Severity,
			Title:       result.Title,
			Description: result.Description,
			Location: terrareg.TerraregSecurityLocation{
				Filename:  result.Location.Filename,
				StartLine: result.Location.StartLine,
				EndLine:   result.Location.EndLine,
			},
		})
	}

	// Configuration
	response.AdditionalTabFiles = mv.GetAdditionalTabFiles()

	customLinks := mv.GetCustomLinks()
	for _, link := range customLinks {
		response.CustomLinks = append(response.CustomLinks, terrareg.TerraregCustomLink{
			Text: link.Text,
			URL:  link.URL,
		})
	}

	terraformExampleVersionString := mv.GetTerraformExampleVersionString()
	if terraformExampleVersionString != "" {
		response.TerraformExampleVersionString = &terraformExampleVersionString
	}

	response.TerraformExampleVersionComment = mv.GetTerraformExampleVersionComment()
	// TODO: Get version constraint from module version details
	response.TerraformVersionConstraint = nil
	response.ModuleExtractionUpToDate = mv.GetModuleExtractionUpToDate()

	return response
}

// Converter functions to convert domain types to terrareg DTOs

func convertInputsToTerrareg(inputs []model.Input) []terrareg.TerraregInput {
	var result []terrareg.TerraregInput
	for _, input := range inputs {
		result = append(result, terrareg.TerraregInput{
			Name:           input.Name,
			Type:           input.Type,
			Description:    input.Description,
			Required:       input.Required,
			Default:        input.Default,
			AdditionalHelp: input.AdditionalHelp,
			QuoteValue:     input.QuoteValue,
			Sensitive:      input.Sensitive,
		})
	}
	return result
}

func convertOutputsToTerrareg(outputs []model.Output) []terrareg.TerraregOutput {
	var result []terrareg.TerraregOutput
	for _, output := range outputs {
		result = append(result, terrareg.TerraregOutput{
			Name:        output.Name,
			Description: output.Description,
			Type:        output.Type,
		})
	}
	return result
}

func convertDependenciesToTerrareg(dependencies []model.Dependency) []terrareg.TerraregDependency {
	var result []terrareg.TerraregDependency
	for _, dep := range dependencies {
		result = append(result, terrareg.TerraregDependency{
			Module:  dep.Module,
			Source:  dep.Source,
			Version: dep.Version,
		})
	}
	return result
}

func convertProviderDepsToTerrareg(providerDeps []model.ProviderDependency) []terrareg.TerraregProviderDep {
	var result []terrareg.TerraregProviderDep
	for _, dep := range providerDeps {
		result = append(result, terrareg.TerraregProviderDep{
			Provider: dep.Provider,
			Source:   dep.Source,
			Version:  dep.Version,
		})
	}
	return result
}

func convertResourcesToTerrareg(resources []model.Resource) []terrareg.TerraregResource {
	var result []terrareg.TerraregResource
	for _, resource := range resources {
		result = append(result, terrareg.TerraregResource{
			Name: resource.Name,
			Type: resource.Type,
		})
	}
	return result
}

func convertModulesToTerrareg(modules []model.Module) []terrareg.TerraregModule {
	var result []terrareg.TerraregModule
	for _, module := range modules {
		result = append(result, terrareg.TerraregModule{
			Name:      module.Name,
			Source:    module.Source,
			Version:   module.Version,
			Key:       module.Key,
			Providers: module.Providers,
		})
	}
	return result
}
