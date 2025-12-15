package service

import (
	"fmt"
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
		AnalyticsTokenPhrase:      s.getEnvStringWithDefault(rawConfig, "ANALYTICS_TOKEN_PHRASE", "analytics token"),
		AnalyticsTokenDescription: s.getEnvStringWithDefault(rawConfig, "ANALYTICS_TOKEN_DESCRIPTION", ""),
		ExampleAnalyticsToken:     s.getEnvStringWithDefault(rawConfig, "EXAMPLE_ANALYTICS_TOKEN", "my-tf-application"),
		DisableAnalytics:          s.parseBool(rawConfig["DISABLE_ANALYTICS"], false),

		// UI configuration
		AutoCreateNamespace:      s.parseBool(rawConfig["AUTO_CREATE_NAMESPACE"], true),
		AutoCreateModuleProvider: s.parseBool(rawConfig["AUTO_CREATE_MODULE_PROVIDER"], true),
		DefaultUiDetailsView:     s.getDefaultUiView(rawConfig["DEFAULT_UI_DETAILS_VIEW"]),
		AdditionalModuleTabs:     s.parseAdditionalModuleTabs(rawConfig["ADDITIONAL_MODULE_TABS"]),
		ModuleLinks:              s.parseStringSlice(rawConfig["MODULE_LINKS"], ","),

		// Terraform example version templates
		TerraformExampleVersionTemplate:          s.getEnvStringWithDefault(rawConfig, "TERRAFORM_EXAMPLE_VERSION_TEMPLATE", "{major}.{minor}.{patch}"),
		TerraformExampleVersionTemplatePreMajor: s.getEnvStringWithDefault(rawConfig, "TERRAFORM_EXAMPLE_VERSION_TEMPLATE_PRE_MAJOR", s.getEnvStringWithDefault(rawConfig, "TERRAFORM_EXAMPLE_VERSION_TEMPLATE", "{major}.{minor}.{patch}")),

		// Provider sources (empty for now, would need more complex parsing)
		ProviderSources: make(map[string]model.ProviderSourceConfig),

		// Module Processing Configuration
		AutoPublishModuleVersions:                 s.parseModuleVersionReindexMode(rawConfig["AUTO_PUBLISH_MODULE_VERSIONS"], model.ModuleVersionReindexModeLegacy),
		ModuleVersionReindexMode:                 s.parseModuleVersionReindexMode(rawConfig["MODULE_VERSION_REINDEX_MODE"], model.ModuleVersionReindexModeLegacy),
		ModuleVersionUseGitCommit:                s.parseBool(rawConfig["MODULE_VERSION_USE_GIT_COMMIT"], false),
		RequiredModuleMetadataAttributes:         s.parseStringSlice(rawConfig["REQUIRED_MODULE_METADATA_ATTRIBUTES"], ","),
		DeleteExternallyHostedArtifacts:          s.parseBool(rawConfig["DELETE_EXTERNALLY_HOSTED_ARTIFACTS"], false),
		AutogenerateModuleProviderDescription:    s.parseBool(rawConfig["AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION"], true),
		AutogenerateUsageBuilderVariables:        s.parseBool(rawConfig["AUTOGENERATE_USAGE_BUILDER_VARIABLES"], true),

		// Terraform Integration Configuration
		Product:                 s.parseProduct(rawConfig["PRODUCT"], model.ProductTerraform),
		DefaultTerraformVersion: s.getEnvStringWithDefault(rawConfig, "DEFAULT_TERRAFORM_VERSION", "1.3.6"),
		TerraformArchiveMirror:  rawConfig["TERRAFORM_ARCHIVE_MIRROR"],
		ManageTerraformRcFile:   s.parseBool(rawConfig["MANAGE_TERRAFORM_RC_FILE"], false),
		ModulesDirectory:        s.getEnvStringWithDefault(rawConfig, "MODULES_DIRECTORY", "modules"),
		ExamplesDirectory:       s.getEnvStringWithDefault(rawConfig, "EXAMPLES_DIRECTORY", "examples"),

		// UI Enhancement Configuration
		OpenIDConnectLoginText: s.getEnvStringWithDefault(rawConfig, "OPENID_CONNECT_LOGIN_TEXT", "Login using OpenID Connect"),
		SAMLLoginText:          s.getEnvStringWithDefault(rawConfig, "SAML2_LOGIN_TEXT", "Login using SAML"),

		// Analytics and Access Configuration
		InternalExtractionAnalyticsToken: s.getEnvStringWithDefault(rawConfig, "INTERNAL_EXTRACTION_ANALYTICS_TOKEN", "internal-terrareg-analytics-token"),
		IgnoreAnalyticsTokenAuthKeys:     s.parseStringSlice(rawConfig["IGNORE_ANALYTICS_TOKEN_AUTH_KEYS"], ","),
		AnalyticsAuthKeys:                s.parseStringSlice(rawConfig["ANALYTICS_AUTH_KEYS"], ","),
		AllowUnidentifiedDownloads:       s.parseBool(rawConfig["ALLOW_UNIDENTIFIED_DOWNLOADS"], false),

		// Redirect Deletion Configuration
		AllowForcefulModuleProviderRedirectDeletion: s.parseBool(rawConfig["ALLOW_FORCEFUL_MODULE_PROVIDER_REDIRECT_DELETION"], false),
		RedirectDeletionLookbackDays:               s.parseInt(rawConfig["REDIRECT_DELETION_LOOKBACK_DAYS"], -1),

		// Example Configuration
		ExampleFileExtensions: s.parseStringSlice(rawConfig["EXAMPLE_FILE_EXTENSIONS"], ","),

		// Provider Registry Configuration
		ProviderSourcesJSON: s.getEnvStringWithDefault(rawConfig, "PROVIDER_SOURCES", "[]"),
		ProviderCategories:  s.getEnvStringWithDefault(rawConfig, "PROVIDER_CATEGORIES", `[{"id": 1, "name": "Example Category", "slug": "example-category", "user-selectable": true}]`),

		// GitHub Integration Configuration
		GithubURL:                                s.getEnvStringWithDefault(rawConfig, "GITHUB_URL", "https://github.com"),
		GithubApiUrl:                             s.getEnvStringWithDefault(rawConfig, "GITHUB_API_URL", "https://api.github.com"),
		GithubAppClientId:                        rawConfig["GITHUB_APP_CLIENT_ID"],
		GithubAppClientSecret:                    rawConfig["GITHUB_APP_CLIENT_SECRET"],
		GithubLoginText:                          s.getEnvStringWithDefault(rawConfig, "GITHUB_LOGIN_TEXT", "Login with Github"),
		AutoGenerateGithubOrganisationNamespaces: s.parseBool(rawConfig["AUTO_GENERATE_GITHUB_ORGANISATION_NAMESPACES"], false),

		// Additional Infracost Configuration
		InfracostTlsInsecureSkipVerify: s.parseBool(rawConfig["INFRACOST_TLS_INSECURE_SKIP_VERIFY"], false),

		// Authentication status (derived from infrastructure)
		OpenIDConnectEnabled: rawConfig["OPENID_CONNECT_CLIENT_ID"] != "" && rawConfig["OPENID_CONNECT_ISSUER"] != "",
		SAMLEnabled:          rawConfig["SAML2_IDP_METADATA_URL"] != "",
		AdminLoginEnabled:    rawConfig["ADMIN_AUTHENTICATION_TOKEN"] != "",
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
		UploadApiKeys:            s.parseStringSlice(rawConfig["UPLOAD_API_KEYS"], ","),
		PublishApiKeys:           s.parseStringSlice(rawConfig["PUBLISH_API_KEYS"], ","),
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
		SentryTracesSampleRate:      s.parseFloat(rawConfig["SENTRY_TRACES_SAMPLE_RATE"], 1.0),

		// Terraform OIDC settings
		TerraformOidcIdpSigningKeyPath:    s.getEnvStringWithDefault(rawConfig, "TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH", "signing_key.pem"),
		TerraformOidcIdpSubjectIdHashSalt: rawConfig["TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT"],
		TerraformOidcIdpSessionExpiry:     s.parseInt(rawConfig["TERRAFORM_OIDC_IDP_SESSION_EXPIRY"], 3600),

		// SSL/TLS Configuration
		SSLCertPrivateKey: rawConfig["SSL_CERT_PRIVATE_KEY"],
		SSLCertPublicKey:   rawConfig["SSL_CERT_PUBLIC_KEY"],

		// Complete SAML Configuration
		SAML2EntityID:       rawConfig["SAML2_ENTITY_ID"],
		SAML2PublicKey:      rawConfig["SAML2_PUBLIC_KEY"],
		SAML2PrivateKey:     rawConfig["SAML2_PRIVATE_KEY"],
		SAML2GroupAttribute: s.getEnvStringWithDefault(rawConfig, "SAML2_GROUP_ATTRIBUTE", "groups"),
		SAML2Debug:          s.parseBool(rawConfig["SAML2_DEBUG"], false),

		// Enhanced OpenID Connect Configuration
		OpenIDConnectScopes: s.parseStringSlice(rawConfig["OPENID_CONNECT_SCOPES"], ","),
		OpenIDConnectDebug:  s.parseBool(rawConfig["OPENID_CONNECT_DEBUG"], false),

		// Access Control Configuration
		AllowUnauthenticatedAccess: s.parseBool(rawConfig["ALLOW_UNAUTHENTICATED_ACCESS"], true),

		// Git Provider Configuration
		GitCloneTimeout:                s.parseInt(rawConfig["GIT_CLONE_TIMEOUT"], 300),
		UpstreamGitCredentialsUsername: rawConfig["UPSTREAM_GIT_CREDENTIALS_USERNAME"],
		UpstreamGitCredentialsPassword: rawConfig["UPSTREAM_GIT_CREDENTIALS_PASSWORD"],

		// Server Configuration
		ServerType:      s.parseServerType(rawConfig["SERVER"], model.ServerTypeBuiltin),
		Threaded:        s.parseBool(rawConfig["THREADED"], true),
		AllowedProviders: s.parseStringSlice(rawConfig["ALLOWED_PROVIDERS"], ","),

		// Terraform Presigned URL Configuration
		TerraformPresignedUrlSecret:        rawConfig["TERRAFORM_PRESIGNED_URL_SECRET"],
		TerraformPresignedUrlExpirySeconds: s.parseInt(rawConfig["TERRAFORM_PRESIGNED_URL_EXPIRY_SECONDS"], 10),
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
// This matches the Python pattern: [attr for attr in os.environ.get(..., '').split(',') if attr]
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

// parseAdditionalModuleTabs handles the special case of ADDITIONAL_MODULE_TABS
// which is a JSON array of arrays in Python, not a simple comma-separated list
func (s *ConfigurationService) parseAdditionalModuleTabs(value string) []string {
	// ADDITIONAL_MODULE_TABS is a special case - it's JSON in Python
	// For now, return it as a single string since parsing the JSON structure
	// would require additional JSON handling and the Go domain model expects []string
	if value == "" {
		return []string{}
	}
	return []string{value}
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

// parseModuleVersionReindexMode parses the MODULE_VERSION_REINDEX_MODE or AUTO_PUBLISH_MODULE_VERSIONS environment variable
func (s *ConfigurationService) parseModuleVersionReindexMode(value string, defaultValue model.ModuleVersionReindexMode) model.ModuleVersionReindexMode {
	if value == "" {
		return defaultValue
	}

	mode := model.ModuleVersionReindexMode(strings.ToLower(value))
	if mode.IsValid() {
		return mode
	}
	return defaultValue
}

// parseProduct parses the PRODUCT environment variable
func (s *ConfigurationService) parseProduct(value string, defaultValue model.Product) model.Product {
	if value == "" {
		return defaultValue
	}

	product := model.Product(strings.ToLower(value))
	if product.IsValid() {
		return product
	}
	return defaultValue
}

// parseServerType parses the SERVER environment variable
func (s *ConfigurationService) parseServerType(value string, defaultValue model.ServerType) model.ServerType {
	if value == "" {
		return defaultValue
	}

	serverType := model.ServerType(strings.ToLower(value))
	if serverType.IsValid() {
		return serverType
	}
	return defaultValue
}