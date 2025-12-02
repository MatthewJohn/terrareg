package model

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// ModuleProvider is the aggregate root for a module provider
type ModuleProvider struct {
	id        int
	namespace *Namespace
	module    string
	provider  string
	verified  bool

	// Git configuration
	gitProvider           *model.GitProvider
	gitProviderID         *int
	repoBaseURLTemplate   *string
	repoCloneURLTemplate  *string
	repoBrowseURLTemplate *string
	gitTagFormat          *string
	gitPath               *string
	archiveGitPath        bool

	// Versions (entities within aggregate)
	versions      []*ModuleVersion
	latestVersion *ModuleVersion

	// Timestamps
	createdAt time.Time
	updatedAt time.Time
}

var (
	moduleNameRegex   = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$`)
	providerNameRegex = regexp.MustCompile(`^[a-z0-9]+$`)
)

// NewModuleProvider creates a new module provider
func NewModuleProvider(namespace *Namespace, moduleName, providerName string) (*ModuleProvider, error) {
	if namespace == nil {
		return nil, fmt.Errorf("namespace cannot be nil")
	}

	if err := ValidateModuleName(moduleName); err != nil {
		return nil, err
	}

	if err := ValidateProviderName(providerName); err != nil {
		return nil, err
	}

	now := time.Now()
	return &ModuleProvider{
		namespace:  namespace,
		module:     moduleName,
		provider:   providerName,
		verified:   false,
		versions:   make([]*ModuleVersion, 0),
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

// ReconstructModuleProvider reconstructs a module provider from persistence
func ReconstructModuleProvider(
	id int,
	namespace *Namespace,
	moduleName, providerName string,
	verified bool,
	gitProviderID *int,
	repoBaseURLTemplate, repoCloneURLTemplate, repoBrowseURLTemplate *string,
	gitTagFormat, gitPath *string,
	archiveGitPath bool,
	createdAt, updatedAt time.Time,
) *ModuleProvider {
	return &ModuleProvider{
		id:                    id,
		namespace:             namespace,
		module:                moduleName,
		provider:              providerName,
		verified:              verified,
		gitProviderID:         gitProviderID,
		repoBaseURLTemplate:   repoBaseURLTemplate,
		repoCloneURLTemplate:  repoCloneURLTemplate,
		repoBrowseURLTemplate: repoBrowseURLTemplate,
		gitTagFormat:          gitTagFormat,
		gitPath:               gitPath,
		archiveGitPath:        archiveGitPath,
		versions:              make([]*ModuleVersion, 0),
		createdAt:             createdAt,
		updatedAt:             updatedAt,
	}
}

// ValidateModuleName validates a module name
func ValidateModuleName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: module name cannot be empty", shared.ErrInvalidName)
	}

	if len(name) < 2 {
		return fmt.Errorf("%w: module name must be at least 2 characters", shared.ErrInvalidName)
	}

	if len(name) > 128 {
		return fmt.Errorf("%w: module name must not exceed 128 characters", shared.ErrInvalidName)
	}

	// Convert to lowercase for validation
	name = strings.ToLower(name)

	if !moduleNameRegex.MatchString(name) {
		return fmt.Errorf("%w: module name must contain only alphanumeric characters, hyphens, and underscores", shared.ErrInvalidName)
	}

	return nil
}

// ValidateProviderName validates a provider name
func ValidateProviderName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: provider name cannot be empty", shared.ErrInvalidProvider)
	}

	if len(name) < 2 {
		return fmt.Errorf("%w: provider name must be at least 2 characters", shared.ErrInvalidProvider)
	}

	if len(name) > 128 {
		return fmt.Errorf("%w: provider name must not exceed 128 characters", shared.ErrInvalidProvider)
	}

	// Provider names must be lowercase alphanumeric only
	if !providerNameRegex.MatchString(name) {
		return fmt.Errorf("%w: provider name must contain only lowercase alphanumeric characters", shared.ErrInvalidProvider)
	}

	return nil
}

// Business methods

// AddVersion adds a new version to this module provider
func (mp *ModuleProvider) AddVersion(version *ModuleVersion) error {
	if version == nil {
		return fmt.Errorf("version cannot be nil")
	}

	// Check if version already exists
	for _, v := range mp.versions {
		if v.Version().Equal(version.Version()) {
			return fmt.Errorf("%w: version %s already exists", shared.ErrAlreadyExists, version.Version())
		}
	}

	// Set the parent reference
	version.setModuleProvider(mp)

	// Add to versions
	mp.versions = append(mp.versions, version)

	// Update latest version if this is not a beta/prerelease and is newer
	mp.updateLatestVersion()

	mp.updatedAt = time.Now()
	return nil
}

// PublishVersion publishes a version
func (mp *ModuleProvider) PublishVersion(versionStr string, details *ModuleDetails, beta bool) (*ModuleVersion, error) {
	version, err := NewModuleVersion(versionStr, details, beta)
	if err != nil {
		return nil, err
	}

	if err := mp.AddVersion(version); err != nil {
		return nil, err
	}

	return version, nil
}

// Verify marks this module provider as verified
func (mp *ModuleProvider) Verify() error {
	if mp.verified {
		return fmt.Errorf("%w: module provider is already verified", shared.ErrDomainViolation)
	}
	mp.verified = true
	mp.updatedAt = time.Now()
	return nil
}

// Unverify removes verification from this module provider
func (mp *ModuleProvider) Unverify() {
	mp.verified = false
	mp.updatedAt = time.Now()
}

// SetGitConfiguration sets the Git configuration
func (mp *ModuleProvider) SetGitConfiguration(
	gitProviderID *int,
	repoBaseURL, repoCloneURL, repoBrowseURL *string,
	gitTagFormat, gitPath *string,
	archiveGitPath bool,
) {
	mp.gitProviderID = gitProviderID
	mp.repoBaseURLTemplate = repoBaseURL
	mp.repoCloneURLTemplate = repoCloneURL
	mp.repoBrowseURLTemplate = repoBrowseURL
	mp.gitTagFormat = gitTagFormat
	mp.gitPath = gitPath
	mp.archiveGitPath = archiveGitPath
	mp.updatedAt = time.Now()
}

// updateLatestVersion updates the latest version to the highest published, non-beta version
func (mp *ModuleProvider) updateLatestVersion() {
	var latest *ModuleVersion
	for _, v := range mp.versions {
		if !v.IsPublished() || v.IsBeta() {
			continue
		}

		if latest == nil || v.Version().GreaterThan(latest.Version()) {
			latest = v
		}
	}
	mp.latestVersion = latest
}

// GetVersion retrieves a specific version
func (mp *ModuleProvider) GetVersion(versionStr string) (*ModuleVersion, error) {
	targetVersion, err := shared.ParseVersion(versionStr)
	if err != nil {
		return nil, err
	}

	for _, v := range mp.versions {
		if v.Version().Equal(targetVersion) {
			return v, nil
		}
	}

	return nil, fmt.Errorf("%w: version %s", shared.ErrNotFound, versionStr)
}

// GetLatestVersion returns the latest published version (excluding betas)
func (mp *ModuleProvider) GetLatestVersion() *ModuleVersion {
	return mp.latestVersion
}

// GetAllVersions returns all versions
func (mp *ModuleProvider) GetAllVersions() []*ModuleVersion {
	return mp.versions
}

// GetPublishedVersions returns only published versions
func (mp *ModuleProvider) GetPublishedVersions() []*ModuleVersion {
	published := make([]*ModuleVersion, 0)
	for _, v := range mp.versions {
		if v.IsPublished() {
			published = append(published, v)
		}
	}
	return published
}

// SetVersions sets the versions (used by repository during reconstruction)
func (mp *ModuleProvider) SetVersions(versions []*ModuleVersion) {
	mp.versions = versions
	for _, v := range versions {
		v.setModuleProvider(mp)
	}
	mp.updateLatestVersion()
}

// Getters

func (mp *ModuleProvider) ID() int {
	return mp.id
}

func (mp *ModuleProvider) Namespace() *Namespace {
	return mp.namespace
}

func (mp *ModuleProvider) Module() string {
	return mp.module
}

func (mp *ModuleProvider) Provider() string {
	return mp.provider
}

func (mp *ModuleProvider) IsVerified() bool {
	return mp.verified
}

func (mp *ModuleProvider) GitProviderID() *int {
	return mp.gitProviderID
}

func (mp *ModuleProvider) SetGitProvider(gitProvider *model.GitProvider) {
	mp.gitProvider = gitProvider
}

func (mp *ModuleProvider) GitProvider() *model.GitProvider {
	return mp.gitProvider
}

func (mp *ModuleProvider) RepoBaseURLTemplate() *string {
	return mp.repoBaseURLTemplate
}

func (mp *ModuleProvider) RepoCloneURLTemplate() *string {
	return mp.repoCloneURLTemplate
}

func (mp *ModuleProvider) RepoBrowseURLTemplate() *string {
	return mp.repoBrowseURLTemplate
}

func (mp *ModuleProvider) GitTagFormat() *string {
	return mp.gitTagFormat
}

func (mp *ModuleProvider) GitPath() *string {
	return mp.gitPath
}

func (mp *ModuleProvider) ArchiveGitPath() bool {
	return mp.archiveGitPath
}

func (mp *ModuleProvider) CreatedAt() time.Time {
	return mp.createdAt
}

func (mp *ModuleProvider) UpdatedAt() time.Time {
	return mp.updatedAt
}

// FullName returns the full module provider name (namespace/module/provider)
func (mp *ModuleProvider) FullName() string {
	return fmt.Sprintf("%s/%s/%s", mp.namespace.Name(), mp.module, mp.provider)
}

// String returns the string representation
func (mp *ModuleProvider) String() string {
	return mp.FullName()
}
