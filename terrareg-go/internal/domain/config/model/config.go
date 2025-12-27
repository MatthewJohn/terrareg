package model

// DomainConfig represents the domain-specific configuration used by the application
// This contains only business logic and UI-relevant configuration, no infrastructure concerns
type DomainConfig struct {
	// Namespace settings
	TrustedNamespaces        []string
	VerifiedModuleNamespaces []string

	// Namespace labels
	TrustedNamespaceLabel     string
	ContributedNamespaceLabel string
	VerifiedModuleLabel       string

	// Analytics configuration
	AnalyticsTokenPhrase      string
	AnalyticsTokenDescription string
	ExampleAnalyticsToken     string
	DisableAnalytics          bool

	// Feature flags (domain concerns)
	AllowModuleHosting              ModuleHostingMode // Maps to ModuleHostingMode enum value
	UploadAPIKeysEnabled            bool
	PublishAPIKeysEnabled           bool
	DisableTerraregExclusiveLabels  bool
	AllowCustomGitURLModuleProvider bool
	AllowCustomGitURLModuleVersion  bool
	SecretKeySet                    bool

	// Authentication status (domain-level only - no secrets)
	OpenIDConnectEnabled bool
	SAMLEnabled          bool
	AdminLoginEnabled    bool

	// UI configuration
	AutoCreateNamespace      bool
	AutoCreateModuleProvider bool
	DefaultUiDetailsView     DefaultUiInputOutputView `env:"DEFAULT_UI_DETAILS_VIEW"`
	AdditionalModuleTabs     []string                 `env:"ADDITIONAL_MODULE_TABS"`
	ModuleLinks              []string                 `env:"MODULE_LINKS"`

	// Terraform example version templates
	TerraformExampleVersionTemplate         string `env:"TERRAFORM_EXAMPLE_VERSION_TEMPLATE"`
	TerraformExampleVersionTemplatePreMajor string `env:"TERRAFORM_EXAMPLE_VERSION_TEMPLATE_PRE_MAJOR"`

	// Provider sources (domain configuration)
	ProviderSources map[string]ProviderSourceConfig

	// Module Processing Configuration
	AutoPublishModuleVersions             bool                     `env:"AUTO_PUBLISH_MODULE_VERSIONS"`
	ModuleVersionReindexMode              ModuleVersionReindexMode `env:"MODULE_VERSION_REINDEX_MODE"`
	ModuleVersionUseGitCommit             bool                     `env:"MODULE_VERSION_USE_GIT_COMMIT"`
	RequiredModuleMetadataAttributes      []string                 `env:"REQUIRED_MODULE_METADATA_ATTRIBUTES"`
	DeleteExternallyHostedArtifacts       bool                     `env:"DELETE_EXTERNALLY_HOSTED_ARTIFACTS"`
	AutogenerateModuleProviderDescription bool                     `env:"AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION"`
	AutogenerateUsageBuilderVariables     bool                     `env:"AUTOGENERATE_USAGE_BUILDER_VARIABLES"`

	// Terraform Integration Configuration
	Product                 Product `env:"PRODUCT"`
	DefaultTerraformVersion string  `env:"DEFAULT_TERRAFORM_VERSION"`
	TerraformArchiveMirror  string  `env:"TERRAFORM_ARCHIVE_MIRROR"`
	ManageTerraformRcFile   bool    `env:"MANAGE_TERRAFORM_RC_FILE"`
	ModulesDirectory        string  `env:"MODULES_DIRECTORY"`
	ExamplesDirectory       string  `env:"EXAMPLES_DIRECTORY"`

	// UI Enhancement Configuration
	OpenIDConnectLoginText string `env:"OPENID_CONNECT_LOGIN_TEXT"`
	SAMLLoginText          string `env:"SAML2_LOGIN_TEXT"`

	// Analytics and Access Configuration
	InternalExtractionAnalyticsToken string   `env:"INTERNAL_EXTRACTION_ANALYTICS_TOKEN"`
	IgnoreAnalyticsTokenAuthKeys     []string `env:"IGNORE_ANALYTICS_TOKEN_AUTH_KEYS"`
	AnalyticsAuthKeys                []string `env:"ANALYTICS_AUTH_KEYS"`
	AllowUnidentifiedDownloads       bool     `env:"ALLOW_UNIDENTIFIED_DOWNLOADS"`

	// Redirect Deletion Configuration
	AllowForcefulModuleProviderRedirectDeletion bool `env:"ALLOW_FORCEFUL_MODULE_PROVIDER_REDIRECT_DELETION"`
	RedirectDeletionLookbackDays                int  `env:"REDIRECT_DELETION_LOOKBACK_DAYS"`

	// Example Configuration
	ExampleFileExtensions []string `env:"EXAMPLE_FILE_EXTENSIONS"`

	// Provider Registry Configuration
	ProviderSourcesJSON string `env:"PROVIDER_SOURCES"`
	ProviderCategories  string `env:"PROVIDER_CATEGORIES"`

	// GitHub Integration Configuration
	GithubURL                                string `env:"GITHUB_URL"`
	GithubApiUrl                             string `env:"GITHUB_API_URL"`
	GithubAppClientId                        string `env:"GITHUB_APP_CLIENT_ID"`
	GithubAppClientSecret                    string `env:"GITHUB_APP_CLIENT_SECRET"`
	GithubLoginText                          string `env:"GITHUB_LOGIN_TEXT"`
	AutoGenerateGithubOrganisationNamespaces bool   `env:"AUTO_GENERATE_GITHUB_ORGANISATION_NAMESPACES"`

	// Additional Infracost Configuration
	InfracostTlsInsecureSkipVerify bool `env:"INFRACOST_TLS_INSECURE_SKIP_VERIFY"`
}

