package config

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
)

// ConfigRepositoryImpl implements the ConfigRepository interface
type ConfigRepositoryImpl struct {
	versionReader *version.VersionReader
	domainConfig  *model.DomainConfig
}

// NewConfigRepositoryImpl creates a new ConfigRepositoryImpl with injected domain config
func NewConfigRepositoryImpl(versionReader *version.VersionReader, domainConfig *model.DomainConfig) *ConfigRepositoryImpl {
	if domainConfig == nil {
		panic("ConfigRepositoryImpl requires domain config to be injected")
	}
	return &ConfigRepositoryImpl{
		versionReader: versionReader,
		domainConfig:  domainConfig,
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

// getProviderSources parses provider source configuration from environment variables
func (r *ConfigRepositoryImpl) getProviderSources() ([]model.ProviderSource, error) {
	// For now, return empty provider sources
	// This would need to be implemented based on the provider source factory pattern
	// from the Python version
	return []model.ProviderSource{}, nil
}
