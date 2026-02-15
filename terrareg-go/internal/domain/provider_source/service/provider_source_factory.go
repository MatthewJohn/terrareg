package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ProviderSourceClass defines the interface for provider source implementations
// Python reference: provider_source/base.py::BaseProviderSource
type ProviderSourceClass interface {
	// Type returns the provider source type (e.g., "github", "gitlab")
	Type() model.ProviderSourceType

	// GenerateDBConfigFromSourceConfig validates and converts user config to DB config
	// Python reference: github.py::generate_db_config_from_source_config()
	GenerateDBConfigFromSourceConfig(sourceConfig map[string]interface{}) (*model.ProviderSourceConfig, error)

	// CreateInstance creates a provider source instance with the given name
	CreateInstance(name string, repo repository.ProviderSourceRepository, db interface{}) (ProviderSourceInstance, error)
}

// ProviderSourceFactory manages provider sources
// Python reference: provider_source/factory.py::ProviderSourceFactory
type ProviderSourceFactory struct {
	repo         repository.ProviderSourceRepository
	classMapping map[model.ProviderSourceType]ProviderSourceClass
	mu           sync.RWMutex
	db           interface{} // Database for provider source instances
}

// NewProviderSourceFactory creates a new ProviderSourceFactory
// This should be called during container initialization
func NewProviderSourceFactory(repo repository.ProviderSourceRepository) *ProviderSourceFactory {
	return &ProviderSourceFactory{
		repo:         repo,
		classMapping: make(map[model.ProviderSourceType]ProviderSourceClass),
	}
}

// SetDatabase sets the database for the factory
// This is called after factory creation to provide database access
func (f *ProviderSourceFactory) SetDatabase(db interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.db = db
}

// getDatabase returns the database (with lock protection)
func (f *ProviderSourceFactory) getDatabase() interface{} {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.db
}

// RegisterProviderSourceClass registers a provider source class
// This replaces Python's introspection-based class discovery
func (f *ProviderSourceFactory) RegisterProviderSourceClass(class ProviderSourceClass) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.classMapping[class.Type()] = class
}

// GetProviderClasses returns all registered provider source classes
// Python reference: factory.py::get_provider_classes()
func (f *ProviderSourceFactory) GetProviderClasses() map[model.ProviderSourceType]ProviderSourceClass {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[model.ProviderSourceType]ProviderSourceClass, len(f.classMapping))
	for k, v := range f.classMapping {
		result[k] = v
	}
	return result
}

// GetProviderSourceClassByType returns the provider source class for a given type
// Python reference: factory.py::get_provider_source_class_by_type()
func (f *ProviderSourceFactory) GetProviderSourceClassByType(type_ model.ProviderSourceType) ProviderSourceClass {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.classMapping[type_]
}

// GetProviderSourceByName retrieves a provider source by its display name
// Python reference: factory.py::get_provider_source_by_name()
func (f *ProviderSourceFactory) GetProviderSourceByName(ctx context.Context, name string) (ProviderSourceInstance, error) {
	source, err := f.repo.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider source by name: %w", err)
	}
	if source == nil {
		return nil, nil
	}

	return &providerSourceInstanceImpl{
		source:  source,
		factory: f,
	}, nil
}

// GetProviderSourceByApiName retrieves a provider source by its API-friendly name
// Python reference: factory.py::get_provider_source_by_api_name()
func (f *ProviderSourceFactory) GetProviderSourceByApiName(ctx context.Context, apiName string) (ProviderSourceInstance, error) {
	source, err := f.repo.FindByApiName(ctx, apiName)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider source by api_name: %w", err)
	}
	if source == nil {
		return nil, nil
	}

	return &providerSourceInstanceImpl{
		source:  source,
		factory: f,
	}, nil
}

// GetAllProviderSources retrieves all provider sources from the database
// Python reference: factory.py::get_all_provider_sources()
func (f *ProviderSourceFactory) GetAllProviderSources(ctx context.Context) ([]ProviderSourceInstance, error) {
	sources, err := f.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all provider sources: %w", err)
	}

	result := make([]ProviderSourceInstance, len(sources))
	for i, source := range sources {
		result[i] = &providerSourceInstanceImpl{source: source, factory: f}
	}
	return result, nil
}

