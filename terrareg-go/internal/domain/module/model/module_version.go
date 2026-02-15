package model

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
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

	if details == nil {
		details = NewModuleDetails([]byte{})
	}

	now := time.Now()
	return &ModuleVersion{
		id:         0, // CRITICAL: Ensure ID starts at 0 for new records
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

	// Ensure details is never nil to prevent panic
	if details == nil {
		details = NewModuleDetails([]byte{})
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
	if mv.details == nil {
		mv.details = NewModuleDetails([]byte{})
	}
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

// ResetID explicitly resets the module version ID to 0
// This is critical for ensuring new records are created instead of updating existing ones
func (mv *ModuleVersion) ResetID() {
	mv.id = 0
	mv.updatedAt = time.Now()
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

// GetGitCloneURL returns the git clone URL for this module version
// Returns empty string if no git URL is configured
// Python reference: models.py get_git_clone_url() (lines 3877-3900)
func (mv *ModuleVersion) GetGitCloneURL(
	ctx context.Context,
	domainConfig *model.DomainConfig,
	gitURLBuilder *service.GitURLBuilderService,
) string {
	// Require module provider to get git URL
	if mv.moduleProvider == nil {
		return ""
	}

	namespace := mv.moduleProvider.Namespace()
	if namespace == nil {
		return ""
	}

	// Determine which template to use based on priority order:
	// 1. Custom version URL (if AllowCustomGitURLModuleVersion enabled)
	// 2. Custom provider URL (if AllowCustomGitURLModuleProvider enabled)
	// 3. Provider source URL from gitProvider

	var template *string

	// Priority 1: Check custom version URL (if allowed)
	if domainConfig.AllowCustomGitURLModuleVersion && mv.repoCloneURLTemplate != nil {
		template = mv.repoCloneURLTemplate
	}

	// Priority 2: Check custom provider URL (if allowed)
	if template == nil && domainConfig.AllowCustomGitURLModuleProvider {
		template = mv.moduleProvider.RepoCloneURLTemplate()
	}

	// Priority 3: Check provider source URL from gitProvider
	if template == nil && mv.moduleProvider.GitProvider() != nil {
		gitProvider := mv.moduleProvider.GitProvider()
		if gitProvider.CloneURLTemplate != "" {
			template = &gitProvider.CloneURLTemplate
		}
	}

	// If no template found, return empty string
	if template == nil || *template == "" {
		return ""
	}

	// Build the clone URL using GitURLBuilderService
	// Get the source git tag for the version
	gitTag := mv.getSourceGitTag()

	req := &service.URLBuilderRequest{
		Template:  *template,
		Namespace: string(namespace.Name()),
		Module:    string(mv.moduleProvider.Module()),
		Provider:  string(mv.moduleProvider.Provider()),
		GitTag:    &gitTag,
	}

	renderedURL, err := gitURLBuilder.BuildCloneURL(req)
	if err != nil {
		// If template building fails, return empty string
		return ""
	}

	return renderedURL
}

// SourceDownloadResult contains the result of GetSourceDownloadURL
// Python reference: models.py get_source_download_url() (lines 3902-3985)
type SourceDownloadResult struct {
	// URL is the download URL (git URL or built-in hosting URL)
	URL string

	// IsGitURL indicates if the URL is a git URL (with git:: prefix)
	IsGitURL bool

	// RequiresAuth indicates if authentication is required for this download
	RequiresAuth bool

	// Error contains any error that occurred during URL generation
	Error error
}

// GetSourceDownloadURL returns the appropriate download URL based on ALLOW_MODULE_HOSTING mode
// Python reference: models.py get_source_download_url() (lines 3902-3985)
func (mv *ModuleVersion) GetSourceDownloadURL(
	ctx context.Context,
	domainConfig *model.DomainConfig,
	gitURLBuilder *service.GitURLBuilderService,
	requestDomain string,
	directHTTPRequest bool,
	path string,
) *SourceDownloadResult {
	result := &SourceDownloadResult{
		URL:          "",
		IsGitURL:     false,
		RequiresAuth: false,
		Error:        nil,
	}

	// Require module provider to get download URL
	if mv.moduleProvider == nil {
		result.Error = fmt.Errorf("module provider not set")
		return result
	}

	namespace := mv.moduleProvider.Namespace()
	if namespace == nil {
		result.Error = fmt.Errorf("namespace not set")
		return result
	}

	// Priority 1: If module hosting is not ENFORCE, attempt to get git clone URL
	if domainConfig.AllowModuleHosting != model.ModuleHostingModeEnforce {
		gitCloneURL := mv.GetGitCloneURL(ctx, domainConfig, gitURLBuilder)
		if gitCloneURL != "" {
			// Build git URL with proper prefix, path, and ref
			result.URL = mv.buildGitURL(gitCloneURL, path, domainConfig)
			result.IsGitURL = true
			return result
		}
	}

	// Priority 2: If a git URL is not present, revert to using built-in module hosting
	// (if not DISALLOW mode)
	if domainConfig.AllowModuleHosting != model.ModuleHostingModeDisallow {
		result.URL = mv.buildBuiltInURL(requestDomain, directHTTPRequest, path, domainConfig)
		result.IsGitURL = false
		result.RequiresAuth = !domainConfig.AllowUnidentifiedDownloads
		return result
	}

	// Priority 3: No valid download method available
	result.Error = fmt.Errorf("module is not configured with a git URL and direct downloads are disabled")
	return result
}

// buildGitURL constructs the git URL with proper prefix, path, and ref
// Python reference: models.py lines 3914-3945
func (mv *ModuleVersion) buildGitURL(gitCloneURL string, path string, domainConfig *model.DomainConfig) string {
	renderedURL := gitCloneURL

	// Check if scheme starts with git::, which is required by Terraform
	// to acknowledge a git repository, and add if not present
	if !strings.HasPrefix(renderedURL, "git::") {
		renderedURL = "git::" + renderedURL
	}

	// Check if git_path has been set and prepend to path, if set
	gitPath := ""
	if mv.gitPath != nil {
		gitPath = *mv.gitPath
	}
	fullPath := strings.Join([]string{gitPath, path}, "/")

	// Check if path is present for module
	if fullPath != "" {
		// Remove any trailing slashes from path
		fullPath = strings.Trim(fullPath, "/")

		renderedURL = fmt.Sprintf("%s//%s", renderedURL, fullPath)
	}

	// Add git ref - if enabled and available, get git commit SHA.
	// Otherwise, fallback to git tag ref
	ref := mv.getSourceGitTag()
	if domainConfig.ModuleVersionUseGitCommit && mv.gitSHA != nil && *mv.gitSHA != "" {
		ref = *mv.gitSHA
	}

	renderedURL = fmt.Sprintf("%s?ref=%s", renderedURL, ref)

	return renderedURL
}

// buildBuiltInURL constructs the built-in hosting URL
// Python reference: models.py lines 3948-3980
func (mv *ModuleVersion) buildBuiltInURL(requestDomain string, directHTTPRequest bool, path string, domainConfig *model.DomainConfig) string {
	// Build base URL: /v1/terrareg/modules/{id}
	url := fmt.Sprintf("/v1/terrareg/modules/%d", mv.id)

	// If authentication is required, generate pre-signed URL
	// (This is a placeholder - actual presigned URL generation would be done elsewhere)
	if !domainConfig.AllowUnidentifiedDownloads {
		// Presigned URL would be generated here
		// For now, just use a placeholder
		url += "/{presign_key}"
	}

	// Add archive filename
	url += "/" + mv.getArchiveName()

	// If archive does not contain just the git_path,
	// check if git_path has been set and prepend to path, if set
	if !mv.archiveGitPath {
		gitPath := ""
		if mv.gitPath != nil {
			gitPath = *mv.gitPath
		}
		fullPath := strings.Join([]string{gitPath, path}, "/")

		// Check if path is present for module
		if fullPath != "" {
			// Remove any trailing slashes from path
			fullPath = strings.Trim(fullPath, "/")

			url = fmt.Sprintf("%s//%s", url, fullPath)
		}
	}

	// If request is a direct HTTP request, provide a full HTTP URL
	if directHTTPRequest && requestDomain != "" {
		url = requestDomain + url
	}

	return url
}

// getArchiveName returns the archive filename for this module version
func (mv *ModuleVersion) getArchiveName() string {
	return fmt.Sprintf("%s.zip", mv.version.String())
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

// parseProviderName splits provider name into namespace and name
// e.g., "hashicorp/aws" -> ("aws", "hashicorp")
// e.g., "aws" -> ("aws", "hashicorp")
// Python reference: /app/terrareg/models.py BaseSubmodule.get_terraform_provider_dependencies()
func parseProviderName(providerName string) (name, namespace string) {
	parts := strings.Split(providerName, "/")
	if len(parts) > 1 {
		// Has namespace - e.g., "hashicorp/aws"
		namespace = parts[0]
		name = strings.Join(parts[1:], "/")
	} else {
		// No namespace - default to hashicorp
		name = providerName
		namespace = "hashicorp"
	}
	return name, namespace
}

// parseTerraformDocs is a helper method to parse terraform-docs JSON into domain models
func (mv *ModuleVersion) parseTerraformDocs(terraformDocsJSON []byte) ([]Input, []Output, []ProviderDependency, []Resource, []Module, []Requirement) {
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
		Modules []struct {
			Name    string `json:"name"`
			Source  string `json:"source"`
			Version string `json:"version,omitempty"`
		} `json:"modules"`
		Requirements []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"requirements"`
	}

	var terraformDocs TerraformDocsJSON
	if err := json.Unmarshal(terraformDocsJSON, &terraformDocs); err != nil {
		// If unmarshaling fails, return empty slices
		return []Input{}, []Output{}, []ProviderDependency{}, []Resource{}, []Module{}, []Requirement{}
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
		name, namespace := parseProviderName(tfProvider.Name)
		providerDependencies[i] = ProviderDependency{
			Name:      name,
			Namespace: namespace,
			Source:    "",
			Version:   tfProvider.Version,
		}
	}

	resources := make([]Resource, len(terraformDocs.Resources))
	for i, tfResource := range terraformDocs.Resources {
		resources[i] = Resource{
			Type: tfResource.Type,
			Name: tfResource.Name,
		}
	}

	modules := make([]Module, len(terraformDocs.Modules))
	for i, tfModule := range terraformDocs.Modules {
		modules[i] = Module{
			Name:    tfModule.Name,
			Source:  tfModule.Source,
			Version: tfModule.Version,
		}
	}

	requirements := make([]Requirement, len(terraformDocs.Requirements))
	for i, tfRequirement := range terraformDocs.Requirements {
		requirements[i] = Requirement{
			Name:    tfRequirement.Name,
			Version: tfRequirement.Version,
		}
	}

	return inputs, outputs, providerDependencies, resources, modules, requirements
}

// convertModulesToDependencies converts modules to dependencies, filtering out local references
// Python reference: /app/terrareg/models.py BaseSubmodule.get_terraform_dependencies()
func convertModulesToDependencies(modules []Module) []Dependency {
	if modules == nil {
		return []Dependency{}
	}
	dependencies := make([]Dependency, 0, len(modules))
	for _, module := range modules {
		// Ignore any modules that reference local directories
		if len(module.Source) >= 2 && (module.Source[0:2] == "./" || module.Source[0:3] == "../") {
			continue
		}
		dependencies = append(dependencies, Dependency{
			Module:  module.Name,
			Source:  module.Source,
			Version: module.Version,
		})
	}
	return dependencies
}

// buildSpecsFromDetails builds ModuleSpecs from ModuleDetails and a path.
// This is a shared helper used by GetRootModuleSpecs, convertSubmoduleToSpecs, and convertExampleToSpecs.
func (mv *ModuleVersion) buildSpecsFromDetails(details *ModuleDetails, path string) *ModuleSpecs {
	// Start with default empty values
	specs := &ModuleSpecs{
		Path:                 path,
		Readme:               "",
		Empty:                true,
		Inputs:               []Input{},
		Outputs:              []Output{},
		Dependencies:         []Dependency{},
		ProviderDependencies: []ProviderDependency{},
		Resources:            []Resource{},
		Modules:              []Module{},
		Requirements:         []Requirement{},
	}

	// Early return if no details
	if details == nil {
		return specs
	}

	// Update with readme content
	specs.Readme = string(details.ReadmeContent())
	specs.Empty = !details.HasReadme()

	// Early return if no terraform docs
	if !details.HasTerraformDocs() {
		return specs
	}

	terraformDocsJSON := details.TerraformDocs()
	if len(terraformDocsJSON) == 0 {
		return specs
	}

	// Parse terraform docs and update specs
	inputs, outputs, providerDeps, resources, modules, requirements := mv.parseTerraformDocs(terraformDocsJSON)
	specs.Inputs = inputs
	specs.Outputs = outputs
	specs.Dependencies = convertModulesToDependencies(modules)
	specs.ProviderDependencies = providerDeps
	specs.Resources = resources
	specs.Modules = modules
	specs.Requirements = requirements
	specs.Empty = !details.HasReadme() && len(terraformDocsJSON) == 0

	return specs
}

// GetRootModuleSpecs returns the module specifications for the root module
func (mv *ModuleVersion) GetRootModuleSpecs() *ModuleSpecs {
	return mv.buildSpecsFromDetails(mv.details, "")
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
	return mv.buildSpecsFromDetails(submodule.Details(), submodule.Path())
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
	return mv.buildSpecsFromDetails(example.Details(), example.Path())
}

// GetProviderDependencies returns provider dependencies for this module version
func (mv *ModuleVersion) GetProviderDependencies() []ProviderDependency {
	// TODO: Parse from terraform docs
	return []ProviderDependency{}
}

// GetTerraformVersionConstraints returns the terraform version constraint if present
// Python reference: /app/terrareg/models.py BaseSubmodule.get_terraform_version_constraints()
func (mv *ModuleVersion) GetTerraformVersionConstraints() *string {
	specs := mv.GetRootModuleSpecs()
	for _, req := range specs.Requirements {
		if req.Name == "terraform" {
			return &req.Version
		}
	}
	return nil
}

// GetTerraformModules returns terraform module dependencies (modules field)
func (mv *ModuleVersion) GetTerraformModules() []Module {
	// TODO: Parse from terraform modules data
	return []Module{}
}

// PrepareModule prepares the module for indexing/extraction
// TODO: Implement module extraction workflow
// Returns true if module should be auto-published
func (mv *ModuleVersion) PrepareModule() (bool, error) {
	// Placeholder until extraction workflow is implemented
	// This would:
	// 1. Extract the uploaded archive
	// 2. Parse terraform-docs
	// 3. Generate README HTML
	// 4. Process examples and submodules
	// 5. Return whether to auto-publish based on configuration
	return false, nil
}

// Delete deletes this module version and cascades to related entities
// TODO: Implement delete with cascade
func (mv *ModuleVersion) Delete() error {
	// Placeholder until delete is implemented
	// This would:
	// 1. Delete all submodules
	// 2. Delete all examples
	// 3. Delete all files
	// 4. Delete module details
	// 5. Update parent's latest version
	return nil
}

// GetModuleDetailsID returns the module details ID
func (mv *ModuleVersion) GetModuleDetailsID() *int {
	// TODO: Return actual module details ID when details persistence is implemented
	return nil
}

// GetRepositoryURLs returns all configured repository URLs
func (mv *ModuleVersion) GetRepositoryURLs() (baseURL, cloneURL, browseURL string) {
	if mv.repoBaseURLTemplate != nil {
		baseURL = *mv.repoBaseURLTemplate
	}
	if mv.repoCloneURLTemplate != nil {
		cloneURL = *mv.repoCloneURLTemplate
	}
	if mv.repoBrowseURLTemplate != nil {
		browseURL = *mv.repoBrowseURLTemplate
	}
	return
}

// GetVariableTemplate returns the variable template
func (mv *ModuleVersion) GetVariableTemplate() []byte {
	return mv.variableTemplate
}

// HasSubmodules checks if module version has any submodules
func (mv *ModuleVersion) HasSubmodules() bool {
	return len(mv.submodules) > 0
}

// HasExamples checks if module version has any examples
func (mv *ModuleVersion) HasExamples() bool {
	return len(mv.examples) > 0
}

// GetExampleVersionConstraint returns the version constraint for examples
// This is used when generating example Terraform code
func (mv *ModuleVersion) GetExampleVersionConstraint() string {
	return mv.GetTerraformExampleVersionString()
}

// GetUsageExample generates a Terraform usage example with the given source URL
// Python reference: /app/terrareg/models.py BaseSubmodule.get_usage_example()
// The sourceURL parameter should be pre-built by the URLService for proper HTTP/HTTPS handling
func (mv *ModuleVersion) GetUsageExample(sourceURL string) string {
	if mv.moduleProvider == nil {
		return ""
	}

	moduleName := mv.moduleProvider.Module()

	namespace := mv.moduleProvider.Namespace()
	if namespace == nil {
		return ""
	}

	// Build terraform block with source and optional version
	// Python: For HTTPS, version is added as a separate attribute
	// For HTTP, version is embedded in the URL (so the sourceURL contains it)
	result := fmt.Sprintf(`module "%s" {
  source = "%s"`, moduleName, sourceURL)

	// Check if version is already in the source URL (HTTP mode)
	// HTTP URL format: http://domain/modules/provider/{version}
	// HTTPS URL format: domain/provider (no version)
	if !strings.Contains(sourceURL, "/"+mv.version.String()) {
		// Version is not in URL (HTTPS mode), add it as a separate attribute
		result += fmt.Sprintf(`
  version = "%s"`, mv.version.String())
	}

	result += `
}`

	return result
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
func (mv *ModuleVersion) expandURLTemplate(
	template string,
	namespace types.NamespaceName,
	moduleName types.ModuleName,
	provider types.ModuleProviderName,
	tag string,
	path string,
) string {
	// Prepare template variables
	tagURIEncoded := url.QueryEscape(tag)

	// Replace template variables
	result := template
	result = strings.ReplaceAll(result, "{namespace}", string(namespace))
	result = strings.ReplaceAll(result, "{module}", string(moduleName))
	result = strings.ReplaceAll(result, "{provider}", string(provider))
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

	// Format: /modules/{namespace}/{moduleName}/{provider}/{version}/graph (front-end page)
	return fmt.Sprintf("/modules/%s/%s/%s/%s/graph", namespace.Name(), moduleName, mv.moduleProvider.Provider(), mv.version.String())
}

// GetGraphDataURL returns the URL for the graph data API endpoint
func (mv *ModuleVersion) GetGraphDataURL() string {
	namespace := mv.moduleProvider.Namespace()
	moduleName := mv.moduleProvider.Module()

	if namespace == nil {
		return ""
	}

	// Format: /v1/terrareg/modules/{namespace}/{moduleName}/{provider}/{version}/graph/data (API endpoint)
	return fmt.Sprintf("/v1/terrareg/modules/%s/%s/%s/%s/graph/data", namespace.Name(), moduleName, mv.moduleProvider.Provider(), mv.version.String())
}

// GetSourceBrowseURL returns the browse URL for the module version with an optional path
// Python reference: /app/terrareg/models.py BaseSubmodule.get_source_browse_url()
func (mv *ModuleVersion) GetSourceBrowseURL(path string) string {
	displayURL := mv.GetDisplaySourceURL("")
	if displayURL == "" {
		return ""
	}
	if path == "" {
		return displayURL
	}
	return displayURL + path
}

// GetSourceBaseURL returns the base source URL without any path
// Python reference: /app/terrareg/models.py ModuleVersion.get_source_base_url()
func (mv *ModuleVersion) GetSourceBaseURL() string {
	// For now, return the display source URL as the base
	// This could be enhanced to handle different scenarios
	return mv.GetDisplaySourceURL("")
}

// ConvertSubmoduleToSpecs converts a submodule to ModuleSpecs by deserializing stored data
// This is an exported version of convertSubmoduleToSpecs for use by queries
func (mv *ModuleVersion) ConvertSubmoduleToSpecs(submodule *Submodule) *ModuleSpecs {
	return mv.convertSubmoduleToSpecs(submodule)
}

// ConvertExampleToSpecs converts an example to ModuleSpecs by deserializing stored data
// This is an exported version of convertExampleToSpecs for use by queries
func (mv *ModuleVersion) ConvertExampleToSpecs(example *Example) *ModuleSpecs {
	return mv.convertExampleToSpecs(example)
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
func (mv *ModuleVersion) VersionedID() types.ModuleProviderVersionFrontendId {
	if mv.moduleProvider == nil {
		return types.ModuleProviderVersionFrontendId("")
	}
	return types.ModuleProviderVersionFrontendId(fmt.Sprintf("%s/%s/%s/%s",
		mv.moduleProvider.Namespace().Name(),
		mv.moduleProvider.Module(),
		mv.moduleProvider.Provider(),
		mv.version.String()))
}

// ModuleProviderID returns the provider ID without version
// Format: namespace/name/provider
func (mv *ModuleVersion) ModuleProviderID() types.ModuleProviderFrontendId {
	if mv.moduleProvider == nil {
		return types.ModuleProviderFrontendId("")
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
	Requirements         []Requirement        `json:"requirements"`
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
// Python reference: /app/terrareg/models.py BaseSubmodule.get_terraform_provider_dependencies()
type ProviderDependency struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Source    string `json:"source,omitempty"`
	Version   string `json:"version,omitempty"`
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

// Requirement represents a terraform module requirement (e.g., terraform version constraint)
// Python reference: /app/terrareg/models.py BaseSubmodule.get_terraform_version_constraints()
type Requirement struct {
	Name    string `json:"name"`
	Version string `json:"version"`
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
