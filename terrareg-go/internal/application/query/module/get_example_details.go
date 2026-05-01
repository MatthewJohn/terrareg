package module

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	apperrors "github.com/matthewjohn/terrareg/terrareg-go/internal/application/errors"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
)

// GetExampleDetailsQuery retrieves details for a specific example
type GetExampleDetailsQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
	moduleVersionRepo  repository.ModuleVersionRepository
	urlService         *service.URLService
}

// NewGetExampleDetailsQuery creates a new query
func NewGetExampleDetailsQuery(
	moduleProviderRepo repository.ModuleProviderRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
	urlService *service.URLService,
) *GetExampleDetailsQuery {
	return &GetExampleDetailsQuery{
		moduleProviderRepo: moduleProviderRepo,
		moduleVersionRepo:  moduleVersionRepo,
		urlService:         urlService,
	}
}

// ExampleDetails represents example details
// Python reference: /app/terrareg/models.py Example.get_terrareg_api_details()
type ExampleDetails struct {
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
	CostAnalysis               *CostAnalysis        `json:"cost_analysis,omitempty"`
}

// CostAnalysis represents infracost cost analysis for an example
// Python reference: /app/terrareg/models.py Example.get_terrareg_api_details()
type CostAnalysis struct {
	YearlyCost *string `json:"yearly_cost,omitempty"`
}

// Execute retrieves example details
// Python reference: /app/terrareg/models.py Example.get_terrareg_api_details()
func (q *GetExampleDetailsQuery) Execute(ctx context.Context, namespace types.NamespaceName, moduleName types.ModuleName, provider types.ModuleProviderName, version types.ModuleVersion, path, requestDomain string) (*ExampleDetails, error) {
	// Get module provider first
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, moduleName, provider)
	if err != nil {
		return nil, err
	}

	if moduleProvider == nil {
		return nil, apperrors.ErrModuleProviderNotFound
	}

	// Get module version from the provider
	// Handle "latest" version - similar to Python's behavior:
	// Python: if version == 'latest': module_version = module_provider.get_latest_version()
	// Python reference: /app/terrareg/server/__init__.py line 994-997
	var moduleVersion *model.ModuleVersion
	if version == "latest" || version == "" {
		// Get the latest version from the module provider
		moduleVersion = moduleProvider.GetLatestVersion()
		if moduleVersion == nil {
			return nil, apperrors.ErrModuleVersionNotFound
		}
	} else {
		// Get specific version
		moduleVersion, err = q.moduleVersionRepo.FindByModuleProviderAndVersion(ctx, moduleProvider.ID(), version)
		if err != nil {
			return nil, err
		}

		if moduleVersion == nil {
			return nil, apperrors.ErrModuleVersionNotFound
		}
	}

	// Check if version is published
	// Python reference: /app/terrareg/server/__init__.py - checks module_version.published
	if !moduleVersion.IsPublished() {
		return nil, apperrors.ErrModuleVersionNotPublished
	}

	// Get example by path
	example := moduleVersion.GetExampleByPath(path)
	if example == nil {
		return nil, apperrors.WrapNotFound(apperrors.ErrExampleNotFound, path)
	}

	// Convert example to module specs
	specs := moduleVersion.ConvertExampleToSpecs(example)
	if specs == nil {
		// Return empty details if no specs available
		// TODO return nil with an error
		return &ExampleDetails{
			Path:   path,
			Readme: "",
			Empty:  true,
		}, nil
	}

	// Get security results
	securityResults := q.getSecurityResults(example)
	securityFailures := len(securityResults)

	// Get cost analysis from infracost
	costAnalysis := q.getCostAnalysis(example)

	// Generate additional fields
	graphURL := fmt.Sprintf("/modules/%d/graph/example/%s", moduleVersion.ID(), path)
	displaySourceURL := moduleVersion.GetSourceBrowseURL(example.Path())
	usageExample := q.getUsageExample(moduleVersion, example, requestDomain)

	// Get terraform version constraint from example details if defined
	var terraformVersionConstraint *string
	if example.Details() != nil && example.Details().HasTerraformVersionConstraint() {
		constraint := string(example.Details().TerraformVersion())
		terraformVersionConstraint = &constraint
	}

	return &ExampleDetails{
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
		CostAnalysis:               costAnalysis,
	}, nil
}

// getSecurityResults extracts tfsec results from example details
func (q *GetExampleDetailsQuery) getSecurityResults(example *model.Example) []SecurityResult {
	details := example.Details()
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

// getCostAnalysis extracts cost analysis from example details
// Python reference: /app/terrareg/models.py Example.get_terrareg_api_details()
func (q *GetExampleDetailsQuery) getCostAnalysis(example *model.Example) *CostAnalysis {
	details := example.Details()
	if details == nil || !details.HasInfracost() {
		return nil
	}

	var infracostData map[string]interface{}
	if err := json.Unmarshal(details.Infracost(), &infracostData); err != nil {
		return nil
	}

	// Extract total_monthly_cost from infracost JSON (snake_case format)
	totalMonthlyCost, ok := infracostData["total_monthly_cost"].(float64)
	if !ok {
		return nil
	}

	// Calculate yearly cost (monthly * 12) and format to 2 decimal places
	yearlyCost := fmt.Sprintf("%.2f", totalMonthlyCost*12)

	return &CostAnalysis{
		YearlyCost: &yearlyCost,
	}
}

// getUsageExample returns a usage example for the example
// Python reference: /app/terrareg/models.py BaseSubmodule.get_usage_example()
func (q *GetExampleDetailsQuery) getUsageExample(moduleVersion *model.ModuleVersion, example *model.Example, requestDomain string) string {
	// Get module name from module provider
	moduleProvider := moduleVersion.ModuleProvider()
	moduleName := moduleProvider.Module()
	if moduleName == "" {
		moduleName = types.ModuleName(example.Path())
	}

	// Build terraform source URL using URL service
	providerID := string(moduleProvider.FrontendID())
	version := moduleVersion.Version().String()
	sourceURL := q.urlService.BuildTerraformSourceURL(providerID, version, example.Path(), requestDomain)

	// Build terraform block with source and optional version
	// Python: For HTTPS, version is added as a separate attribute
	// For HTTP, version is embedded in the URL (so the sourceURL contains it)
	result := fmt.Sprintf("module \"%s\" {\n  source = \"%s\"", string(moduleName), sourceURL)

	// Check if version is already in the source URL (HTTP mode)
	// HTTP URL format: http://domain/modules/provider/{version}
	// HTTPS URL format: domain/provider (no version)
	if !strings.Contains(sourceURL, "/"+version) {
		// Version is not in URL (HTTPS mode), add it as a separate attribute
		result += fmt.Sprintf("\n  version = \"%s\"", version)
	}

	result += "\n}"

	return result
}