// NameToApiName converts a display name to an API-friendly name
// Matches Python behavior exactly
// Python reference: factory.py::_name_to_api_name()
func (f *ProviderSourceFactory) NameToApiName(name string) string {
	if name == "" {
		return ""
	}

	// Convert to lower case
	name = strings.ToLower(name)

	// Replace spaces with dashes
	name = strings.ReplaceAll(name, " ", "-")

	// Remove any non alphanumeric/dashes
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	name = reg.ReplaceAllString(name, "")

	// Remove any leading/trailing dashes
	name = strings.Trim(name, "-")

	return name
}

// InitialiseFromConfig loads provider sources from config into database
// Python reference: factory.py::initialise_from_config()
func (f *ProviderSourceFactory) InitialiseFromConfig(ctx context.Context, configJSON string) error {
	// Parse JSON config
	var providerSourceConfigs []map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &providerSourceConfigs); err != nil {
		return model.NewInvalidProviderSourceConfigError("Provider source config is not a valid JSON list of objects")
	}

	seenNames := make(map[string]bool)

	for _, providerSourceConfig := range providerSourceConfigs {
		// Validate config is a dict
		if providerSourceConfig == nil {
			return model.NewInvalidProviderSourceConfigError("Provider source config is not a valid JSON list of objects")
		}

		// Validate required attributes
		nameVal, nameOk := providerSourceConfig["name"]
		if !nameOk {
			return model.NewInvalidProviderSourceConfigError("Provider source config does not contain required attribute: name")
		}

		name, ok := nameVal.(string)
		if !ok || name == "" {
			return model.NewInvalidProviderSourceConfigError("Provider source name is empty")
		}

		// Check for duplicates (case-insensitive)
		nameLower := strings.ToLower(name)
		if seenNames[nameLower] {
			return model.NewInvalidProviderSourceConfigError(fmt.Sprintf("Duplicate Provider Source name found: %s", name))
		}
		seenNames[nameLower] = true

		// Validate and generate api_name
		apiName := f.NameToApiName(name)
		if apiName == "" {
			return model.NewInvalidProviderSourceConfigError("Invalid provider source config: Name must contain some alphanumeric characters")
		}

		// Validate type attribute
		typeVal, typeOk := providerSourceConfig["type"]
		if !typeOk {
			return model.NewInvalidProviderSourceConfigError("Provider source config does not contain required attribute: type")
		}

		typeStr, ok := typeVal.(string)
		if !ok {
			return model.NewInvalidProviderSourceConfigError("Invalid provider source type")
		}

		// Convert type string to ProviderSourceType
		type_ := model.ProviderSourceType(strings.ToLower(typeStr))
		if !type_.IsValid() {
			return model.NewInvalidProviderSourceConfigError(
				fmt.Sprintf("Invalid provider source type. Valid types: %s", model.GetValidProviderTypes()),
			)
		}

		// Get provider source class
		providerSourceClass := f.GetProviderSourceClassByType(type_)
		if providerSourceClass == nil {
			return model.NewInvalidProviderSourceConfigError(
				fmt.Sprintf("Internal Exception, could not find class for %s", type_),
			)
		}

		// Generate DB config from source config
		// Python reference: factory.py line 179 passes the full provider_source_config dict
		// The dict has keys: name, type, and all the config values at the top level
		providerDBConfig, err := providerSourceClass.GenerateDBConfigFromSourceConfig(providerSourceConfig)
		if err != nil {
			// Convert to InvalidProviderSourceConfigError if needed
			if _, ok := err.(*model.InvalidProviderSourceConfigError); ok {
				return err
			}
			return model.NewInvalidProviderSourceConfigError(err.Error())
		}

		// Create or update provider source
		source := model.NewProviderSource(name, apiName, type_, providerDBConfig)

		// Upsert to database
		if err := f.repo.Upsert(ctx, source); err != nil {
			return fmt.Errorf("failed to upsert provider source: %w", err)
		}
	}

	return nil
}

