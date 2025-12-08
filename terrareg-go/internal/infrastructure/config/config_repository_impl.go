package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
)

// ConfigRepositoryImpl implements the ConfigRepository interface
type ConfigRepositoryImpl struct {
	versionReader *version.VersionReader
	domainConfig  *model.DomainConfig
}

// NewConfigRepositoryImpl creates a new ConfigRepositoryImpl
func NewConfigRepositoryImpl(versionReader *version.VersionReader) *ConfigRepositoryImpl {
	return &ConfigRepositoryImpl{
		versionReader: versionReader,
	}
}

// NewConfigRepositoryImplWithConfig creates a new ConfigRepositoryImpl with injected domain config
func NewConfigRepositoryImplWithConfig(versionReader *version.VersionReader, domainConfig *model.DomainConfig) *ConfigRepositoryImpl {
	return &ConfigRepositoryImpl{
		versionReader: versionReader,
		domainConfig:  domainConfig,
	}
}

// GetConfig retrieves the complete configuration from injected domain config
func (r *ConfigRepositoryImpl) GetConfig(ctx context.Context) (*model.Config, error) {
	var config *model.Config

	if r.domainConfig != nil {
		// Use injected domain config
		providerSources, err := r.getProviderSources()
		if err != nil {
			return nil, fmt.Errorf("failed to get provider sources: %w", err)
		}

		config = &model.Config{
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

			// Authentication status - these will be checked dynamically
			OpenIDConnectEnabled:   r.IsOpenIDConnectEnabled(ctx),
			OpenIDConnectLoginText: r.getEnvString("OPENID_CONNECT_LOGIN_TEXT", "Login with OpenID"),
			SAMLEnabled:            r.IsSAMLEnabled(ctx),
			SAMLLoginText:          r.getEnvString("SAML2_LOGIN_TEXT", "Login with SAML"),
			AdminLoginEnabled:      r.IsAdminLoginEnabled(ctx),

			// UI configuration
			AdditionalModuleTabs:     r.domainConfig.AdditionalModuleTabs,
			AutoCreateNamespace:      r.domainConfig.AutoCreateNamespace,
			AutoCreateModuleProvider: r.domainConfig.AutoCreateModuleProvider,
			DefaultUiDetailsView:     string(r.domainConfig.DefaultUiDetailsView),

			// Provider sources
			ProviderSources: providerSources,
		}
	} else {
		// Fallback to environment variable loading (legacy)
		providerSources, err := r.getProviderSources()
		if err != nil {
			return nil, fmt.Errorf("failed to get provider sources: %w", err)
		}

		// Parse additional module tabs
		additionalModuleTabs := r.parseAdditionalModuleTabs()

		config = &model.Config{
			// Namespace labels
			TrustedNamespaceLabel:     r.getEnvString("TRUSTED_NAMESPACE_LABEL", "Trusted"),
			ContributedNamespaceLabel: r.getEnvString("CONTRIBUTED_NAMESPACE_LABEL", "Contributed"),
			VerifiedModuleLabel:       r.getEnvString("VERIFIED_MODULE_LABEL", "Verified"),

			// Analytics configuration
			AnalyticsTokenPhrase:      r.getEnvString("ANALYTICS_TOKEN_PHRASE", "analytics token"),
			AnalyticsTokenDescription: r.getEnvString("ANALYTICS_TOKEN_DESCRIPTION", ""),
			ExampleAnalyticsToken:     r.getEnvString("EXAMPLE_ANALYTICS_TOKEN", ""),
			DisableAnalytics:          r.getEnvBool("DISABLE_ANALYTICS", false),

			// Feature flags
			AllowModuleHosting:              r.getModuleHostingMode(),
			UploadAPIKeysEnabled:            r.getEnvString("UPLOAD_API_KEYS", "") != "",
			PublishAPIKeysEnabled:           r.getEnvString("PUBLISH_API_KEYS", "") != "",
			DisableTerraregExclusiveLabels:  r.getEnvBool("DISABLE_TERRAREG_EXCLUSIVE_LABELS", false),
			AllowCustomGitURLModuleProvider: r.getEnvBool("ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER", true),
			AllowCustomGitURLModuleVersion:  r.getEnvBool("ALLOW_CUSTOM_GIT_URL_MODULE_VERSION", true),
			SecretKeySet:                    r.getEnvString("SECRET_KEY", "") != "",

			// Authentication status - these will be checked dynamically
			OpenIDConnectEnabled:   r.IsOpenIDConnectEnabled(ctx),
			OpenIDConnectLoginText: r.getEnvString("OPENID_CONNECT_LOGIN_TEXT", "Login with OpenID"),
			SAMLEnabled:            r.IsSAMLEnabled(ctx),
			SAMLLoginText:          r.getEnvString("SAML2_LOGIN_TEXT", "Login with SAML"),
			AdminLoginEnabled:      r.IsAdminLoginEnabled(ctx),

			// UI configuration
			AdditionalModuleTabs:     additionalModuleTabs,
			AutoCreateNamespace:      r.getEnvBool("AUTO_CREATE_NAMESPACE", true),
			AutoCreateModuleProvider: r.getEnvBool("AUTO_CREATE_MODULE_PROVIDER", true),
			DefaultUiDetailsView:     r.getEnvString("DEFAULT_UI_DETAILS_VIEW", "table"),

			// Provider sources
			ProviderSources: providerSources,
		}
	}

	return config, nil
}

