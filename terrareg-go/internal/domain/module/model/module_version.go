package model

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// ModuleVersion is an entity within the ModuleProvider aggregate
type ModuleVersion struct {
	id             int
	moduleProvider *ModuleProvider // Parent aggregate
	version        *shared.Version
	details        *ModuleDetails
	beta           bool
	internal       bool
	published      bool
	publishedAt    *time.Time

	// Git information
	gitSHA         *string
	gitPath        *string
	archiveGitPath bool

	// Repository URLs (can override module provider defaults)
	repoBaseURLTemplate   *string
	repoCloneURLTemplate  *string
	repoBrowseURLTemplate *string

	// Metadata
	owner       *string
	description *string

	// Variable template
	variableTemplate []byte

	// Extraction version for tracking module extraction changes
	extractionVersion *int

	// Submodules and examples (entities)
	submodules []*Submodule
	examples   []*Example

	// Additional files
	files []*ModuleFile

	createdAt time.Time
	updatedAt time.Time
}

// NewModuleVersion creates a new module version
func NewModuleVersion(versionStr string, details *ModuleDetails, beta bool) (*ModuleVersion, error) {
	version, err := shared.ParseVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}

	now := time.Now()
	return &ModuleVersion{
		version:    version,
		details:    details,
		beta:       beta,
		internal:   false,
		published:  false,
		submodules: make([]*Submodule, 0),
		examples:   make([]*Example, 0),
		files:      make([]*ModuleFile, 0),
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

// ReconstructModuleVersion reconstructs a module version from persistence
func ReconstructModuleVersion(
	id int,
	versionStr string,
	details *ModuleDetails,
	beta, internal, published bool,
	publishedAt *time.Time,
	gitSHA, gitPath *string,
	archiveGitPath bool,
	repoBaseURLTemplate, repoCloneURLTemplate, repoBrowseURLTemplate *string,
	owner, description *string,
	variableTemplate []byte,
	extractionVersion *int,
	createdAt, updatedAt time.Time,
) (*ModuleVersion, error) {
	version, err := shared.ParseVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}

	return &ModuleVersion{
		id:                    id,
		version:               version,
		details:               details,
		beta:                  beta,
		internal:              internal,
		published:             published,
		publishedAt:           publishedAt,
		gitSHA:                gitSHA,
		gitPath:               gitPath,
		archiveGitPath:        archiveGitPath,
		repoBaseURLTemplate:   repoBaseURLTemplate,
		repoCloneURLTemplate:  repoCloneURLTemplate,
		repoBrowseURLTemplate: repoBrowseURLTemplate,
		owner:                 owner,
		description:           description,
		variableTemplate:      variableTemplate,
		extractionVersion:     extractionVersion,
		submodules:            make([]*Submodule, 0),
		examples:              make([]*Example, 0),
		files:                 make([]*ModuleFile, 0),
		createdAt:             createdAt,
		updatedAt:             updatedAt,
	}, nil
}

// Business methods

// Publish publishes this module version
func (mv *ModuleVersion) Publish() error {
	if mv.published {
		return fmt.Errorf("%w: version %s is already published", shared.ErrDomainViolation, mv.version)
	}

	now := time.Now()
	mv.published = true
	mv.publishedAt = &now
	mv.updatedAt = now

	// Update parent's latest version
	if mv.moduleProvider != nil {
		mv.moduleProvider.updateLatestVersion()
	}

	return nil
}

// Unpublish unpublishes this module version
func (mv *ModuleVersion) Unpublish() {
	mv.published = false
	mv.publishedAt = nil
	mv.updatedAt = time.Now()

	// Update parent's latest version
	if mv.moduleProvider != nil {
		mv.moduleProvider.updateLatestVersion()
	}
}

// MarkInternal marks this version as internal
func (mv *ModuleVersion) MarkInternal() {
	mv.internal = true
	mv.updatedAt = time.Now()
}

// MarkPublic marks this version as public
func (mv *ModuleVersion) MarkPublic() {
	mv.internal = false
	mv.updatedAt = time.Now()
}

// SetGitInfo sets Git information
func (mv *ModuleVersion) SetGitInfo(gitSHA, gitPath *string, archiveGitPath bool) {
	mv.gitSHA = gitSHA
	mv.gitPath = gitPath
	mv.archiveGitPath = archiveGitPath
	mv.updatedAt = time.Now()
}