// ProviderSourceInstance represents an instance of a provider source
// This is a simplified interface for the factory's return values
// Python reference: provider_source/base.py::BaseProviderSource
type ProviderSourceInstance interface {
	Name() string
	ApiName(ctx context.Context) (string, error)
	Type() model.ProviderSourceType

	// OAuth methods for GitHub provider sources
	// These methods are only available for certain provider source types
	GetLoginRedirectURL(ctx context.Context) (string, error)
	GetUserAccessToken(ctx context.Context, code string) (string, error)
	GetUsername(ctx context.Context, accessToken string) (string, error)
	GetUserOrganizations(ctx context.Context, accessToken string) []string

	// Provider source API methods for GitHub integration
	// These methods implement the organization/repository management endpoints

	// GetUserOrganizationsList returns organizations with full details for authenticated user
	// Python reference: github_organisations.py
	GetUserOrganizationsList(ctx context.Context, sessionID string) ([]*model.Organization, error)

	// GetUserRepositories returns repositories for the authenticated user
	// Python reference: github_repositories.py
	GetUserRepositories(ctx context.Context, sessionID string) ([]*model.Repository, error)

	// RefreshNamespaceRepositories refreshes repositories for a given namespace
	// Python reference: github_refresh_namespace.py
	RefreshNamespaceRepositories(ctx context.Context, namespace string) error

	// PublishProviderFromRepository publishes a provider from a repository
	// Python reference: github_repository_publish_provider.py
	PublishProviderFromRepository(ctx context.Context, repoID int, categoryID int, namespace string) (*PublishProviderResult, error)

	// Provider extraction methods for downloading release artifacts
	// NOTE: These methods require repository context and access token for GitHub API calls

	// GetReleaseArtifact downloads a specific release artifact by name
	// Python reference: provider_extractor.py::ProviderExtractor::_download_artifact
	GetReleaseArtifact(ctx context.Context, repo *sqldb.RepositoryDB, artifact *model.ReleaseArtifactMetadata, accessToken string) ([]byte, error)

	// GetReleaseArchive downloads the release source archive (.tar.gz)
	// Returns the archive data and the subdirectory to extract (if any)
	// Python reference: provider_extractor.py::ProviderExtractor::_obtain_source_code
	GetReleaseArchive(ctx context.Context, repo *sqldb.RepositoryDB, releaseMetadata *model.RepositoryReleaseMetadata, accessToken string) ([]byte, string, error)
}

// PublishProviderResult contains the result of publishing a provider from a repository
type PublishProviderResult struct {
	Name      string
	Namespace string
}

// providerSourceInstanceImpl implements ProviderSourceInstance
type providerSourceInstanceImpl struct {
	source  *model.ProviderSource
	factory *ProviderSourceFactory
}

func (p *providerSourceInstanceImpl) Name() string {
	return p.source.Name()
}

func (p *providerSourceInstanceImpl) ApiName(ctx context.Context) (string, error) {
	return p.source.ApiName(), nil
}

func (p *providerSourceInstanceImpl) Type() model.ProviderSourceType {
	return p.source.Type()
}

// GetLoginRedirectURL returns the OAuth login redirect URL
func (p *providerSourceInstanceImpl) GetLoginRedirectURL(ctx context.Context) (string, error) {
	impl, err := p.factory.getProviderSourceImplementation(ctx, p.source)
	if err != nil {
		return "", err
	}
	return impl.GetLoginRedirectURL(ctx)
}

// GetUserAccessToken exchanges the OAuth code for an access token
func (p *providerSourceInstanceImpl) GetUserAccessToken(ctx context.Context, code string) (string, error) {
	impl, err := p.factory.getProviderSourceImplementation(ctx, p.source)
	if err != nil {
		return "", err
	}
	return impl.GetUserAccessToken(ctx, code)
}

