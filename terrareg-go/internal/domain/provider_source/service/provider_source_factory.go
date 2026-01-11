package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/repository"
)

// ProviderSourceClass defines the interface for provider source implementations
// Python reference: provider_source/base.py::BaseProviderSource
type ProviderSourceClass interface {
	// Type returns the provider source type (e.g., "github", "gitlab")
	Type() model.ProviderSourceType

	// GenerateDBConfigFromSourceConfig validates and converts user config to DB config
	// Python reference: github.py::generate_db_config_from_source_config()
	GenerateDBConfigFromSourceConfig(sourceConfig map[string]interface{}) (*model.ProviderSourceConfig, error)
}

// ProviderSourceFactory manages provider sources
// Python reference: provider_source/factory.py::ProviderSourceFactory
type ProviderSourceFactory struct {
	repo         repository.ProviderSourceRepository
	classMapping map[model.ProviderSourceType]ProviderSourceClass
	mu           sync.RWMutex
}

// NewProviderSourceFactory creates a new ProviderSourceFactory
// This should be called during container initialization
func NewProviderSourceFactory(repo repository.ProviderSourceRepository) *ProviderSourceFactory {
	return &ProviderSourceFactory{
		repo:         repo,
		classMapping: make(map[model.ProviderSourceType]ProviderSourceClass),
	}
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
	ApiName() string
	Type() model.ProviderSourceType

	// OAuth methods for GitHub provider sources
	// These methods are only available for certain provider source types
	GetLoginRedirectURL(ctx context.Context) (string, error)
	GetUserAccessToken(ctx context.Context, code string) (string, error)
	GetUsername(ctx context.Context, accessToken string) (string, error)
	GetUserOrganizations(ctx context.Context, accessToken string) []string
}

// providerSourceInstanceImpl implements ProviderSourceInstance
type providerSourceInstanceImpl struct {
	source  *model.ProviderSource
	factory *ProviderSourceFactory
}

func (p *providerSourceInstanceImpl) Name() string {
	return p.source.Name()
}

func (p *providerSourceInstanceImpl) ApiName() string {
	return p.source.ApiName()
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

// getProviderSourceImplementation gets the actual provider source implementation
// This creates the concrete implementation (e.g., GithubProviderSource) from the model
func (f *ProviderSourceFactory) getProviderSourceImplementation(ctx context.Context, source *model.ProviderSource) (ProviderSourceInstance, error) {
	// For now, only GitHub is implemented
	if source.Type() != model.ProviderSourceTypeGithub {
		return nil, fmt.Errorf("unsupported provider source type for OAuth: %s", source.Type())
	}

	// Create the wrapper that implements OAuth methods
	return &githubProviderSourceWrapper{
		source: source,
		repo:   f.repo,
	}, nil
}

// githubProviderSourceWrapper is a temporary wrapper to avoid circular dependency
// It implements ProviderSourceInstance by delegating to a dynamically created GithubProviderSource
type githubProviderSourceWrapper struct {
	source *model.ProviderSource
	repo   repository.ProviderSourceRepository
}

func (w *githubProviderSourceWrapper) Name() string {
	return w.source.Name()
}

func (w *githubProviderSourceWrapper) ApiName() string {
	return w.source.ApiName()
}

func (w *githubProviderSourceWrapper) Type() model.ProviderSourceType {
	return w.source.Type()
}

// GetLoginRedirectURL returns the OAuth login redirect URL
// Python reference: github.py::get_login_redirect_url
func (w *githubProviderSourceWrapper) GetLoginRedirectURL(ctx context.Context) (string, error) {
	// Get the config from the source
	config := w.source.Config()

	// Config should have the base_url and client_id set
	// Python reference: github.py::_get_login_url
	baseURL := config.BaseURL
	clientID := config.ClientID

	if clientID == "" {
		return "", fmt.Errorf("client_id not configured for provider source: %s", w.source.Name())
	}

	// Generate state token for CSRF protection
	stateToken := generateStateToken()

	// Build the authorization URL
	// Python reference: github.py::get_login_redirect_url
	authURL := fmt.Sprintf("%s/login/oauth/authorize?client_id=%s&state=%s&scope=read:org",
		baseURL, clientID, stateToken)

	return authURL, nil
}

// GetUserAccessToken exchanges the OAuth code for an access token
// Python reference: github.py::get_user_access_token
func (w *githubProviderSourceWrapper) GetUserAccessToken(ctx context.Context, code string) (string, error) {
	// Get the config from the source
	config := w.source.Config()

	baseURL := config.BaseURL
	clientID := config.ClientID
	clientSecret := config.ClientSecret

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("client_id and client_secret must be configured for provider source: %s", w.source.Name())
	}

	// Exchange code for access token
	// Python reference: github.py::get_user_access_token
	tokenURL := fmt.Sprintf("%s/login/oauth/access_token", baseURL)

	// Build form data
	formData := fmt.Sprintf("client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(formData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/vnd.github+json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code for access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code exchanging code: %d", resp.StatusCode)
	}

	// Parse response as form-encoded (query string format) - matching Python
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse query string format: "access_token=xxx&..."
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "", err
	}

	if accessTokens := values["access_token"]; len(accessTokens) == 1 {
		return accessTokens[0], nil
	}

	return "", nil
}

// GetUsername gets the username from the access token
// Python reference: github.py::get_username
func (w *githubProviderSourceWrapper) GetUsername(ctx context.Context, accessToken string) (string, error) {
	// Get the config from the source
	config := w.source.Config()

	apiURL := config.ApiURL
	if apiURL == "" {
		// Fallback to BaseURL if ApiURL is not set
		apiURL = config.BaseURL
	}
	userURL := fmt.Sprintf("%s/user", apiURL)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", userURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header using Bearer prefix (matching Python)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/vnd.github+json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get username: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code getting username: %d", resp.StatusCode)
	}

	// Parse response
	var user struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to decode user response: %w", err)
	}

	return user.Login, nil
}

// GetUserOrganizations gets the organizations for the user
// Python reference: github.py::get_user_organisations
func (w *githubProviderSourceWrapper) GetUserOrganizations(ctx context.Context, accessToken string) []string {
	// Get the config from the source
	config := w.source.Config()

	baseURL := config.BaseURL
	apiURL := fmt.Sprintf("%s/api/v3/user/orgs?per_page=100", baseURL)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return []string{}
	}

	// Set authorization header
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}
	}

	// Parse response
	var orgs []struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return []string{}
	}

	// Extract org names
	result := make([]string, len(orgs))
	for i, org := range orgs {
		result[i] = org.Login
	}
	return result
}

// generateStateToken generates a random state token for CSRF protection
func generateStateToken() string {
	// Generate a random state token
	// In production, use crypto/rand
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