// GetVersion retrieves the application version
func (r *ConfigRepositoryImpl) GetVersion(ctx context.Context) (string, error) {
	return r.versionReader.ReadVersion()
}

// IsOpenIDConnectEnabled checks if OpenID Connect authentication is enabled
func (r *ConfigRepositoryImpl) IsOpenIDConnectEnabled(ctx context.Context) bool {
	clientID := r.getEnvString("OPENID_CONNECT_CLIENT_ID", "")
	issuer := r.getEnvString("OPENID_CONNECT_ISSUER", "")
	return clientID != "" && issuer != ""
}

// IsSAMLEnabled checks if SAML authentication is enabled
func (r *ConfigRepositoryImpl) IsSAMLEnabled(ctx context.Context) bool {
	metadataURL := r.getEnvString("SAML2_IDP_METADATA_URL", "")
	return metadataURL != ""
}

// IsAdminLoginEnabled checks if admin login is enabled
func (r *ConfigRepositoryImpl) IsAdminLoginEnabled(ctx context.Context) bool {
	adminToken := r.getEnvString("ADMIN_AUTHENTICATION_TOKEN", "")
	return adminToken != ""
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

// parseAdditionalModuleTabs parses the ADDITIONAL_MODULE_TABS environment variable
func (r *ConfigRepositoryImpl) parseAdditionalModuleTabs() []string {
	tabsStr := r.getEnvString("ADDITIONAL_MODULE_TABS", "")
	if tabsStr == "" {
		return []string{}
	}

	tabs := strings.Split(tabsStr, ",")
	result := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		tab = strings.TrimSpace(tab)
		if tab != "" {
			result = append(result, tab)
		}
	}
	return result
}

// Helper functions for environment variable access
func (r *ConfigRepositoryImpl) getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (r *ConfigRepositoryImpl) getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		lowerValue := strings.ToLower(value)
		return lowerValue == "true" || lowerValue == "1" || lowerValue == "yes"
	}
	return defaultValue
}

func (r *ConfigRepositoryImpl) getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getModuleHostingMode parses the ALLOW_MODULE_HOSTING environment variable
// and returns the corresponding ModuleHostingMode enum value
func (r *ConfigRepositoryImpl) getModuleHostingMode() model.ModuleHostingMode {
	value := strings.ToLower(r.getEnvString("ALLOW_MODULE_HOSTING", "true"))

	// Validate the value against allowed enum values
	switch value {
	case "true":
		return model.ModuleHostingModeAllow
	case "false":
		return model.ModuleHostingModeDisallow
	case "enforce":
		return model.ModuleHostingModeEnforce
	default:
		// Default to "true" if invalid value provided
		return model.ModuleHostingModeAllow
	}
}