// GetUsername gets the username from the access token
func (p *providerSourceInstanceImpl) GetUsername(ctx context.Context, accessToken string) (string, error) {
	impl, err := p.factory.getProviderSourceImplementation(ctx, p.source)
	if err != nil {
		return "", err
	}
	return impl.GetUsername(ctx, accessToken)
}

// GetUserOrganizations gets the organizations for the user
func (p *providerSourceInstanceImpl) GetUserOrganizations(ctx context.Context, accessToken string) []string {
	impl, err := p.factory.getProviderSourceImplementation(ctx, p.source)
	if err != nil {
		return []string{}
	}
	return impl.GetUserOrganizations(ctx, accessToken)
}

// GetUserOrganizationsList returns organizations with full details for authenticated user
func (p *providerSourceInstanceImpl) GetUserOrganizationsList(ctx context.Context, sessionID string) ([]*model.Organization, error) {
	impl, err := p.factory.getProviderSourceImplementation(ctx, p.source)
	if err != nil {
		return nil, err
	}
	return impl.GetUserOrganizationsList(ctx, sessionID)
}

// GetUserRepositories returns repositories for the authenticated user
func (p *providerSourceInstanceImpl) GetUserRepositories(ctx context.Context, sessionID string) ([]*model.Repository, error) {
	impl, err := p.factory.getProviderSourceImplementation(ctx, p.source)
	if err != nil {
		return nil, err
	}
	return impl.GetUserRepositories(ctx, sessionID)
}

// RefreshNamespaceRepositories refreshes repositories for a given namespace
func (p *providerSourceInstanceImpl) RefreshNamespaceRepositories(ctx context.Context, namespace string) error {
	impl, err := p.factory.getProviderSourceImplementation(ctx, p.source)
	if err != nil {
		return err
	}
	return impl.RefreshNamespaceRepositories(ctx, namespace)
}

// PublishProviderFromRepository publishes a provider from a repository
func (p *providerSourceInstanceImpl) PublishProviderFromRepository(ctx context.Context, repoID int, categoryID int, namespace string) (*PublishProviderResult, error) {
	impl, err := p.factory.getProviderSourceImplementation(ctx, p.source)
	if err != nil {
		return nil, err
	}
	return impl.PublishProviderFromRepository(ctx, repoID, categoryID, namespace)
}

// GetReleaseArtifact downloads a specific release artifact by name
// Python reference: provider_extractor.py::ProviderExtractor::_download_artifact
func (p *providerSourceInstanceImpl) GetReleaseArtifact(ctx context.Context, repo *sqldb.RepositoryDB, artifact *model.ReleaseArtifactMetadata, accessToken string) ([]byte, error) {
	impl, err := p.factory.getProviderSourceImplementation(ctx, p.source)
	if err != nil {
		return nil, err
	}
	return impl.GetReleaseArtifact(ctx, repo, artifact, accessToken)
}

// GetReleaseArchive downloads the release source archive (.tar.gz)
// Returns the archive data and the subdirectory to extract (if any)
// Python reference: provider_extractor.py::ProviderExtractor::_obtain_source_code
func (p *providerSourceInstanceImpl) GetReleaseArchive(ctx context.Context, repo *sqldb.RepositoryDB, releaseMetadata *model.RepositoryReleaseMetadata, accessToken string) ([]byte, string, error) {
	impl, err := p.factory.getProviderSourceImplementation(ctx, p.source)
	if err != nil {
		return nil, "", err
	}
	return impl.GetReleaseArchive(ctx, repo, releaseMetadata, accessToken)
}

// getProviderSourceImplementation gets the actual provider source implementation
// This creates the concrete implementation (e.g., GithubProviderSource) from the model
func (f *ProviderSourceFactory) getProviderSourceImplementation(ctx context.Context, source *model.ProviderSource) (ProviderSourceInstance, error) {
	// Get the provider source class from factory
	sourceClass := f.GetProviderSourceClassByType(source.Type())
	if sourceClass == nil {
		return nil, fmt.Errorf("provider source class not registered for type: %s", source.Type())
	}

	// Use the class's CreateInstance method to create the actual implementation
	return sourceClass.CreateInstance(source.Name(), f.repo, f.getDatabase())
}