// ProviderSource represents an external provider source configuration (for UI compatibility)
type ProviderSource struct {
	Name            string
	APIName         string
	LoginButtonText string
}

// UIConfig represents the configuration that is safe to expose to the UI
// This is a read-only view of domain configuration optimized for presentation
type UIConfig struct {
	// Namespace labels
	TrustedNamespaceLabel     string `json:"trusted_namespace_label"`
	ContributedNamespaceLabel string `json:"contributed_namespace_label"`
	VerifiedModuleLabel       string `json:"verified_module_label"`

	// Analytics configuration
	AnalyticsTokenPhrase      string `json:"analytics_token_phrase"`
	AnalyticsTokenDescription string `json:"analytics_token_description"`
	ExampleAnalyticsToken     string `json:"example_analytics_token"`
	DisableAnalytics          bool   `json:"disable_analytics"`

	// Feature flags
	AllowModuleHosting              ModuleHostingMode `json:"allow_module_hosting"`
	UploadAPIKeysEnabled            bool              `json:"upload_api_keys_enabled"`
	PublishAPIKeysEnabled           bool              `json:"publish_api_keys_enabled"`
	DisableTerraregExclusiveLabels  bool              `json:"disable_terrareg_exclusive_labels"`
	AllowCustomGitURLModuleProvider bool              `json:"allow_custom_git_url_module_provider"`
	AllowCustomGitURLModuleVersion  bool              `json:"allow_custom_git_url_module_version"`
	SecretKeySet                    bool              `json:"secret_key_set"`

	// Authentication status
	OpenIDConnectEnabled   bool   `json:"openid_connect_enabled"`
	OpenIDConnectLoginText string `json:"openid_connect_login_text"`
	SAMLEnabled            bool   `json:"saml_enabled"`
	SAMLLoginText          string `json:"saml_login_text"`
	AdminLoginEnabled      bool   `json:"admin_login_enabled"`

	// UI configuration
	AdditionalModuleTabs     []string `json:"additional_module_tabs"`
	AutoCreateNamespace      bool     `json:"auto_create_namespace"`
	AutoCreateModuleProvider bool     `json:"auto_create_module_provider"`
	DefaultUiDetailsView     string   `json:"default_ui_details_view"`

	// Provider sources
	ProviderSources []ProviderSource `json:"provider_sources"`
}
