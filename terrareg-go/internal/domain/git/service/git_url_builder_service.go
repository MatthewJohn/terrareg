package service

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// GitURLBuilderService handles sophisticated URL template building for Git operations
// Matches Python's URL template system with credential injection
type GitURLBuilderService struct {
	// No dependencies needed for pure URL building
}

// NewGitURLBuilderService creates a new Git URL builder service
func NewGitURLBuilderService() *GitURLBuilderService {
	return &GitURLBuilderService{}
}

// URLBuilderRequest contains parameters for URL building
type URLBuilderRequest struct {
	Template    string                `json:"template"`
	Namespace   string                `json:"namespace"`
	Module      string                `json:"module"`
	Provider    string                `json:"provider"`
	GitTag      *string               `json:"git_tag,omitempty"`
	Version     *shared.Version       `json:"version,omitempty"`
	Credentials *model.GitCredentials `json:"credentials,omitempty"`
}

// BuildCloneURL builds a clone URL from template with variable substitution
// Matches Python's get_git_clone_url functionality
func (s *GitURLBuilderService) BuildCloneURL(req *URLBuilderRequest) (string, error) {
	// Create and validate template
	template := model.NewURLTemplate(req.Template)
	if !template.IsValid {
		return "", fmt.Errorf("invalid URL template: %s", req.Template)
	}

	// Substitute placeholders
	renderedURL := s.substitutePlaceholders(template.Raw, req)

	// Inject credentials for HTTPS URLs
	if req.Credentials != nil && template.RequiresCredentials() {
		renderedURL = s.injectCredentials(renderedURL, req.Credentials)
	}

	return renderedURL, nil
}

// BuildBrowseURL builds a browse URL from template
func (s *GitURLBuilderService) BuildBrowseURL(template string, namespace, module, provider, path string) (*string, error) {
	urlTemplate := model.NewURLTemplate(template)
	if !urlTemplate.IsValid {
		return nil, fmt.Errorf("invalid URL template: %s", template)
	}

	req := &URLBuilderRequest{
		Template:  template,
		Namespace: namespace,
		Module:    module,
		Provider:  provider,
	}

	renderedURL := s.substitutePlaceholders(urlTemplate.Raw, req)

	// Append path if provided
	if path != "" {
		if !strings.HasPrefix(path, "/") {
			renderedURL += "/"
		}
		renderedURL += path
	}

	return &renderedURL, nil
}

// BuildArchiveURL builds an archive URL for module downloads
func (s *GitURLBuilderService) BuildArchiveURL(template string, namespace, module, provider string, version *shared.Version) (*string, error) {
	urlTemplate := model.NewURLTemplate(template)
	if !urlTemplate.IsValid {
		return nil, fmt.Errorf("invalid URL template: %s", template)
	}

	req := &URLBuilderRequest{
		Template:  template,
		Namespace: namespace,
		Module:    module,
		Provider:  provider,
		Version:   version,
	}

	renderedURL := s.substitutePlaceholders(urlTemplate.Raw, req)
	return &renderedURL, nil
}

// ValidateTemplate validates a URL template for correct placeholders
func (s *GitURLBuilderService) ValidateTemplate(template string) error {
	urlTemplate := model.NewURLTemplate(template)
	if !urlTemplate.IsValid {
		return fmt.Errorf("invalid URL template: unsupported placeholders or format")
	}
	return nil
}

// substitutePlaceholders replaces all placeholders with actual values
func (s *GitURLBuilderService) substitutePlaceholders(template string, req *URLBuilderRequest) string {
	result := template

	// Replace namespace
	result = strings.ReplaceAll(result, model.PlaceholderNamespace, req.Namespace)

	// Replace module
	result = strings.ReplaceAll(result, model.PlaceholderModule, req.Module)

	// Replace provider
	result = strings.ReplaceAll(result, model.PlaceholderProvider, req.Provider)

	// Replace git_tag if provided
	if req.GitTag != nil {
		result = strings.ReplaceAll(result, model.PlaceholderGitTag, *req.GitTag)
	}

	// Replace version if provided
	if req.Version != nil {
		result = strings.ReplaceAll(result, model.PlaceholderVersion, req.Version.String())
	}

	return result
}

// injectCredentials injects credentials into HTTPS URLs
// Matches Python's credential injection logic in _clone_repository
func (s *GitURLBuilderService) injectCredentials(urlStr string, credentials *model.GitCredentials) string {
	if credentials.Username == "" && credentials.Password == "" {
		return urlStr // No credentials to inject
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr // Return original if parsing fails
	}

	// Only inject credentials for HTTP/HTTPS URLs
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return urlStr
	}

	// Extract domain without existing credentials
	domain := parsedURL.Host
	if strings.Contains(domain, "@") {
		// Remove existing credentials if present
		parts := strings.Split(domain, "@")
		domain = parts[len(parts)-1]
	}

	// Build userinfo with credentials
	if credentials.Username != "" {
		if credentials.Password != "" {
			parsedURL.User = url.UserPassword(credentials.Username, credentials.Password)
		} else {
			parsedURL.User = url.User(credentials.Username)
		}
	}
	return parsedURL.String()
}

// ParseTemplateVariables extracts variables from a template string
// Useful for validation and debugging
func (s *GitURLBuilderService) ParseTemplateVariables(template string) ([]string, error) {
	urlTemplate := model.NewURLTemplate(template)
	if !urlTemplate.IsValid {
		return nil, fmt.Errorf("invalid template")
	}

	return urlTemplate.GetPlaceholders(), nil
}

// IsSSHTemplate checks if a template is for SSH URLs
func (s *GitURLBuilderService) IsSSHTemplate(template string) bool {
	urlTemplate := model.NewURLTemplate(template)
	return urlTemplate.IsSSH()
}

// RequiresCredentials checks if a template requires credential injection
func (s *GitURLBuilderService) RequiresCredentials(template string) bool {
	urlTemplate := model.NewURLTemplate(template)
	return urlTemplate.RequiresCredentials()
}