// SetRepositoryURLs sets repository URLs
func (mv *ModuleVersion) SetRepositoryURLs(baseURL, cloneURL, browseURL *string) {
	mv.repoBaseURLTemplate = baseURL
	mv.repoCloneURLTemplate = cloneURL
	mv.repoBrowseURLTemplate = browseURL
	mv.updatedAt = time.Now()
}

// SetMetadata sets owner and description
func (mv *ModuleVersion) SetMetadata(owner, description *string) {
	mv.owner = owner
	mv.description = description
	mv.updatedAt = time.Now()
}

func (mv *ModuleVersion) SetDetails(details *ModuleDetails) {
	mv.details = details
	mv.updatedAt = time.Now()
}

// SetVariableTemplate sets the variable template
func (mv *ModuleVersion) SetVariableTemplate(template []byte) {
	mv.variableTemplate = template
	mv.updatedAt = time.Now()
}

// AddSubmodule adds a submodule
func (mv *ModuleVersion) AddSubmodule(submodule *Submodule) {
	mv.submodules = append(mv.submodules, submodule)
	mv.updatedAt = time.Now()
}

// AddExample adds an example
func (mv *ModuleVersion) AddExample(example *Example) {
	mv.examples = append(mv.examples, example)
	mv.updatedAt = time.Now()
}

// AddFile adds an additional file
func (mv *ModuleVersion) AddFile(file *ModuleFile) {
	mv.files = append(mv.files, file)
	mv.updatedAt = time.Now()
}

// setModuleProvider sets the parent module provider (internal use by aggregate)
func (mv *ModuleVersion) setModuleProvider(mp *ModuleProvider) {
	mv.moduleProvider = mp
}

// Getters

func (mv *ModuleVersion) ID() int {
	return mv.id
}

func (mv *ModuleVersion) ModuleProvider() *ModuleProvider {
	return mv.moduleProvider
}

func (mv *ModuleVersion) Version() *shared.Version {
	return mv.version
}

func (mv *ModuleVersion) Details() *ModuleDetails {
	return mv.details
}

func (mv *ModuleVersion) IsBeta() bool {
	return mv.beta
}

func (mv *ModuleVersion) IsInternal() bool {
	return mv.internal
}

func (mv *ModuleVersion) IsPublished() bool {
	return mv.published
}

func (mv *ModuleVersion) PublishedAt() *time.Time {
	return mv.publishedAt
}

func (mv *ModuleVersion) GitSHA() *string {
	return mv.gitSHA
}

func (mv *ModuleVersion) GitPath() *string {
	return mv.gitPath
}

func (mv *ModuleVersion) ArchiveGitPath() bool {
	return mv.archiveGitPath
}

func (mv *ModuleVersion) RepoBaseURLTemplate() *string {
	return mv.repoBaseURLTemplate
}

func (mv *ModuleVersion) RepoCloneURLTemplate() *string {
	return mv.repoCloneURLTemplate
}

func (mv *ModuleVersion) RepoBrowseURLTemplate() *string {
	return mv.repoBrowseURLTemplate
}

func (mv *ModuleVersion) Owner() *string {
	return mv.owner
}

func (mv *ModuleVersion) Description() *string {
	return mv.description
}

func (mv *ModuleVersion) VariableTemplate() []byte {
	return mv.variableTemplate
}

func (mv *ModuleVersion) ExtractionVersion() *int {
	return mv.extractionVersion
}

func (mv *ModuleVersion) Submodules() []*Submodule {
	return mv.submodules
}

func (mv *ModuleVersion) Examples() []*Example {
	return mv.examples
}

func (mv *ModuleVersion) Files() []*ModuleFile {
	return mv.files
}

func (mv *ModuleVersion) CreatedAt() time.Time {
	return mv.createdAt
}

func (mv *ModuleVersion) UpdatedAt() time.Time {
	return mv.updatedAt
}

// Additional domain methods for terrareg API functionality

// GetDownloads returns the download count for this module version
// TODO: Integrate with analytics service when available
func (mv *ModuleVersion) GetDownloads() int {
	// Placeholder until analytics integration is implemented
	return 0
}

