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

var (
	instance *ProviderSourceFactory
	once     sync.Once
)

// Get returns the singleton instance of ProviderSourceFactory
// Python reference: factory.py::get()
func Get() *ProviderSourceFactory {
	// Note: This is a simplified singleton that assumes Initialize has been called
	// In production, the factory should be initialized through the container
	return instance
}

// NewProviderSourceFactory creates a new ProviderSourceFactory
// This should be called during container initialization
func NewProviderSourceFactory(repo repository.ProviderSourceRepository) *ProviderSourceFactory {
	factory := &ProviderSourceFactory{
		repo:         repo,
		classMapping: make(map[model.ProviderSourceType]ProviderSourceClass),
	}

	// Store singleton instance
	once.Do(func() {
		instance = factory
	})

	return factory
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
		source: source,
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
		source: source,
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
		result[i] = &providerSourceInstanceImpl{source: source}
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
}

// providerSourceInstanceImpl implements ProviderSourceInstance
type providerSourceInstanceImpl struct {
	source *model.ProviderSource
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
