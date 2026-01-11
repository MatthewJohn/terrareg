package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ProviderSourceType defines the type of provider source
type ProviderSourceType string

const (
	ProviderSourceTypeGithub    ProviderSourceType = "github"
	ProviderSourceTypeGitlab    ProviderSourceType = "gitlab"
	ProviderSourceTypeBitbucket ProviderSourceType = "bitbucket"
)

// String returns the string representation of the provider source type
func (t ProviderSourceType) String() string {
	return string(t)
}

// IsValid checks if the provider source type is valid
func (t ProviderSourceType) IsValid() bool {
	switch t {
	case ProviderSourceTypeGithub, ProviderSourceTypeGitlab, ProviderSourceTypeBitbucket:
		return true
	default:
		return false
	}
}

// ProviderSourceConfig represents the configuration stored in database
type ProviderSourceConfig struct {
	BaseURL                string `json:"base_url"`
	ApiURL                 string `json:"api_url"`
	ClientID               string `json:"client_id"`
	ClientSecret           string `json:"client_secret"`
	PrivateKeyPath         string `json:"private_key_path"`
	AppID                  string `json:"app_id"`
	DefaultAccessToken     string `json:"default_access_token,omitempty"`
	DefaultInstallationID  string `json:"default_installation_id,omitempty"`
	LoginButtonText        string `json:"login_button_text"`
	AutoGenerateNamespaces bool   `json:"auto_generate_namespaces"`
}

// ValidateForType validates the configuration for a specific provider type
func (c *ProviderSourceConfig) ValidateForType(typeStr string) error {
	type_ := ProviderSourceType(strings.ToLower(typeStr))

	if !type_.IsValid() {
		return NewInvalidProviderSourceConfigError(fmt.Sprintf("Invalid provider source type. Valid types: %s", getValidProviderTypes()))
	}

	if type_ == ProviderSourceTypeGithub {
		return c.validateGitHubConfig()
	}

	// Add validation for other provider types when implemented
	return nil
}

// validateGitHubConfig validates GitHub-specific configuration
func (c *ProviderSourceConfig) validateGitHubConfig() error {
	requiredFields := map[string]string{
		"base_url":          c.BaseURL,
		"api_url":           c.ApiURL,
		"client_id":         c.ClientID,
		"client_secret":     c.ClientSecret,
		"login_button_text": c.LoginButtonText,
		"private_key_path":   c.PrivateKeyPath,
		"app_id":            c.AppID,
	}

	for fieldName, value := range requiredFields {
		if value == "" {
			return NewInvalidProviderSourceConfigError(fmt.Sprintf("Missing required Github provider source config: %s", fieldName))
		}
	}

	return nil
}

// ProviderSource represents a provider source entity (Domain Entity)
type ProviderSource struct {
	name    string
	apiName string
	type_   ProviderSourceType
	config  *ProviderSourceConfig
}

// NewProviderSource creates a new provider source entity
func NewProviderSource(name, apiName string, type_ ProviderSourceType, config *ProviderSourceConfig) *ProviderSource {
	return &ProviderSource{
		name:    name,
		apiName: apiName,
		type_:   type_,
		config:  config,
	}
}

// Name returns the provider source name
func (ps *ProviderSource) Name() string {
	return ps.name
}

// ApiName returns the API-friendly name
func (ps *ProviderSource) ApiName() string {
	return ps.apiName
}

// Type returns the provider source type
func (ps *ProviderSource) Type() ProviderSourceType {
	return ps.type_
}

// Config returns the provider source configuration
func (ps *ProviderSource) Config() *ProviderSourceConfig {
	return ps.config
}

// SetConfig updates the configuration (used for upsert operations)
func (ps *ProviderSource) SetConfig(config *ProviderSourceConfig) {
	ps.config = config
}

// Validate ensures the provider source is valid
func (ps *ProviderSource) Validate() error {
	if ps.name == "" {
		return NewInvalidProviderSourceConfigError("Provider source name cannot be empty")
	}
	if ps.config == nil {
		return NewInvalidProviderSourceConfigError("Provider source configuration cannot be nil")
	}
	return ps.config.ValidateForType(ps.type_.String())
}

// ToDBModel converts the domain model to the DB model
// This is used by the repository for persistence
func (ps *ProviderSource) ToDBModel() *sqldb.ProviderSourceDB {
	configJSON := sqldb.EncodeBlob(ps.config)

	var apiNamePtr *string
	if ps.apiName != "" {
		apiNamePtr = &ps.apiName
	}

	return &sqldb.ProviderSourceDB{
		Name:               ps.name,
		APIName:            apiNamePtr,
		ProviderSourceType: sqldb.ProviderSourceType(ps.type_),
		Config:             configJSON,
	}
}

// GetValidProviderTypes returns a comma-separated list of valid provider types
// Exported for use by factory
func GetValidProviderTypes() string {
	return strings.Join([]string{
		ProviderSourceTypeGithub.String(),
		// Add GitLab and Bitbucket when implemented
		// ProviderSourceTypeGitlab.String(),
		// ProviderSourceTypeBitbucket.String(),
	}, ", ")
}

// getValidProviderTypes returns a comma-separated list of valid provider types
func getValidProviderTypes() string {
	return GetValidProviderTypes()
}

// nameToApiName converts a display name to an API-friendly name
// Matches Python behavior exactly
// Python reference: provider_source/factory.py::_name_to_api_name
func nameToApiName(name string) string {
	if name == "" {
		return ""
	}

	// Convert to lower case
	name = strings.ToLower(name)

	// Replace spaces with dashes
	name = strings.ReplaceAll(name, " ", "-")

	// Remove any non alphanumeric/dashes
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	name = reg.ReplaceAllString(name, "")

	// Remove any leading/trailing dashes
	name = strings.Trim(name, "-")

	return name
}

// InvalidProviderSourceConfigError represents an error in provider source configuration
type InvalidProviderSourceConfigError struct {
	Message string
}

func (e *InvalidProviderSourceConfigError) Error() string {
	return e.Message
}

// NewInvalidProviderSourceConfigError creates a new config error
func NewInvalidProviderSourceConfigError(msg string) *InvalidProviderSourceConfigError {
	return &InvalidProviderSourceConfigError{Message: msg}
}

// ProviderSourceNotFoundError represents a provider source not found error
type ProviderSourceNotFoundError struct {
	Name string
}

func (e *ProviderSourceNotFoundError) Error() string {
	return fmt.Sprintf("provider source not found: %s", e.Name)
}

// NewProviderSourceNotFoundError creates a new not found error
func NewProviderSourceNotFoundError(name string) *ProviderSourceNotFoundError {
	return &ProviderSourceNotFoundError{Name: name}
}