// providerSourceInstanceAdapter adapts a ProviderSource model to ProviderSourceInstance interface
// This is a bridge between the domain model and the factory interface
type providerSourceInstanceAdapter struct {
	source *model.ProviderSource
}

func (a *providerSourceInstanceAdapter) Name() string {
	return a.source.Name()
}

func (a *providerSourceInstanceAdapter) ApiName(ctx context.Context) (string, error) {
	return a.source.ApiName(), nil
}

func (a *providerSourceInstanceAdapter) Type() model.ProviderSourceType {
	return a.source.Type()
}

// OAuth methods - these require the actual GithubProviderSource implementation
// which contains the HTTP infrastructure logic

// GetLoginRedirectURL returns the OAuth login redirect URL
// Python reference: github.py::get_login_redirect_url
func (a *providerSourceInstanceAdapter) GetLoginRedirectURL(ctx context.Context) (string, error) {
	return "", fmt.Errorf("OAuth methods require GithubProviderSource instance")
}

// GetUserAccessToken exchanges the OAuth code for an access token
// Python reference: github.py::get_user_access_token
func (a *providerSourceInstanceAdapter) GetUserAccessToken(ctx context.Context, code string) (string, error) {
	return "", fmt.Errorf("OAuth methods require GithubProviderSource instance")
}

// GetUsername gets the username from the access token
// Python reference: github.py::get_username
func (a *providerSourceInstanceAdapter) GetUsername(ctx context.Context, accessToken string) (string, error) {
	return "", fmt.Errorf("OAuth methods require GithubProviderSource instance")
}

// GetUserOrganizations gets the organizations for the user
// Python reference: github.py::get_user_organisations
func (a *providerSourceInstanceAdapter) GetUserOrganizations(ctx context.Context, accessToken string) []string {
	return []string{}
}

// GetUserOrganizationsList returns organizations with full details for authenticated user
// Python reference: github_organisations.py
func (a *providerSourceInstanceAdapter) GetUserOrganizationsList(ctx context.Context, sessionID string) ([]*model.Organization, error) {
	return []*model.Organization{}, nil
}

// GetUserRepositories returns repositories for the authenticated user
// Python reference: github_repositories.py
func (a *providerSourceInstanceAdapter) GetUserRepositories(ctx context.Context, sessionID string) ([]*model.Repository, error) {
	return []*model.Repository{}, nil
}

// RefreshNamespaceRepositories refreshes repositories for a given namespace
// Python reference: github_refresh_namespace.py
func (a *providerSourceInstanceAdapter) RefreshNamespaceRepositories(ctx context.Context, namespace string) error {
	return nil
}

// PublishProviderFromRepository publishes a provider from a repository
// Python reference: github_repository_publish_provider.py
func (a *providerSourceInstanceAdapter) PublishProviderFromRepository(ctx context.Context, repoID int, categoryID int, namespace string) (*PublishProviderResult, error) {
	return nil, fmt.Errorf("publish provider from repository not yet implemented")
}

// GetReleaseArtifact downloads a specific release artifact by name
// Python reference: provider_extractor.py::ProviderExtractor::_download_artifact
func (a *providerSourceInstanceAdapter) GetReleaseArtifact(ctx context.Context, releaseMetadata *model.RepositoryReleaseMetadata, artifactName string) ([]byte, error) {
	return nil, fmt.Errorf("release artifact download requires actual provider source implementation")
}

// GetReleaseArchive downloads the release source archive (.tar.gz)
// Returns the archive data and the subdirectory to extract (if any)
// Python reference: provider_extractor.py::ProviderExtractor::_obtain_source_code
func (a *providerSourceInstanceAdapter) GetReleaseArchive(ctx context.Context, repo *sqldb.RepositoryDB, releaseMetadata *model.RepositoryReleaseMetadata, accessToken string) ([]byte, string, error) {
	return nil, "", fmt.Errorf("release archive download requires actual provider source implementation")
}
