package model

import "fmt"

// ProviderSourceConfig represents a provider source configuration for Terraform provider registries
// Matches the Python PROVIDER_SOURCES configuration structure
type ProviderSourceConfig struct {
	Name                     string `json:"name"`
	Type                     string `json:"type"`
	BaseURL                  string `json:"base_url"`
	ApiURL                   string `json:"api_url"`
	ClientID                 string `json:"client_id"`
	ClientSecret             string `json:"client_secret"`
	PrivateKeyPath           string `json:"private_key_path"`
	DefaultAccessToken       string `json:"default_access_token,omitempty"`
	DefaultInstallationID    string `json:"default_installation_id,omitempty"`
	LoginButtonText          string `json:"login_button_text"`
	AutoGenerateNamespaces   bool   `json:"auto_generate_namespaces"`
}

// ProviderCategory represents a provider category configuration
// Matches the Python PROVIDER_CATEGORIES configuration structure
type ProviderCategory struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	UserSelectable  bool   `json:"user-selectable"`
}

// ProviderSourceConfigCollection represents a collection of provider source configurations
type ProviderSourceConfigCollection []ProviderSourceConfig

// ProviderCategoryCollection represents a collection of provider categories
type ProviderCategoryCollection []ProviderCategory

// IsValid validates the provider source configuration
func (p *ProviderSourceConfig) IsValid() bool {
	return p.Name != "" && p.Type != "" && p.ApiURL != ""
}

// IsValid validates the provider category configuration
func (c *ProviderCategory) IsValid() bool {
	return c.ID > 0 && c.Name != "" && c.Slug != ""
}

// FindByName finds a provider source configuration by name
func (p ProviderSourceConfigCollection) FindByName(name string) *ProviderSourceConfig {
	for _, source := range p {
		if source.Name == name {
			return &source
		}
	}
	return nil
}

// FindByID finds a provider category by ID
func (c ProviderCategoryCollection) FindByID(id int) *ProviderCategory {
	for _, category := range c {
		if category.ID == id {
			return &category
		}
	}
	return nil
}

// Validate validates all provider source configurations in the collection
func (p ProviderSourceConfigCollection) Validate() error {
	for i, source := range p {
		if !source.IsValid() {
			return fmt.Errorf("provider source at index %d is invalid: missing required fields", i)
		}
	}
	return nil
}

// Validate validates all provider categories in the collection
func (c ProviderCategoryCollection) Validate() error {
	seenIDs := make(map[int]bool)
	for i, category := range c {
		if !category.IsValid() {
			return fmt.Errorf("provider category at index %d is invalid: missing required fields", i)
		}
		if seenIDs[category.ID] {
			return fmt.Errorf("duplicate provider category ID found: %d", category.ID)
		}
		seenIDs[category.ID] = true
	}
	return nil
}