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
	OpenIDConnectEnabled   bool
	OpenIDConnectLoginText string
	SAMLEnabled            bool
	SAMLLoginText          string
	AdminLoginEnabled      bool

	// UI configuration
	AdditionalModuleTabs     []string
	AutoCreateNamespace      bool
	AutoCreateModuleProvider bool
	DefaultUiDetailsView     DefaultUiInputOutputView

	// Terraform example version templates
	TerraformExampleVersionTemplate          string
	TerraformExampleVersionTemplatePreMajor string

	// Provider sources (domain configuration)
	ProviderSources map[string]ProviderSourceConfig
}

// ProviderSource represents an external provider source configuration (for UI compatibility)
type ProviderSource struct {
	Name            string
	APIName         string
	LoginButtonText string
}

// ProviderSourceConfig holds configuration for external provider sources (domain configuration)
type ProviderSourceConfig struct {
	Type         string
	APIName      string
	ClientID     string
	ClientSecret string
	LoginURL     string
	CallbackURL  string
}

// ModuleHostingMode represents the module hosting mode (from Python enum)
type ModuleHostingMode string

const (
	ModuleHostingModeAllow    ModuleHostingMode = "true"
	ModuleHostingModeDisallow ModuleHostingMode = "false"
	ModuleHostingModeEnforce  ModuleHostingMode = "enforce"
)

// DefaultUiInputOutputView represents the default UI view (from Python enum)
type DefaultUiInputOutputView string

const (
	DefaultUiInputOutputViewTable    DefaultUiInputOutputView = "table"
	DefaultUiInputOutputViewExpanded DefaultUiInputOutputView = "expanded"
)

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
	UploadAPIKeysEnabled            bool             `json:"upload_api_keys_enabled"`
	PublishAPIKeysEnabled           bool             `json:"publish_api_keys_enabled"`
	DisableTerraregExclusiveLabels  bool             `json:"disable_terrareg_exclusive_labels"`
	AllowCustomGitURLModuleProvider bool             `json:"allow_custom_git_url_module_provider"`
	AllowCustomGitURLModuleVersion  bool             `json:"allow_custom_git_url_module_version"`
	SecretKeySet                    bool             `json:"secret_key_set"`

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
