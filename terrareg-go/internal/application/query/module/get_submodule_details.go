package module

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	apperrors "github.com/matthewjohn/terrareg/terrareg-go/internal/application/errors"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
)

// GetSubmoduleDetailsQuery retrieves details for a specific submodule
type GetSubmoduleDetailsQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	moduleVersionRepo  repository.ModuleVersionRepository
	urlService         *service.URLService
}

// NewGetSubmoduleDetailsQuery creates a new query
func NewGetSubmoduleDetailsQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
	urlService *service.URLService,
) *GetSubmoduleDetailsQuery {
	return &GetSubmoduleDetailsQuery{
		moduleProviderRepo: moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
		urlService:         urlService,
	}
}

// SubmoduleDetails represents submodule details
// Python reference: /app/terrareg/models.py BaseSubmodule.get_terrareg_api_details()
type SubmoduleDetails struct {
	Path                       string               `json:"path"`
	Readme                     string               `json:"readme"`
	Empty                      bool                 `json:"empty"`
	Inputs                     []Input              `json:"inputs"`
	Outputs                    []Output             `json:"outputs"`
	Dependencies               []Dependency         `json:"dependencies"`
	ProviderDependencies       []ProviderDependency `json:"provider_dependencies"`
	Resources                  []Resource           `json:"resources"`
	Modules                    []Module             `json:"modules"`
	DisplaySourceURL           string               `json:"display_source_url,omitempty"`
	SecurityFailures           int                  `json:"security_failures"`
	SecurityResults            []SecurityResult     `json:"security_results,omitempty"`
	GraphURL                   string               `json:"graph_url,omitempty"`
	UsageExample               string               `json:"usage_example,omitempty"`
	TerraformVersionConstraint *string              `json:"terraform_version_constraint,omitempty"`
}

// Input represents a terraform input variable
type Input struct {
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Description    *string     `json:"description,omitempty"`
	Required       bool        `json:"required"`
	Default        interface{} `json:"default,omitempty"`
	AdditionalHelp *string     `json:"additional_help,omitempty"`
	QuoteValue     bool        `json:"quote_value,omitempty"`
}

// Output represents a terraform output
type Output struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Type        *string `json:"type,omitempty"`
}

// Dependency represents a terraform module dependency
type Dependency struct {
	Module  string `json:"module"`
	Source  string `json:"source"`
	Version string `json:"version,omitempty"`
}

// ProviderDependency represents a terraform provider dependency
// Python reference: /app/terrareg/models.py BaseSubmodule.get_terraform_provider_dependencies()
type ProviderDependency struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Source    string `json:"source,omitempty"`
	Version   string `json:"version,omitempty"`
}

// Resource represents a terraform resource
type Resource struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// Module represents a terraform module call
type Module struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Version string `json:"version,omitempty"`
}

