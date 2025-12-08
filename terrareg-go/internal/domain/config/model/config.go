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
