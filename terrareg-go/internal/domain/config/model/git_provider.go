package model

import "fmt"

// GitProviderConfig represents a single git provider configuration
// Matches the Python GitProviderConfig structure
type GitProviderConfig struct {
	Name      string `json:"name"`
	BaseURL   string `json:"base_url"`
	CloneURL  string `json:"clone_url"`
	BrowseURL string `json:"browse_url"`
	GitPath   string `json:"git_path,omitempty"`
}

// GitProviderConfigCollection represents a collection of git provider configurations
type GitProviderConfigCollection []GitProviderConfig

// IsValid validates the git provider configuration
func (g *GitProviderConfig) IsValid() bool {
	return g.Name != "" && g.BaseURL != "" && g.CloneURL != "" && g.BrowseURL != ""
}

// FindByName finds a git provider configuration by name
func (g GitProviderConfigCollection) FindByName(name string) *GitProviderConfig {
	for _, provider := range g {
		if provider.Name == name {
			return &provider
		}
	}
	return nil
}

// Validate validates all git provider configurations in the collection
func (g GitProviderConfigCollection) Validate() error {
	for i, provider := range g {
		if !provider.IsValid() {
			return fmt.Errorf("git provider at index %d is invalid: missing required fields", i)
		}
	}
	return nil
}