// SecurityResult represents a tfsec security result
type SecurityResult struct {
	RuleID      string `json:"rule_id"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
}

// Execute retrieves submodule details
// Python reference: /app/terrareg/models.py BaseSubmodule.get_terrareg_api_details()
func (q *GetSubmoduleDetailsQuery) Execute(ctx context.Context, namespace types.NamespaceName, moduleName types.ModuleName, provider types.ModuleProviderName, version types.ModuleVersion, path, requestDomain string) (*SubmoduleDetails, error) {
	// Get module provider first
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, moduleName, provider)
	if err != nil {
		return nil, err
	}

	if moduleProvider == nil {
		return nil, apperrors.ErrModuleProviderNotFound
	}

	// Get module version from the provider
	moduleVersion, err := q.moduleVersionRepo.FindByModuleProviderAndVersion(ctx, moduleProvider.ID(), version)
	if err != nil {
		return nil, err
	}

	if moduleVersion == nil {
		return nil, apperrors.ErrModuleVersionNotFound
	}

	// Check if version is published
	if !moduleVersion.IsPublished() {
		return nil, apperrors.ErrModuleVersionNotPublished
	}

	// Get submodule by path
	submodule := moduleVersion.GetSubmoduleByPath(path)
	if submodule == nil {
		return nil, apperrors.WrapNotFound(apperrors.ErrSubmoduleNotFound, path)
	}

	// Convert submodule to module specs
	specs := moduleVersion.ConvertSubmoduleToSpecs(submodule)
	if specs == nil {
		// Return empty details if no specs available
		return &SubmoduleDetails{
			Path:                 path,
			Readme:               "",
			Empty:                true,
			Inputs:               []Input{},
			Outputs:              []Output{},
			Dependencies:         []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources:            []Resource{},
			Modules:              []Module{},
			SecurityResults:      []SecurityResult{},
		}, nil
	}

	// Get security results
	securityResults := q.getSecurityResults(submodule)
	securityFailures := len(securityResults)

	// Generate additional fields
	graphURL := fmt.Sprintf("/modules/%d/graph/submodule/%s", moduleVersion.ID(), path)
	displaySourceURL := q.getDisplaySourceURL(moduleVersion, submodule)
	usageExample := q.getUsageExample(moduleVersion, submodule, requestDomain)

	// Get terraform version constraint from submodule details if defined
	var terraformVersionConstraint *string
	if submodule.Details() != nil && submodule.Details().HasTerraformVersionConstraint() {
		constraint := string(submodule.Details().TerraformVersion())
		terraformVersionConstraint = &constraint
	}

	return &SubmoduleDetails{
		Path:                       specs.Path,
		Readme:                     specs.Readme,
		Empty:                      specs.Empty,
		Inputs:                     convertInputs(specs.Inputs),
		Outputs:                    convertOutputs(specs.Outputs),
		Dependencies:               convertDependencies(specs.Dependencies),
		ProviderDependencies:       convertProviderDependencies(specs.ProviderDependencies),
		Resources:                  convertResources(specs.Resources),
		Modules:                    convertModules(specs.Modules),
		DisplaySourceURL:           displaySourceURL,
		SecurityFailures:           securityFailures,
		SecurityResults:            securityResults,
		GraphURL:                   graphURL,
		UsageExample:               usageExample,
		TerraformVersionConstraint: terraformVersionConstraint,
	}, nil
}

// getSecurityResults extracts tfsec results from submodule details
func (q *GetSubmoduleDetailsQuery) getSecurityResults(submodule *model.Submodule) []SecurityResult {
	details := submodule.Details()
	if details == nil || !details.HasTfsec() {
		return []SecurityResult{}
	}

	var tfsecData map[string]interface{}
	if err := json.Unmarshal(details.Tfsec(), &tfsecData); err != nil {
		return []SecurityResult{}
	}

	// Parse tfsec results - the structure is an array of results
	results, ok := tfsecData["results"].([]interface{})
	if !ok {
		return []SecurityResult{}
	}

	var securityResults []SecurityResult
	for _, result := range results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		securityResult := SecurityResult{
			RuleID:      getStringValue(resultMap, "rule_id"),
			Severity:    getStringValue(resultMap, "severity"),
			Title:       getStringValue(resultMap, "title"),
			Description: getStringValue(resultMap, "description"),
		}

		if location, ok := resultMap["location"].(map[string]interface{}); ok {
			securityResult.Location = getStringValue(location, "filename")
		}

		securityResults = append(securityResults, securityResult)
	}

	return securityResults
}

// getDisplaySourceURL returns the source browse URL for the submodule
func (q *GetSubmoduleDetailsQuery) getDisplaySourceURL(moduleVersion *model.ModuleVersion, submodule *model.Submodule) string {
	// Get source browse URL from module version with submodule path
	return moduleVersion.GetSourceBrowseURL(submodule.Path())
}

// getUsageExample returns a usage example for the submodule
// Python reference: /app/terrareg/models.py BaseSubmodule.get_usage_example()
func (q *GetSubmoduleDetailsQuery) getUsageExample(moduleVersion *model.ModuleVersion, submodule *model.Submodule, requestDomain string) string {
	// Get module name from module provider
	moduleName := types.ModuleName("")
	if moduleVersion.ModuleProvider() != nil {
		moduleName = moduleVersion.ModuleProvider().Module()
	}
	if moduleName == "" {
		moduleName = types.ModuleName(submodule.Path())
	}

	// Build terraform source URL using URL service
	var sourceURL string
	if moduleVersion.ModuleProvider() != nil {
		providerID := string(moduleVersion.ModuleProvider().FrontendID())
		version := moduleVersion.Version().String()
		sourceURL = q.urlService.BuildTerraformSourceURL(providerID, version, submodule.Path(), requestDomain)
	} else {
		sourceURL = submodule.Path()
	}

	// Build terraform block with source and optional version
	// Python: For HTTPS, version is added as a separate attribute
	// For HTTP, version is embedded in the URL (so the sourceURL contains it)
	result := fmt.Sprintf("module \"%s\" {\n  source = \"%s\"", string(moduleName), sourceURL)

	// Check if version is already in the source URL (HTTP mode)
	// HTTP URL format: http://domain/modules/provider/{version}
	// HTTPS URL format: domain/provider (no version)
	version := moduleVersion.Version().String()
	if !strings.Contains(sourceURL, "/"+version) {
		// Version is not in URL (HTTPS mode), add it as a separate attribute
		result += fmt.Sprintf("\n  version = \"%s\"", version)
	}

	result += "\n}"

	return result
}

// Helper functions to convert domain model types to DTO types
func convertInputs(inputs []model.Input) []Input {
	if inputs == nil {
		return []Input{}
	}
	result := make([]Input, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, Input{
			Name:           input.Name,
			Type:           input.Type,
			Description:    input.Description,
			Required:       input.Required,
			Default:        input.Default,
			AdditionalHelp: input.AdditionalHelp,
			QuoteValue:     input.QuoteValue,
		})
	}
	return result
}

func convertOutputs(outputs []model.Output) []Output {
	if outputs == nil {
		return []Output{}
	}
	result := make([]Output, 0, len(outputs))
	for _, output := range outputs {
		result = append(result, Output{
			Name:        output.Name,
			Description: output.Description,
			Type:        output.Type,
		})
	}
	return result
}

func convertDependencies(dependencies []model.Dependency) []Dependency {
	if dependencies == nil {
		return []Dependency{}
	}
	result := make([]Dependency, 0, len(dependencies))
	for _, dep := range dependencies {
		result = append(result, Dependency{
			Module:  dep.Module,
			Source:  dep.Source,
			Version: dep.Version,
		})
	}
	return result
}

func convertProviderDependencies(providerDeps []model.ProviderDependency) []ProviderDependency {
	if providerDeps == nil {
		return []ProviderDependency{}
	}
	result := make([]ProviderDependency, 0, len(providerDeps))
	for _, dep := range providerDeps {
		result = append(result, ProviderDependency{
			Name:      dep.Name,
			Namespace: dep.Namespace,
			Source:    dep.Source,
			Version:   dep.Version,
		})
	}
	return result
}

func convertResources(resources []model.Resource) []Resource {
	if resources == nil {
		return []Resource{}
	}
	result := make([]Resource, 0, len(resources))
	for _, res := range resources {
		result = append(result, Resource{
			Type: res.Type,
			Name: res.Name,
		})
	}
	return result
}

func convertModules(modules []model.Module) []Module {
	if modules == nil {
		return []Module{}
	}
	result := make([]Module, 0, len(modules))
	for _, mod := range modules {
		result = append(result, Module{
			Name:    mod.Name,
			Source:  mod.Source,
			Version: mod.Version,
		})
	}
	return result
}

func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
