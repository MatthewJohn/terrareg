package config

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	providerSourceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
)

// ConfigRepositoryImpl implements the ConfigRepository interface
type ConfigRepositoryImpl struct {
	versionReader      *version.VersionReader
	domainConfig       *model.DomainConfig
	providerSourceFactory *providerSourceService.ProviderSourceFactory
}

// NewConfigRepositoryImpl creates a new ConfigRepositoryImpl with injected dependencies
func NewConfigRepositoryImpl(
	versionReader *version.VersionReader,
	domainConfig *model.DomainConfig,
	providerSourceFactory *providerSourceService.ProviderSourceFactory,
) *ConfigRepositoryImpl {
	if domainConfig == nil {
		panic("ConfigRepositoryImpl requires domain config to be injected")
	}
	return &ConfigRepositoryImpl{
		versionReader:      versionReader,
		domainConfig:       domainConfig,
		providerSourceFactory: providerSourceFactory,
	}
}

// GetConfig retrieves the UI configuration from injected domain config
func (r *ConfigRepositoryImpl) GetConfig(ctx context.Context) (*model.UIConfig, error) {
	if r.domainConfig == nil {
		return nil, fmt.Errorf("ConfigRepositoryImpl requires domain config to be injected")
	}

	// Get provider sources
	providerSources, err := r.getProviderSources()
	if err != nil {
		return nil, fmt.Errorf("failed to get provider sources: %w", err)
	}

	// Convert domain config to UI config
	config := &model.UIConfig{
		// Use domain config values directly
		TrustedNamespaceLabel:     r.domainConfig.TrustedNamespaceLabel,
		ContributedNamespaceLabel: r.domainConfig.ContributedNamespaceLabel,
		VerifiedModuleLabel:       r.domainConfig.VerifiedModuleLabel,

		// Analytics configuration
		AnalyticsTokenPhrase:      r.domainConfig.AnalyticsTokenPhrase,
		AnalyticsTokenDescription: r.domainConfig.AnalyticsTokenDescription,
		ExampleAnalyticsToken:     r.domainConfig.ExampleAnalyticsToken,
		DisableAnalytics:          r.domainConfig.DisableAnalytics,

		// Feature flags
		AllowModuleHosting:              r.domainConfig.AllowModuleHosting,
		UploadAPIKeysEnabled:            r.domainConfig.UploadAPIKeysEnabled,
		PublishAPIKeysEnabled:           r.domainConfig.PublishAPIKeysEnabled,
		DisableTerraregExclusiveLabels:  r.domainConfig.DisableTerraregExclusiveLabels,
		AllowCustomGitURLModuleProvider: r.domainConfig.AllowCustomGitURLModuleProvider,
		AllowCustomGitURLModuleVersion:  r.domainConfig.AllowCustomGitURLModuleVersion,
		SecretKeySet:                    r.domainConfig.SecretKeySet,

		// Authentication status - from domain config
		OpenIDConnectEnabled:   r.domainConfig.OpenIDConnectEnabled,
		OpenIDConnectLoginText: r.domainConfig.OpenIDConnectLoginText,
		SAMLEnabled:            r.domainConfig.SAMLEnabled,
		SAMLLoginText:          r.domainConfig.SAMLLoginText,
		AdminLoginEnabled:      r.domainConfig.AdminLoginEnabled,

		// UI configuration
		AdditionalModuleTabs:     r.domainConfig.AdditionalModuleTabs,
		AutoCreateNamespace:      r.domainConfig.AutoCreateNamespace,
		AutoCreateModuleProvider: r.domainConfig.AutoCreateModuleProvider,
		DefaultUiDetailsView:     string(r.domainConfig.DefaultUiDetailsView),

		// Provider sources
		ProviderSources: providerSources,
	}

	return config, nil
}

// GetVersion retrieves the application version
func (r *ConfigRepositoryImpl) GetVersion(ctx context.Context) (string, error) {
	return r.versionReader.ReadVersion()
}

// IsOpenIDConnectEnabled checks if OpenID Connect authentication is enabled
func (r *ConfigRepositoryImpl) IsOpenIDConnectEnabled(ctx context.Context) bool {
	return r.domainConfig.OpenIDConnectEnabled
}

// IsSAMLEnabled checks if SAML authentication is enabled
func (r *ConfigRepositoryImpl) IsSAMLEnabled(ctx context.Context) bool {
	return r.domainConfig.SAMLEnabled
}

// IsAdminLoginEnabled checks if admin login is enabled
func (r *ConfigRepositoryImpl) IsAdminLoginEnabled(ctx context.Context) bool {
	return r.domainConfig.AdminLoginEnabled
}

// GetProviderSources retrieves all configured provider sources
func (r *ConfigRepositoryImpl) GetProviderSources(ctx context.Context) ([]model.ProviderSource, error) {
	return r.getProviderSources()
}

// getProviderSources loads provider sources from the database via ProviderSourceFactory
func (r *ConfigRepositoryImpl) getProviderSources() ([]model.ProviderSource, error) {
	// If factory is not injected, fall back to empty list
	if r.providerSourceFactory == nil {
		return []model.ProviderSource{}, nil
	}

	ctx := context.Background()
	sources, err := r.providerSourceFactory.GetAllProviderSources(ctx)
	if err != nil {
		return nil, err
	}

	// Convert domain provider sources to UI model
	providerSources := make([]model.ProviderSource, 0, len(sources))
	for _, source := range sources {
		name := source.Name()
		loginButtonText, _ := source.LoginButtonText(ctx)
		providerSources = append(providerSources, model.ProviderSource{
			Name:            name,
			APIName:         name,
			LoginButtonText: loginButtonText,
		})
	}
	return providerSources, nil
}
