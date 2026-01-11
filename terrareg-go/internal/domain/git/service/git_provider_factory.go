package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/repository"
)

// GitProviderFactory manages git provider configuration and instantiation
// Python reference: models.py::GitProvider
type GitProviderFactory struct {
	repo repository.GitProviderRepository
}

// NewGitProviderFactory creates a new GitProvider factory
func NewGitProviderFactory(repo repository.GitProviderRepository) *GitProviderFactory {
	return &GitProviderFactory{
		repo: repo,
	}
}

// GetByName retrieves a git provider by its name
// Python reference: models.py::GitProvider.get_by_name()
func (f *GitProviderFactory) GetByName(ctx context.Context, name string) (*model.GitProvider, error) {
	return f.repo.FindByName(ctx, name)
}

// GetAll retrieves all git providers
// Python reference: models.py::GitProvider.get_all()
func (f *GitProviderFactory) GetAll(ctx context.Context) ([]*model.GitProvider, error) {
	return f.repo.FindAll(ctx)
}

// InitialiseFromConfig loads git providers from config JSON into database
// Python reference: models.py::GitProvider.initialise_from_config()
func (f *GitProviderFactory) InitialiseFromConfig(ctx context.Context, configJSON string) error {
	// Parse JSON config
	var gitProviderConfigs []map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &gitProviderConfigs); err != nil {
		return &model.InvalidGitProviderConfigError{
			Message: "Git provider config is not valid JSON",
		}
	}

	for _, gitProviderConfig := range gitProviderConfigs {
		// Validate required attributes
		requiredAttrs := []string{"name", "base_url", "clone_url", "browse_url"}
		for _, attr := range requiredAttrs {
			if _, ok := gitProviderConfig[attr]; !ok {
				return &model.InvalidGitProviderConfigError{
					Message: fmt.Sprintf("Git provider config does not contain required attribute: %s", attr),
				}
			}
		}

		// Extract config values
		name, _ := gitProviderConfig["name"].(string)
		baseURL, _ := gitProviderConfig["base_url"].(string)
		cloneURL, _ := gitProviderConfig["clone_url"].(string)
		browseURL, _ := gitProviderConfig["browse_url"].(string)

		// Obtain git path value, defaulting to empty string
		// Python reference: models.py line 498
		gitPathTemplate := ""
		if val, ok := gitProviderConfig["git_path"]; ok && val != nil {
			if strVal, ok := val.(string); ok {
				gitPathTemplate = strVal
			}
		}

		// Validate URL templates
		// Python reference: models.py lines 500-520
		// Append git_path to URLs for validation as placeholders may be delegated to git_path
		if err := f.validateURLTemplates(baseURL+gitPathTemplate, cloneURL+gitPathTemplate, browseURL+gitPathTemplate); err != nil {
			return err
		}

		// If git_path template is an empty string, it should be stored as empty string
		// (Go strings are already empty by default, unlike Python's None)

		// Create or update git provider
		provider := &model.GitProvider{
			Name:              name,
			BaseURLTemplate:   baseURL,
			CloneURLTemplate:  cloneURL,
			BrowseURLTemplate: browseURL,
			GitPathTemplate:   gitPathTemplate,
		}

		if err := f.repo.Upsert(ctx, provider); err != nil {
			return fmt.Errorf("failed to upsert git provider %s: %w", name, err)
		}
	}

	return nil
}

// validateURLTemplates validates the URL templates using GitUrlTemplateValidator
// Python reference: models.py lines 503-520
func (f *GitProviderFactory) validateURLTemplates(baseURL, cloneURL, browseURL string) error {
	// Validate base_url + git_path: requires namespace and module placeholders
	// Python reference: models.py lines 503-508
	baseValidator := model.NewGitUrlTemplateValidator(baseURL)
	if err := baseValidator.Validate(model.ValidateRequest{
		RequiresNamespacePlaceholder: true,
		RequiresModulePlaceholder:    true,
		RequiresTagPlaceholder:       false,
		RequiresPathPlaceholder:      false,
	}); err != nil {
		return err
	}

	// Validate clone_url + git_path: requires namespace and module placeholders
	// Python reference: models.py lines 509-514
	cloneValidator := model.NewGitUrlTemplateValidator(cloneURL)
	if err := cloneValidator.Validate(model.ValidateRequest{
		RequiresNamespacePlaceholder: true,
		RequiresModulePlaceholder:    true,
		RequiresTagPlaceholder:       false,
		RequiresPathPlaceholder:      false,
	}); err != nil {
		return err
	}

	// Validate browse_url + git_path: requires namespace, module, tag, and path placeholders
	// Python reference: models.py lines 515-520
	browseValidator := model.NewGitUrlTemplateValidator(browseURL)
	if err := browseValidator.Validate(model.ValidateRequest{
		RequiresNamespacePlaceholder: true,
		RequiresModulePlaceholder:    true,
		RequiresTagPlaceholder:       true,
		RequiresPathPlaceholder:      true,
	}); err != nil {
		return err
	}

	return nil
}
