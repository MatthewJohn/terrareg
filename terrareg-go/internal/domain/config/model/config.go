package model

// Config represents the configuration used by the UI
// This matches the Python terrareg.config structure
type Config struct {
	// Namespace labels
	TrustedNamespaceLabel     string
	ContributedNamespaceLabel string
	VerifiedModuleLabel       string

	// Analytics configuration
	AnalyticsTokenPhrase      string
	AnalyticsTokenDescription string
	ExampleAnalyticsToken     string
	DisableAnalytics          bool

	// Feature flags
	AllowModuleHosting              string // Maps to ModuleHostingMode enum value
	UploadAPIKeysEnabled            bool
	PublishAPIKeysEnabled           bool
	DisableTerraregExclusiveLabels  bool
	AllowCustomGitURLModuleProvider bool
	AllowCustomGitURLModuleVersion  bool
	SecretKeySet                    bool

	// Authentication status
	OpenIDConnectEnabled   bool
	OpenIDConnectLoginText string
	SAMLEnabled            bool
	SAMLLoginText          string
	AdminLoginEnabled      bool

	// UI configuration
	AdditionalModuleTabs     []string
	AutoCreateNamespace      bool
	AutoCreateModuleProvider bool
	DefaultUiDetailsView     string // Maps to DefaultUiInputOutputView enum value

	// Provider sources
	ProviderSources []ProviderSource
}

// ProviderSource represents an external provider source configuration
type ProviderSource struct {
	Name            string
	APIName         string
	LoginButtonText string
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
