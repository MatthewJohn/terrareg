package terrareg

import "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"

// ConfigResponse represents the API response for the config endpoint
// This matches exactly the Python API response structure
type ConfigResponse struct {
	TRUSTED_NAMESPACE_LABEL              string              `json:"TRUSTED_NAMESPACE_LABEL"`
	CONTRIBUTED_NAMESPACE_LABEL          string              `json:"CONTRIBUTED_NAMESPACE_LABEL"`
	VERIFIED_MODULE_LABEL                string              `json:"VERIFIED_MODULE_LABEL"`
	ANALYTICS_TOKEN_PHRASE               string              `json:"ANALYTICS_TOKEN_PHRASE"`
	ANALYTICS_TOKEN_DESCRIPTION          string              `json:"ANALYTICS_TOKEN_DESCRIPTION"`
	EXAMPLE_ANALYTICS_TOKEN              string              `json:"EXAMPLE_ANALYTICS_TOKEN"`
	DISABLE_ANALYTICS                    bool                `json:"DISABLE_ANALYTICS"`
	ALLOW_MODULE_HOSTING                 string              `json:"ALLOW_MODULE_HOSTING"`
	UPLOAD_API_KEYS_ENABLED              bool                `json:"UPLOAD_API_KEYS_ENABLED"`
	PUBLISH_API_KEYS_ENABLED             bool                `json:"PUBLISH_API_KEYS_ENABLED"`
	DISABLE_TERRAREG_EXCLUSIVE_LABELS    bool                `json:"DISABLE_TERRAREG_EXCLUSIVE_LABELS"`
	ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER bool                `json:"ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER"`
	ALLOW_CUSTOM_GIT_URL_MODULE_VERSION  bool                `json:"ALLOW_CUSTOM_GIT_URL_MODULE_VERSION"`
	SECRET_KEY_SET                       bool                `json:"SECRET_KEY_SET"`
	ADDITIONAL_MODULE_TABS               []string            `json:"ADDITIONAL_MODULE_TABS"`
	OPENID_CONNECT_ENABLED               bool                `json:"OPENID_CONNECT_ENABLED"`
	OPENID_CONNECT_LOGIN_TEXT            string              `json:"OPENID_CONNECT_LOGIN_TEXT"`
	PROVIDER_SOURCES                     []ProviderSourceDTO `json:"PROVIDER_SOURCES"`
	SAML_ENABLED                         bool                `json:"SAML_ENABLED"`
	SAML_LOGIN_TEXT                      string              `json:"SAML_LOGIN_TEXT"`
	ADMIN_LOGIN_ENABLED                  bool                `json:"ADMIN_LOGIN_ENABLED"`
	AUTO_CREATE_NAMESPACE                bool                `json:"AUTO_CREATE_NAMESPACE"`
	AUTO_CREATE_MODULE_PROVIDER          bool                `json:"AUTO_CREATE_MODULE_PROVIDER"`
	DEFAULT_UI_DETAILS_VIEW              string              `json:"DEFAULT_UI_DETAILS_VIEW"`
}

// ProviderSourceDTO represents a provider source in the API response
type ProviderSourceDTO struct {
	Name            string `json:"name"`
	APIName         string `json:"api_name"`
	LoginButtonText string `json:"login_button_text"`
}

// VersionResponse represents the API response for the version endpoint
type VersionResponse struct {
	Version string `json:"version"`
}

// NewConfigResponse creates a new ConfigResponse from a domain model
func NewConfigResponse(config *model.Config) ConfigResponse {
	providerSources := make([]ProviderSourceDTO, len(config.ProviderSources))
	for i, ps := range config.ProviderSources {
		providerSources[i] = ProviderSourceDTO{
			Name:            ps.Name,
			APIName:         ps.APIName,
			LoginButtonText: ps.LoginButtonText,
		}
	}

	return ConfigResponse{
		TRUSTED_NAMESPACE_LABEL:              config.TrustedNamespaceLabel,
		CONTRIBUTED_NAMESPACE_LABEL:          config.ContributedNamespaceLabel,
		VERIFIED_MODULE_LABEL:                config.VerifiedModuleLabel,
		ANALYTICS_TOKEN_PHRASE:               config.AnalyticsTokenPhrase,
		ANALYTICS_TOKEN_DESCRIPTION:          config.AnalyticsTokenDescription,
		EXAMPLE_ANALYTICS_TOKEN:              config.ExampleAnalyticsToken,
		DISABLE_ANALYTICS:                    config.DisableAnalytics,
		ALLOW_MODULE_HOSTING:                 string(config.AllowModuleHosting),
		UPLOAD_API_KEYS_ENABLED:              config.UploadAPIKeysEnabled,
		PUBLISH_API_KEYS_ENABLED:             config.PublishAPIKeysEnabled,
		DISABLE_TERRAREG_EXCLUSIVE_LABELS:    config.DisableTerraregExclusiveLabels,
		ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER: config.AllowCustomGitURLModuleProvider,
		ALLOW_CUSTOM_GIT_URL_MODULE_VERSION:  config.AllowCustomGitURLModuleVersion,
		SECRET_KEY_SET:                       config.SecretKeySet,
		ADDITIONAL_MODULE_TABS:               config.AdditionalModuleTabs,
		OPENID_CONNECT_ENABLED:               config.OpenIDConnectEnabled,
		OPENID_CONNECT_LOGIN_TEXT:            config.OpenIDConnectLoginText,
		PROVIDER_SOURCES:                     providerSources,
		SAML_ENABLED:                         config.SAMLEnabled,
		SAML_LOGIN_TEXT:                      config.SAMLLoginText,
		ADMIN_LOGIN_ENABLED:                  config.AdminLoginEnabled,
		AUTO_CREATE_NAMESPACE:                config.AutoCreateNamespace,
		AUTO_CREATE_MODULE_PROVIDER:          config.AutoCreateModuleProvider,
		DEFAULT_UI_DETAILS_VIEW:              config.DefaultUiDetailsView,
	}
}

// NewVersionResponse creates a new VersionResponse
func NewVersionResponse(version string) VersionResponse {
	return VersionResponse{
		Version: version,
	}
}
