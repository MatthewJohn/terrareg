package service

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
)

// ConfigurationService is the single source of truth for all configuration loading
type ConfigurationService struct {
	envLoader  *config.EnvironmentLoader
	validator  *config.ConfigValidator
}

// ConfigurationServiceOptions provides configuration for the service
type ConfigurationServiceOptions struct {
	AllowHotReload bool
	ConfigFile     string
}

// NewConfigurationService creates a new configuration service
func NewConfigurationService(opts ConfigurationServiceOptions, versionReader *version.VersionReader) *ConfigurationService {
	return &ConfigurationService{
		envLoader: config.NewEnvironmentLoader(),
		validator: config.NewConfigValidator(),
	}
}

// LoadConfiguration loads all configuration from environment
func (s *ConfigurationService) LoadConfiguration() (*model.DomainConfig, *config.InfrastructureConfig, error) {
	// Load all environment variables once
	rawConfig := s.envLoader.LoadAllEnvironmentVariables()

	// Validate configuration
	if err := s.validator.Validate(rawConfig); err != nil {
		return nil, nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Split into domain and infrastructure configs
	domainConfig := s.buildDomainConfig(rawConfig)
	infrastructureConfig := s.buildInfrastructureConfig(rawConfig)

	return domainConfig, infrastructureConfig, nil
}

// buildDomainConfig creates domain configuration from raw environment variables
func (s *ConfigurationService) buildDomainConfig(rawConfig map[string]string) *model.DomainConfig {
	return &model.DomainConfig{
		// Feature flags
		AllowModuleHosting:              s.parseModuleHostingMode(rawConfig["ALLOW_MODULE_HOSTING"]),
		UploadAPIKeysEnabled:            rawConfig["UPLOAD_API_KEYS"] != "",
		PublishAPIKeysEnabled:           rawConfig["PUBLISH_API_KEYS"] != "",
		DisableTerraregExclusiveLabels:  s.parseBool(rawConfig["DISABLE_TERRAREG_EXCLUSIVE_LABELS"], false),
		AllowCustomGitURLModuleProvider: s.parseBool(rawConfig["ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER"], true),
		AllowCustomGitURLModuleVersion:  s.parseBool(rawConfig["ALLOW_CUSTOM_GIT_URL_MODULE_VERSION"], true),
		SecretKeySet:                    rawConfig["SECRET_KEY"] != "",

		// Namespace settings
		TrustedNamespaces:        s.parseStringSlice(rawConfig["TRUSTED_NAMESPACES"], ","),
		VerifiedModuleNamespaces: s.parseStringSlice(rawConfig["VERIFIED_MODULE_NAMESPACES"], ","),

		// UI configuration
		TrustedNamespaceLabel:     s.getEnvStringWithDefault(rawConfig, "TRUSTED_NAMESPACE_LABEL", "Trusted"),
		ContributedNamespaceLabel: s.getEnvStringWithDefault(rawConfig, "CONTRIBUTED_NAMESPACE_LABEL", "Contributed"),
		VerifiedModuleLabel:       s.getEnvStringWithDefault(rawConfig, "VERIFIED_MODULE_LABEL", "Verified"),

		// Analytics configuration
		AnalyticsTokenPhrase:      rawConfig["ANALYTICS_TOKEN_PHRASE"],
		AnalyticsTokenDescription: rawConfig["ANALYTICS_TOKEN_DESCRIPTION"],
		ExampleAnalyticsToken:     rawConfig["EXAMPLE_ANALYTICS_TOKEN"],
		DisableAnalytics:          s.parseBool(rawConfig["DISABLE_ANALYTICS"], false),

		// Additional UI settings
		AdditionalModuleTabs: s.parseStringSlice(rawConfig["ADDITIONAL_MODULE_TABS"], ","),
		DefaultUiDetailsView: s.getDefaultUiView(rawConfig["DEFAULT_UI_DETAILS_VIEW"]),
		AutoCreateNamespace:  s.parseBool(rawConfig["AUTO_CREATE_NAMESPACE"], true),
		AutoCreateModuleProvider: s.parseBool(rawConfig["AUTO_CREATE_MODULE_PROVIDER"], true),

		// Authentication status (derived from infrastructure)
		OpenIDConnectEnabled:   rawConfig["OPENID_CONNECT_CLIENT_ID"] != "" && rawConfig["OPENID_CONNECT_ISSUER"] != "",
		OpenIDConnectLoginText: s.getEnvStringWithDefault(rawConfig, "OPENID_CONNECT_LOGIN_TEXT", "Login with OpenID"),
		SAMLEnabled:            rawConfig["SAML2_IDP_METADATA_URL"] != "",
		SAMLLoginText:          s.getEnvStringWithDefault(rawConfig, "SAML2_LOGIN_TEXT", "Login with SAML"),
		AdminLoginEnabled:      rawConfig["ADMIN_AUTHENTICATION_TOKEN"] != "",

		// Provider sources (empty for now, would need more complex parsing)
		ProviderSources: make(map[string]model.ProviderSourceConfig),
	}
}

// buildInfrastructureConfig creates infrastructure configuration from raw environment variables
func (s *ConfigurationService) buildInfrastructureConfig(rawConfig map[string]string) *config.InfrastructureConfig {
	return &config.InfrastructureConfig{
		// Server settings
		ListenPort: s.parseInt(rawConfig["LISTEN_PORT"], 5000),
		PublicURL:  rawConfig["PUBLIC_URL"],
		DomainName: rawConfig["DOMAIN_NAME"],
		Debug:      s.parseBool(rawConfig["DEBUG"], false),

		// Database settings
		DatabaseURL: s.getEnvStringWithDefault(rawConfig, "DATABASE_URL", "sqlite:///modules.db"),

		// Storage settings
		DataDirectory:   s.getEnvStringWithDefault(rawConfig, "DATA_DIRECTORY", "./data"),
		UploadDirectory: s.getEnvStringWithDefault(rawConfig, "UPLOAD_DIRECTORY", "./data/upload"),

		// Git provider settings
		GitProviderConfig: rawConfig["GIT_PROVIDER_CONFIG"],

		// Authentication settings (infrastructure)
		// SAML
		SAML2IDPMetadataURL: rawConfig["SAML2_IDP_METADATA_URL"],
		SAML2IssuerEntityID: rawConfig["SAML2_ISSUER_ENTITY_ID"],

		// OpenID Connect
		OpenIDConnectClientID:     rawConfig["OPENID_CONNECT_CLIENT_ID"],
		OpenIDConnectClientSecret: rawConfig["OPENID_CONNECT_CLIENT_SECRET"],
		OpenIDConnectIssuer:       rawConfig["OPENID_CONNECT_ISSUER"],

		// Admin authentication
		AdminAuthenticationToken: rawConfig["ADMIN_AUTHENTICATION_TOKEN"],
		SecretKey:                rawConfig["SECRET_KEY"],

		// Feature flags (infrastructure)
		AllowProviderHosting:   s.parseBool(rawConfig["ALLOW_PROVIDER_HOSTING"], true),
		AllowCustomGitProvider: s.parseBool(rawConfig["ALLOW_CUSTOM_GIT_PROVIDER"], true),
		EnableAccessControls:   s.parseBool(rawConfig["ENABLE_ACCESS_CONTROLS"], false),
		EnableSecurityScanning: s.parseBool(rawConfig["ENABLE_SECURITY_SCANNING"], true),

		// UI Customization (infrastructure assets)
		ApplicationName: s.getEnvStringWithDefault(rawConfig, "APPLICATION_NAME", "Terrareg"),
		LogoURL:         s.getEnvStringWithDefault(rawConfig, "LOGO_URL", "/static/images/logo.png"),
		SiteWarning:     rawConfig["SITE_WARNING"],

		// Session settings
		SessionExpiry:          s.parseDuration(rawConfig["SESSION_EXPIRY_MINS"], 60) * time.Minute,
		AdminSessionExpiryMins: s.parseInt(rawConfig["ADMIN_SESSION_EXPIRY_MINS"], 60),
		SessionCookieName:      s.getEnvStringWithDefault(rawConfig, "SESSION_COOKIE_NAME", "terrareg_session"),
		SessionRefreshAge:      s.parseDuration(rawConfig["SESSION_REFRESH_MINS"], 25) * time.Minute,

		// External service settings
		InfracostAPIKey:             rawConfig["INFRACOST_API_KEY"],
		InfracostPricingAPIEndpoint: rawConfig["INFRACOST_PRICING_API_ENDPOINT"],
		SentryDSN:                   rawConfig["SENTRY_DSN"],
		SentryTracesSampleRate:      s.parseFloat(rawConfig["SENTRY_TRACES_SAMPLE_RATE"], 0.1),
	}
}

// Helper methods using the best logic from existing implementations

// parseModuleHostingMode parses the ALLOW_MODULE_HOSTING environment variable
// and returns the corresponding ModuleHostingMode enum value
func (s *ConfigurationService) parseModuleHostingMode(value string) model.ModuleHostingMode {
	if value == "" {
		value = "true" // Default to allow
	}

	value = strings.ToLower(value)

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

// getEnvStringWithDefault gets a string value from config map with default
func (s *ConfigurationService) getEnvStringWithDefault(rawConfig map[string]string, key, defaultValue string) string {
	if value, exists := rawConfig[key]; exists && value != "" {
		return value
	}
	return defaultValue
}

// parseBool parses a boolean value from string with default
func (s *ConfigurationService) parseBool(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}

	lowerValue := strings.ToLower(value)
	return lowerValue == "true" || lowerValue == "1" || lowerValue == "yes"
}

// parseInt parses an integer value from string with default
func (s *ConfigurationService) parseInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

// parseFloat parses a float64 value from string with default
func (s *ConfigurationService) parseFloat(value string, defaultValue float64) float64 {
	if value == "" {
		return defaultValue
	}

	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}
	return defaultValue
}

// parseDuration parses minutes as time.Duration with default
func (s *ConfigurationService) parseDuration(value string, defaultMinutes int) time.Duration {
	minutes := s.parseInt(value, defaultMinutes)
	return time.Duration(minutes) * time.Minute
}

// parseStringSlice parses a comma-separated string into a slice
func (s *ConfigurationService) parseStringSlice(value, separator string) []string {
	if value == "" {
		return []string{}
	}

	items := strings.Split(value, separator)
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

// getDefaultUiView parses the default UI details view
func (s *ConfigurationService) getDefaultUiView(value string) model.DefaultUiInputOutputView {
	switch s.getEnvStringWithDefault(map[string]string{"DEFAULT_UI_DETAILS_VIEW": value}, "DEFAULT_UI_DETAILS_VIEW", "table") {
	case "expanded":
		return model.DefaultUiInputOutputViewExpanded
	default:
		return model.DefaultUiInputOutputViewTable
	}
}