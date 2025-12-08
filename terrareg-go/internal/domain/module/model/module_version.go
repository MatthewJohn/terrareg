package model

import (
	"fmt"
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

// GetRootModuleSpecs returns the module specifications for the root module
func (mv *ModuleVersion) GetRootModuleSpecs() *ModuleSpecs {
	if mv.details == nil {
		return &ModuleSpecs{
			Path:     "",
			Readme:   "",
			Empty:    true,
			Inputs:   []Input{},
			Outputs:  []Output{},
			Dependencies: []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources: []Resource{},
			Modules:   []Module{},
		}
	}

	// Parse terraform docs from module details
	// TODO: Implement proper parsing from mv.details.TerraformDocs()
	return &ModuleSpecs{
		Path:     "",
		Readme:   string(mv.details.ReadmeContent()),
		Empty:    !mv.details.HasReadme(),
		Inputs:   []Input{},   // TODO: Parse from terraform docs
		Outputs:  []Output{},  // TODO: Parse from terraform docs
		Dependencies: []Dependency{}, // TODO: Parse from terraform docs
		ProviderDependencies: []ProviderDependency{}, // TODO: Parse from terraform docs
		Resources: []Resource{}, // TODO: Parse from terraform docs
		Modules:   []Module{},   // TODO: Parse from terraform docs
	}
}

// GetSubmodules returns module specifications for all submodules
func (mv *ModuleVersion) GetSubmodules() []*ModuleSpecs {
	var specs []*ModuleSpecs
	for _, submodule := range mv.submodules {
		// TODO: Implement proper parsing from submodule details
		specs = append(specs, &ModuleSpecs{
			Path:     submodule.Path(),
			Readme:   "",
			Empty:    true,
			Inputs:   []Input{},
			Outputs:  []Output{},
			Dependencies: []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources: []Resource{},
			Modules:   []Module{},
		})
	}
	return specs
}

// GetExamples returns module specifications for all examples
func (mv *ModuleVersion) GetExamples() []*ModuleSpecs {
	var specs []*ModuleSpecs
	for _, example := range mv.examples {
		// TODO: Implement proper parsing from example details
		specs = append(specs, &ModuleSpecs{
			Path:     example.Path(),
			Readme:   "",
			Empty:    true,
			Inputs:   []Input{},
			Outputs:  []Output{},
			Dependencies: []Dependency{},
			ProviderDependencies: []ProviderDependency{},
			Resources: []Resource{},
			Modules:   []Module{},
		})
	}
	return specs
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

	// TODO: Parse tfsec JSON results into SecurityResult structs
	return []SecurityResult{}
}

// GetSecurityFailures returns the count of security scan failures
func (mv *ModuleVersion) GetSecurityFailures() int {
	results := mv.GetSecurityResults()
	return len(results)
}

// GetCustomLinks returns formatted custom links for this module version
func (mv *ModuleVersion) GetCustomLinks() []CustomLink {
	// TODO: Implement custom links with template formatting
	return []CustomLink{}
}

// GetAdditionalTabFiles returns additional tab file configuration
func (mv *ModuleVersion) GetAdditionalTabFiles() map[string]string {
	// TODO: Return additional tab configuration from module provider or global config
	return make(map[string]string)
}

// GetPublishedAtDisplay returns formatted publication date for UI display
func (mv *ModuleVersion) GetPublishedAtDisplay() string {
	if !mv.published || mv.publishedAt == nil {
		return ""
	}

	// TODO: Use proper date formatting
	return mv.publishedAt.Format("2006-01-02")
}

// GetDisplaySourceURL returns the browse URL or fallback
func (mv *ModuleVersion) GetDisplaySourceURL(requestDomain string) string {
	// TODO: Generate browse URL from git configuration or fallback
	return ""
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
	// TODO: Generate appropriate version constraint for examples
	return ""
}

// GetTerraformExampleVersionComment returns version comments for examples
func (mv *ModuleVersion) GetTerraformExampleVersionComment() []string {
	// TODO: Generate version comments for examples
	return []string{}
}

// String returns the string representation
func (mv *ModuleVersion) String() string {
	return mv.version.String()
}

// Module specifications types for terrareg API

// ModuleSpecs represents terraform module specifications
type ModuleSpecs struct {
	Path                 string                `json:"path"`
	Readme               string                `json:"readme"`
	Empty                bool                  `json:"empty"`
	Inputs               []Input               `json:"inputs"`
	Outputs              []Output              `json:"outputs"`
	Dependencies         []Dependency          `json:"dependencies"`
	ProviderDependencies []ProviderDependency  `json:"provider_dependencies"`
	Resources            []Resource            `json:"resources"`
	Modules              []Module              `json:"modules"`
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
	RuleID      string         `json:"rule_id"`
	Severity    string         `json:"severity"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
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
