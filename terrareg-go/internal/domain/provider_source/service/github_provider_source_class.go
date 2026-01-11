package service

import (
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
)

// GithubProviderSourceClass implements ProviderSourceClass for GitHub
// Python reference: provider_source/github.py::GithubProviderSource
type GithubProviderSourceClass struct{}

// NewGithubProviderSourceClass creates a new GitHub provider source class
func NewGithubProviderSourceClass() *GithubProviderSourceClass {
	return &GithubProviderSourceClass{}
}

// Type returns the provider source type for GitHub
func (g *GithubProviderSourceClass) Type() model.ProviderSourceType {
	return model.ProviderSourceTypeGithub
}

// GenerateDBConfigFromSourceConfig validates and converts user config to DB config
// Python reference: github.py::generate_db_config_from_source_config()
func (g *GithubProviderSourceClass) GenerateDBConfigFromSourceConfig(sourceConfig map[string]interface{}) (*model.ProviderSourceConfig, error) {
	config := &model.ProviderSourceConfig{}

	// Required string attributes
	requiredStringAttrs := []string{
		"base_url", "api_url", "client_id", "client_secret",
		"login_button_text", "private_key_path", "app_id",
	}

	for _, attr := range requiredStringAttrs {
		val, ok := sourceConfig[attr]
		if !ok {
			return nil, model.NewInvalidProviderSourceConfigError(
				fmt.Sprintf("Missing required Github provider source config: %s", attr),
			)
		}

		strVal, ok := val.(string)
		if !ok || strVal == "" {
			return nil, model.NewInvalidProviderSourceConfigError(
				fmt.Sprintf("Missing required Github provider source config: %s", attr),
			)
		}

		// Set the field based on attribute name
		switch attr {
		case "base_url":
			config.BaseURL = strVal
		case "api_url":
			config.ApiURL = strVal
		case "client_id":
			config.ClientID = strVal
		case "client_secret":
			config.ClientSecret = strVal
		case "login_button_text":
			config.LoginButtonText = strVal
		case "private_key_path":
			config.PrivateKeyPath = strVal
		case "app_id":
			config.AppID = strVal
		}
	}

	// Optional string attributes
	optionalStringAttrs := []string{
		"default_access_token", "default_installation_id",
	}

	for _, attr := range optionalStringAttrs {
		val, ok := sourceConfig[attr]
		if ok {
			if strVal, ok := val.(string); ok {
				switch attr {
				case "default_access_token":
					config.DefaultAccessToken = strVal
				case "default_installation_id":
					config.DefaultInstallationID = strVal
				}
			}
		}
	}

	// Boolean attributes
	// Python reference: github.py line 46 uses "auto_generate_github_organisation_namespaces"
	if val, ok := sourceConfig["auto_generate_github_organisation_namespaces"]; ok {
		if boolVal, ok := val.(bool); ok {
			config.AutoGenerateNamespaces = boolVal
		} else {
			return nil, model.NewInvalidProviderSourceConfigError(
				"auto_generate_github_organisation_namespaces must be a boolean",
			)
		}
	}

	return config, nil
}

// ValidateConfig validates a provider source configuration for GitHub
func (g *GithubProviderSourceClass) ValidateConfig(config *model.ProviderSourceConfig) error {
	return config.ValidateForType(string(model.ProviderSourceTypeGithub))
}