// parseTerraformDocs is a helper method to parse terraform-docs JSON into domain models
func (mv *ModuleVersion) parseTerraformDocs(terraformDocsJSON []byte) ([]Input, []Output, []ProviderDependency, []Resource) {
	// Define struct for unmarshaling terraform-docs JSON
	type TerraformDocsJSON struct {
		Inputs []struct {
			Name        string      `json:"name"`
			Type        string      `json:"type"`
			Description string      `json:"description"`
			Default     interface{} `json:"default"`
			Required    bool        `json:"required"`
		} `json:"inputs"`
		Outputs []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"outputs"`
		Providers []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"providers"`
		Resources []struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"resources"`
	}

	var terraformDocs TerraformDocsJSON
	if err := json.Unmarshal(terraformDocsJSON, &terraformDocs); err != nil {
		// If unmarshaling fails, return empty slices
		return []Input{}, []Output{}, []ProviderDependency{}, []Resource{}
	}

	// Convert terraform docs to domain models
	inputs := make([]Input, len(terraformDocs.Inputs))
	for i, tfInput := range terraformDocs.Inputs {
		inputs[i] = Input{
			Name:           tfInput.Name,
			Type:           tfInput.Type,
			Description:    &tfInput.Description,
			Required:       tfInput.Required,
			Default:        tfInput.Default,
			AdditionalHelp: nil,   // terraform-docs doesn't provide this
			QuoteValue:     false, // terraform-docs doesn't provide this
			Sensitive:      false, // terraform-docs doesn't provide this
		}
	}

	outputs := make([]Output, len(terraformDocs.Outputs))
	for i, tfOutput := range terraformDocs.Outputs {
		outputs[i] = Output{
			Name:        tfOutput.Name,
			Description: &tfOutput.Description,
		}
	}

	providerDependencies := make([]ProviderDependency, len(terraformDocs.Providers))
	for i, tfProvider := range terraformDocs.Providers {
		providerDependencies[i] = ProviderDependency{
			Provider: tfProvider.Name,
			Version:  tfProvider.Version,
		}
	}

	resources := make([]Resource, len(terraformDocs.Resources))
	for i, tfResource := range terraformDocs.Resources {
		resources[i] = Resource{
			Type: tfResource.Type,
			Name: tfResource.Name,
		}
	}

	return inputs, outputs, providerDependencies, resources
}

// GetRootModuleSpecs returns the module specifications for the root module
func (mv *ModuleVersion) GetRootModuleSpecs() *ModuleSpecs {
	if mv.details == nil {
		return &ModuleSpecs{
			Path:                 "",
			Readme:               "",
			Empty:                true,
			Inputs:               []Input{},
			Outputs:              []Output{},
			Dependencies:         []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources:            []Resource{},
			Modules:              []Module{},
		}
	}

	// Get terraform docs from module details (stored as JSON during indexing)
	if !mv.details.HasTerraformDocs() {
		return &ModuleSpecs{
			Path:                 "",
			Readme:               string(mv.details.ReadmeContent()),
			Empty:                !mv.details.HasReadme(),
			Inputs:               []Input{},
			Outputs:              []Output{},
			Dependencies:         []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources:            []Resource{},
			Modules:              []Module{},
		}
	}

	// Parse terraform docs JSON from stored data
	terraformDocsJSON := mv.details.TerraformDocs()
	if len(terraformDocsJSON) == 0 {
		return &ModuleSpecs{
			Path:                 "",
			Readme:               string(mv.details.ReadmeContent()),
			Empty:                !mv.details.HasReadme(),
			Inputs:               []Input{},
			Outputs:              []Output{},
			Dependencies:         []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources:            []Resource{},
			Modules:              []Module{},
		}
	}

	// Use the helper method to parse terraform docs JSON
	inputs, outputs, providerDependencies, resources := mv.parseTerraformDocs(terraformDocsJSON)

	// Dependencies and modules would need to be parsed from terraform configuration files
	// For now, return empty slices as terraform-docs doesn't provide this information
	dependencies := []Dependency{}
	modules := []Module{}

	return &ModuleSpecs{
		Path:                 "",
		Readme:               string(mv.details.ReadmeContent()),
		Empty:                !mv.details.HasReadme() && len(terraformDocsJSON) == 0,
		Inputs:               inputs,
		Outputs:              outputs,
		Dependencies:         dependencies,
		ProviderDependencies: providerDependencies,
		Resources:            resources,
		Modules:              modules,
	}
}

// GetSubmodules returns module specifications for all submodules
func (mv *ModuleVersion) GetSubmodules() []*ModuleSpecs {
	var specs []*ModuleSpecs
	for _, submodule := range mv.submodules {
		specs = append(specs, mv.convertSubmoduleToSpecs(submodule))
	}
	return specs
}

// convertSubmoduleToSpecs converts a submodule to ModuleSpecs by deserializing stored data
func (mv *ModuleVersion) convertSubmoduleToSpecs(submodule *Submodule) *ModuleSpecs {
	// Get module details for submodule (stored during indexing)
	details := submodule.Details()
	if details == nil {
		// Return empty specs if no details available
		return &ModuleSpecs{
			Path:                 submodule.Path(),
			Readme:               "",
			Empty:                true,
			Inputs:               []Input{},
			Outputs:              []Output{},
			Dependencies:         []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources:            []Resource{},
			Modules:              []Module{},
		}
	}

	// Deserialize terraform-docs JSON if available
	if !details.HasTerraformDocs() {
		return &ModuleSpecs{
			Path:                 submodule.Path(),
			Readme:               string(details.ReadmeContent()),
			Empty:                !details.HasReadme(),
			Inputs:               []Input{},
			Outputs:              []Output{},
			Dependencies:         []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources:            []Resource{},
			Modules:              []Module{},
		}
	}

	terraformDocsJSON := details.TerraformDocs()
	if len(terraformDocsJSON) == 0 {
		return &ModuleSpecs{
			Path:                 submodule.Path(),
			Readme:               string(details.ReadmeContent()),
			Empty:                !details.HasReadme(),
			Inputs:               []Input{},
			Outputs:              []Output{},
			Dependencies:         []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources:            []Resource{},
			Modules:              []Module{},
		}
	}

	// Reuse the same terraform-docs parsing logic as GetRootModuleSpecs
	inputs, outputs, providerDeps, resources := mv.parseTerraformDocs(terraformDocsJSON)

	return &ModuleSpecs{
		Path:                 submodule.Path(),
		Readme:               string(details.ReadmeContent()),
		Empty:                !details.HasReadme() && len(terraformDocsJSON) == 0,
		Inputs:               inputs,
		Outputs:              outputs,
		Dependencies:         []Dependency{}, // terraform-docs doesn't provide this
		ProviderDependencies: providerDeps,
		Resources:            resources,
		Modules:              []Module{}, // terraform-docs doesn't provide this
	}
}

// GetExamples returns module specifications for all examples
func (mv *ModuleVersion) GetExamples() []*ModuleSpecs {
	var specs []*ModuleSpecs
	for _, example := range mv.examples {
		specs = append(specs, mv.convertExampleToSpecs(example))
	}
	return specs
}

// GetSubmoduleByPath returns a specific submodule by its path
func (mv *ModuleVersion) GetSubmoduleByPath(path string) *Submodule {
	for _, submodule := range mv.submodules {
		if submodule.Path() == path {
			return submodule
		}
	}
	return nil
}

// GetExampleByPath returns a specific example by its path
func (mv *ModuleVersion) GetExampleByPath(path string) *Example {
	for _, example := range mv.examples {
		if example.Path() == path {
			return example
		}
	}
	return nil
}

// convertExampleToSpecs converts an example to ModuleSpecs by deserializing stored data
func (mv *ModuleVersion) convertExampleToSpecs(example *Example) *ModuleSpecs {
	// Get module details for example (stored during indexing)
	details := example.Details()
	if details == nil {
		// Return empty specs if no details available
		return &ModuleSpecs{
			Path:                 example.Path(),
			Readme:               "",
			Empty:                true,
			Inputs:               []Input{},
			Outputs:              []Output{},
			Dependencies:         []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources:            []Resource{},
			Modules:              []Module{},
		}
	}

	// Deserialize terraform-docs JSON if available
	if !details.HasTerraformDocs() {
		return &ModuleSpecs{
			Path:                 example.Path(),
			Readme:               string(details.ReadmeContent()),
			Empty:                !details.HasReadme(),
			Inputs:               []Input{},
			Outputs:              []Output{},
			Dependencies:         []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources:            []Resource{},
			Modules:              []Module{},
		}
	}

	terraformDocsJSON := details.TerraformDocs()
	if len(terraformDocsJSON) == 0 {
		return &ModuleSpecs{
			Path:                 example.Path(),
			Readme:               string(details.ReadmeContent()),
			Empty:                !details.HasReadme(),
			Inputs:               []Input{},
			Outputs:              []Output{},
			Dependencies:         []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources:            []Resource{},
			Modules:              []Module{},
		}
	}

	// Reuse the same terraform-docs parsing logic as GetRootModuleSpecs
	inputs, outputs, providerDeps, resources := mv.parseTerraformDocs(terraformDocsJSON)

	return &ModuleSpecs{
		Path:                 example.Path(),
		Readme:               string(details.ReadmeContent()),
		Empty:                !details.HasReadme() && len(terraformDocsJSON) == 0,
		Inputs:               inputs,
		Outputs:              outputs,
		Dependencies:         []Dependency{}, // terraform-docs doesn't provide this
		ProviderDependencies: providerDeps,
		Resources:            resources,
		Modules:              []Module{}, // terraform-docs doesn't provide this
	}
}

// GetProviderDependencies returns provider dependencies for this module version
func (mv *ModuleVersion) GetProviderDependencies() []ProviderDependency {
	// TODO: Parse from terraform docs
	return []ProviderDependency{}
}

// GetTerraformModules returns terraform module dependencies (modules field)
func (mv *ModuleVersion) GetTerraformModules() []Module {
	// TODO: Parse from terraform modules data
	return []Module{}
}

// GetUsageExample generates a Terraform usage example with the given request domain
func (mv *ModuleVersion) GetUsageExample(requestDomain string) string {
	if mv.moduleProvider == nil {
		return ""
	}

	// Generate usage example format: module "name" { source = "domain/namespace/module/provider" version = "1.0.0" }
	// TODO: Improve formatting and include inputs if available
	namespace := mv.moduleProvider.Namespace()
	moduleName := mv.moduleProvider.Module()
	provider := mv.moduleProvider.Provider()

	if namespace == nil {
		return ""
	}

	source := fmt.Sprintf("%s/%s/%s/%s", requestDomain, namespace.Name(), moduleName, provider)
	return fmt.Sprintf(`module "%s" {
  source  = "%s"
  version = "%s"
}`, moduleName, source, mv.version.String())
}

// GetSecurityResults returns security scan results (tfsec)
func (mv *ModuleVersion) GetSecurityResults() []SecurityResult {
	if mv.details == nil || !mv.details.HasTfsec() {
		return []SecurityResult{}
	}

	// Get tfsec JSON results from stored data (processed during indexing)
	tfsecJSON := mv.details.Tfsec()
	if len(tfsecJSON) == 0 {
		return []SecurityResult{}
	}

	// Define struct for unmarshaling tfsec JSON results
	type TfsecResults struct {
		Results []struct {
			Description string `json:"description"`
			Impact      string `json:"impact"`
			Location    struct {
				EndLine   int    `json:"end_line"`
				Filename  string `json:"filename"`
				StartLine int    `json:"start_line"`
			} `json:"location"`
			RuleID   string `json:"rule_id"`
			Severity string `json:"severity"`
			Status   int    `json:"status"`
			Title    string `json:"title"`
		} `json:"results"`
	}

	var tfsecResults TfsecResults
	if err := json.Unmarshal(tfsecJSON, &tfsecResults); err != nil {
		// If unmarshaling fails, return empty results
		return []SecurityResult{}
	}

	// Convert tfsec results to domain SecurityResult structs
	securityResults := make([]SecurityResult, len(tfsecResults.Results))
	for i, result := range tfsecResults.Results {
		securityResults[i] = SecurityResult{
			RuleID:      result.RuleID,
			Severity:    result.Severity,
			Title:       result.Title,
			Description: result.Description,
			Location: SecurityLocation{
				Filename:  result.Location.Filename,
				StartLine: result.Location.StartLine,
				EndLine:   result.Location.EndLine,
			},
		}
	}

	return securityResults
}

// GetSecurityFailures returns the count of security scan failures
func (mv *ModuleVersion) GetSecurityFailures() int {
	results := mv.GetSecurityResults()
	return len(results)
}

// GetCustomLinks returns formatted custom links for this module version
func (mv *ModuleVersion) GetCustomLinks() []CustomLink {
	// @TODO
	// For now, return empty slice
	// In a full implementation, this would:
	// 1. Load MODULE_LINKS configuration from environment/database
	// 2. Filter links based on namespace restrictions
	// 3. Format template variables (namespace, module, provider, version)
	// 4. Return formatted CustomLink structs

	// Example of what a configured link might look like:
	// {
	//   Text: "View in Git Repository",
	//   URL: "https://github.com/{namespace}/{module}/tree/{tag}",
	// }

	return []CustomLink{}
}

// GetAdditionalTabFiles returns additional tab file configuration
func (mv *ModuleVersion) GetAdditionalTabFiles() map[string]string {
	// @TODO
	// For now, return empty map
	// In a full implementation, this would:
	// 1. Load ADDITIONAL_MODULE_TABS configuration from environment/database
	// 2. Get list of all files in this module version
	// 3. Check which configured tab files exist in the module
	// 4. Return mapping of tab name -> file path

	// Example of what tab files might look like:
	// {
	//   "Release Notes": "RELEASE_NOTES.md",
	//   "License": "LICENSE",
	//   "Contributing": "CONTRIBUTING.md",
	// }

	return make(map[string]string)
}

// GetPublishedAtDisplay returns formatted publication date for UI display
func (mv *ModuleVersion) GetPublishedAtDisplay() string {
	if !mv.published || mv.publishedAt == nil {
		return ""
	}

	// Format: "January 2, 2006 at 3:04pm" (more readable for UI)
	return mv.publishedAt.Format("January 2, 2006 at 3:04pm")
}

// GetDisplaySourceURL returns the browse URL or fallback
func (mv *ModuleVersion) GetDisplaySourceURL(requestDomain string) string {
	if mv.moduleProvider == nil {
		return ""
	}

	// Get repository browse URL template following hierarchy:
	// 1. ModuleVersion template
	// 2. ModuleProvider template
	// 3. GitProvider template
	var template *string
	if mv.repoBrowseURLTemplate != nil {
		template = mv.repoBrowseURLTemplate
	} else if mv.moduleProvider != nil {
		template = mv.moduleProvider.RepoBrowseURLTemplate()
		// If still nil, check GitProvider
		if template == nil && mv.moduleProvider.gitProvider != nil {
			gitProviderTemplate := mv.moduleProvider.gitProvider.BrowseURLTemplate
			template = &gitProviderTemplate
		}
	}

	if template == nil || *template == "" {
		// No template configured, return empty string
		// Could implement fallback to requestDomain here if needed
		return ""
	}

	// Get template variables
	namespace := mv.moduleProvider.Namespace().Name()
	moduleName := mv.moduleProvider.Module()
	provider := mv.moduleProvider.Provider()
	tag := mv.getSourceGitTag()

	return mv.expandURLTemplate(*template, namespace, moduleName, provider, tag, "")
}

// getSourceGitTag generates the git tag from version and git tag format
func (mv *ModuleVersion) getSourceGitTag() string {
	versionStr := mv.version.String()

	// Get git tag format from module provider
	gitTagFormat := mv.moduleProvider.GitTagFormat()
	if gitTagFormat == nil || *gitTagFormat == "" {
		// Default format is just the version
		return versionStr
	}

	// Parse version for components (simplified - could use semver library if needed)
	// For now, just replace {version} placeholder
	template := *gitTagFormat
	result := strings.ReplaceAll(template, "{version}", versionStr)

	return result
}

// expandURLTemplate replaces template variables with actual values
func (mv *ModuleVersion) expandURLTemplate(template, namespace, moduleName, provider, tag, path string) string {
	// Prepare template variables
	tagURIEncoded := url.QueryEscape(tag)

	// Replace template variables
	result := template
	result = strings.ReplaceAll(result, "{namespace}", namespace)
	result = strings.ReplaceAll(result, "{module}", moduleName)
	result = strings.ReplaceAll(result, "{provider}", provider)
	result = strings.ReplaceAll(result, "{tag}", tag)
	result = strings.ReplaceAll(result, "{tag_uri_encoded}", tagURIEncoded)
	result = strings.ReplaceAll(result, "{path}", path)

	return result
}

// GetGraphURL returns the graph URL for this module
func (mv *ModuleVersion) GetGraphURL() string {
	if mv.moduleProvider == nil {
		return ""
	}

	namespace := mv.moduleProvider.Namespace()
	moduleName := mv.moduleProvider.Module()

	if namespace == nil {
		return ""
	}

	// Format: /modules/{namespace}/{moduleName}/graph
	return fmt.Sprintf("/modules/%s/%s/graph", namespace.Name(), moduleName)
}

// GetModuleExtractionUpToDate checks if module extraction is current
func (mv *ModuleVersion) GetModuleExtractionUpToDate() bool {
	// TODO: Implement extraction version checking
	return mv.extractionVersion != nil
}

// GetTerraformExampleVersionString returns version constraint for examples
func (mv *ModuleVersion) GetTerraformExampleVersionString() string {
	// For beta versions or non-latest versions, return exact version
	if mv.beta || !mv.isLatestVersion() {
		return mv.version.String()
	}

	// For latest published versions, generate version constraint from template
	// Parse version components
	version := mv.version
	major := version.Major()
	minor := version.Minor()
	patch := version.Patch()

	// Create template variables
	templateVars := map[string]interface{}{
		"major":           major,
		"minor":           minor,
		"patch":           patch,
		"major_plus_one":  major + 1,
		"minor_plus_one":  minor + 1,
		"patch_plus_one":  patch + 1,
		"major_minus_one": max(0, major-1),
		"minor_minus_one": max(0, minor-1),
		"patch_minus_one": max(0, patch-1),
	}

	// Determine which template to use
	// Default template matches Python defaults
	defaultTemplate := "{major}.{minor}.{patch}"
	preMajorTemplate := "{major}.{minor}.{patch}" // Falls back to default in Python

	template := defaultTemplate

	// For pre-1.0.0 versions, use pre-major template (currently same as default)
	if major == 0 {
		template = preMajorTemplate
	}

	// Expand all template variables
	result := template
	result = strings.ReplaceAll(result, "{major}", fmt.Sprintf("%d", templateVars["major"]))
	result = strings.ReplaceAll(result, "{minor}", fmt.Sprintf("%d", templateVars["minor"]))
	result = strings.ReplaceAll(result, "{patch}", fmt.Sprintf("%d", templateVars["patch"]))
	result = strings.ReplaceAll(result, "{major_plus_one}", fmt.Sprintf("%d", templateVars["major_plus_one"]))
	result = strings.ReplaceAll(result, "{minor_plus_one}", fmt.Sprintf("%d", templateVars["minor_plus_one"]))
	result = strings.ReplaceAll(result, "{patch_plus_one}", fmt.Sprintf("%d", templateVars["patch_plus_one"]))
	result = strings.ReplaceAll(result, "{major_minus_one}", fmt.Sprintf("%d", templateVars["major_minus_one"]))
	result = strings.ReplaceAll(result, "{minor_minus_one}", fmt.Sprintf("%d", templateVars["minor_minus_one"]))
	result = strings.ReplaceAll(result, "{patch_minus_one}", fmt.Sprintf("%d", templateVars["patch_minus_one"]))

	return result
}

// GetTerraformExampleVersionComment returns version comments for examples
func (mv *ModuleVersion) GetTerraformExampleVersionComment() []string {
	// Check if version is published
	if !mv.published {
		return []string{
			"This version of this module has not yet been published,",
			"meaning that it cannot yet be used by Terraform",
		}
	}

	// Check if version is beta
	if mv.beta {
		return []string{
			"This version of the module is a beta version.",
			"To use this version, it must be pinned in Terraform",
		}
	}

	// Check if this is the latest version
	if !mv.isLatestVersion() {
		return []string{
			"This version of the module is not the latest version.",
			"To use this specific version, it must be pinned in Terraform",
		}
	}

	// For latest published versions, no comments needed
	return []string{}
}

// isLatestVersion checks if this version is the latest published version
func (mv *ModuleVersion) isLatestVersion() bool {
	if mv.moduleProvider == nil {
		return false
	}

	latestVersion := mv.moduleProvider.GetLatestVersion()
	if latestVersion == nil {
		return false
	}

	return mv.version.String() == latestVersion.version.String()
}

// String returns the string representation
func (mv *ModuleVersion) String() string {
	return mv.version.String()
}

// VersionedID returns the versioned module ID for terrareg API calls
// Format: namespace/name/provider/version
func (mv *ModuleVersion) VersionedID() string {
	if mv.moduleProvider == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s/%s",
		mv.moduleProvider.Namespace().Name(),
		mv.moduleProvider.Module(),
		mv.moduleProvider.Provider(),
		mv.version.String())
}

// ModuleProviderID returns the provider ID without version
// Format: namespace/name/provider
func (mv *ModuleVersion) ModuleProviderID() string {
	if mv.moduleProvider == nil {
		return ""
	}
	return mv.moduleProvider.FrontendID()
}

// Module specifications types for terrareg API

// ModuleSpecs represents terraform module specifications
type ModuleSpecs struct {
	Path                 string               `json:"path"`
	Readme               string               `json:"readme"`
	Empty                bool                 `json:"empty"`
	Inputs               []Input              `json:"inputs"`
	Outputs              []Output             `json:"outputs"`
	Dependencies         []Dependency         `json:"dependencies"`
	ProviderDependencies []ProviderDependency `json:"provider_dependencies"`
	Resources            []Resource           `json:"resources"`
	Modules              []Module             `json:"modules"`
}

// Input represents a terraform input variable
type Input struct {
	Name           string      `json:"name"`
	Type           string      `json:"type"`
	Description    *string     `json:"description"`
	Required       bool        `json:"required"`
	Default        interface{} `json:"default"`
	AdditionalHelp *string     `json:"additional_help,omitempty"`
	QuoteValue     bool        `json:"quote_value,omitempty"`
	Sensitive      bool        `json:"sensitive,omitempty"`
}

// Output represents a terraform output
type Output struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Type        *string `json:"type,omitempty"`
}

// Dependency represents a terraform module dependency
type Dependency struct {
	Module  string `json:"module"`
	Source  string `json:"source"`
	Version string `json:"version"`
}

// ProviderDependency represents a terraform provider dependency
type ProviderDependency struct {
	Provider string `json:"provider"`
	Source   string `json:"source,omitempty"`
	Version  string `json:"version,omitempty"`
}

// Resource represents a terraform resource
type Resource struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Module represents a terraform module dependency (modules field)
type Module struct {
	Name      string   `json:"name"`
	Source    string   `json:"source"`
	Version   string   `json:"version"`
	Key       string   `json:"key"`
	Providers []string `json:"providers"`
}

// SecurityResult represents a security scan result
type SecurityResult struct {
	RuleID      string           `json:"rule_id"`
	Severity    string           `json:"severity"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Location    SecurityLocation `json:"location"`
}

// SecurityLocation represents the location of a security issue
type SecurityLocation struct {
	Filename  string `json:"filename"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// CustomLink represents a custom link with template formatting
type CustomLink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

// Submodule represents a submodule within a module version
type Submodule struct {
	id      int
	path    string
	name    *string
	subType *string
	details *ModuleDetails
}

// NewSubmodule creates a new submodule
func NewSubmodule(path string, name, subType *string, details *ModuleDetails) *Submodule {
	return &Submodule{
		path:    path,
		name:    name,
		subType: subType,
		details: details,
	}
}

func (s *Submodule) ID() int                 { return s.id }
func (s *Submodule) Path() string            { return s.path }
func (s *Submodule) Name() *string           { return s.name }
func (s *Submodule) Type() *string           { return s.subType }
func (s *Submodule) Details() *ModuleDetails { return s.details }

// Example represents an example within a module version
type Example struct {
	id      int
	path    string
	name    *string
	details *ModuleDetails
	files   []*ExampleFile
}

// NewExample creates a new example
func NewExample(path string, name *string, details *ModuleDetails) *Example {
	return &Example{
		path:    path,
		name:    name,
		details: details,
		files:   make([]*ExampleFile, 0),
	}
}

func (e *Example) ID() int                 { return e.id }
func (e *Example) Path() string            { return e.path }
func (e *Example) Name() *string           { return e.name }
func (e *Example) Details() *ModuleDetails { return e.details }
func (e *Example) Files() []*ExampleFile   { return e.files }

func (e *Example) AddFile(file *ExampleFile) {
	e.files = append(e.files, file)
}

// ExampleFile represents a file within an example
type ExampleFile struct {
	id      int
	path    string
	content []byte
}

// NewExampleFile creates a new example file
func NewExampleFile(path string, content []byte) *ExampleFile {
	return &ExampleFile{
		path:    path,
		content: content,
	}
}

func (ef *ExampleFile) ID() int         { return ef.id }
func (ef *ExampleFile) Path() string    { return ef.path }
func (ef *ExampleFile) Content() []byte { return ef.content }

// ModuleFile represents an additional file in the module version
type ModuleFile struct {
	id      int
	path    string
	content []byte
}

// NewModuleFile creates a new module file
func NewModuleFile(path string, content []byte) *ModuleFile {
	return &ModuleFile{
		path:    path,
		content: content,
	}
}

func (mf *ModuleFile) ID() int         { return mf.id }
func (mf *ModuleFile) Path() string    { return mf.path }
func (mf *ModuleFile) Content() []byte { return mf.content }
