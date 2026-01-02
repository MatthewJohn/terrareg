package model

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// GitUrlTemplateValidator validates URL templates for Git providers
// Python reference: validators.py::GitUrlValidator
type GitUrlTemplateValidator struct {
	template string
}

// Valid placeholders for GitProvider URL templates (matching Python's GitProvider)
// These are different from the git_url_builder placeholders
const (
	GitProviderPlaceholderNamespace     = "{namespace}"
	GitProviderPlaceholderModule        = "{module}"
	GitProviderPlaceholderProvider      = "{provider}"
	GitProviderPlaceholderPath          = "{path}"
	GitProviderPlaceholderTag           = "{tag}"
	GitProviderPlaceholderTagURIEncoded = "{tag_uri_encoded}"
)

// NewGitUrlTemplateValidator creates a new Git URL template validator
func NewGitUrlTemplateValidator(template string) *GitUrlTemplateValidator {
	return &GitUrlTemplateValidator{
		template: template,
	}
}

// ValidateRequest defines validation requirements
type ValidateRequest struct {
	RequiresNamespacePlaceholder bool
	RequiresModulePlaceholder    bool
	RequiresTagPlaceholder       bool
	RequiresPathPlaceholder      bool
}

// Validate validates the URL template against the given requirements
// Python reference: validators.py::GitUrlValidator::validate()
func (v *GitUrlTemplateValidator) Validate(req ValidateRequest) error {
	// Ensure template contains only valid placeholders
	if err := v.validatePlaceholders(); err != nil {
		return err
	}

	// Check for required namespace placeholder
	if req.RequiresNamespacePlaceholder {
		if !strings.Contains(v.template, GitProviderPlaceholderNamespace) {
			return &RepositoryUrlParseError{
				Message: "Namespace placeholder not present in URL",
			}
		}
		// Verify placeholder actually works (not malformed)
		if !v.placeholderWorks(GitProviderPlaceholderNamespace) {
			return &RepositoryUrlParseError{
				Message: "Template does not contain valid namespace placeholder",
			}
		}
	}

	// Check for required module placeholder
	if req.RequiresModulePlaceholder {
		if !strings.Contains(v.template, GitProviderPlaceholderModule) {
			return &RepositoryUrlParseError{
				Message: "Module placeholder not present in URL",
			}
		}
		// Verify placeholder actually works
		if !v.placeholderWorks(GitProviderPlaceholderModule) {
			return &RepositoryUrlParseError{
				Message: "Template does not contain valid module placeholder",
			}
		}
	}

	// Check for required tag placeholder (either {tag} or {tag_uri_encoded})
	if req.RequiresTagPlaceholder {
		if !strings.Contains(v.template, GitProviderPlaceholderTag) && !strings.Contains(v.template, GitProviderPlaceholderTagURIEncoded) {
			return &RepositoryUrlParseError{
				Message: "tag or tag_uri_encoded placeholder not present in URL",
			}
		}
		// Verify placeholder works
		if !v.placeholderWorks(GitProviderPlaceholderTag) && !v.placeholderWorks(GitProviderPlaceholderTagURIEncoded) {
			return &RepositoryUrlParseError{
				Message: "Template does not contain valid tag/tag_uri_encoded placeholder",
			}
		}
	}

	// Check for required path placeholder
	if req.RequiresPathPlaceholder {
		if !strings.Contains(v.template, GitProviderPlaceholderPath) {
			return &RepositoryUrlParseError{
				Message: "Path placeholder not present in URL",
			}
		}
		// Verify placeholder works
		if !v.placeholderWorks(GitProviderPlaceholderPath) {
			return &RepositoryUrlParseError{
				Message: "Template does not contain valid path placeholder",
			}
		}
	}

	return nil
}

// validatePlaceholders ensures template only contains valid placeholders
// Python reference: validators.py lines 19-33
func (v *GitUrlTemplateValidator) validatePlaceholders() error {
	// Try to substitute all placeholders with empty strings
	// This will catch invalid/unknown placeholders
	_, err := v.GetValue("", "", "", "", "")
	if err != nil {
		return err
	}
	return nil
}

// placeholderWorks checks if a placeholder actually works when substituted
// Python reference: validators.py lines 36-46, 48-58, etc.
// Uses a "really random string" to verify the placeholder is functional
func (v *GitUrlTemplateValidator) placeholderWorks(placeholder string) bool {
	const randomString = "D3f1N1t3LyW0nt3x15t!"

	// Build substitution map
	substitutions := map[string]string{
		GitProviderPlaceholderNamespace:     "",
		GitProviderPlaceholderModule:        "",
		GitProviderPlaceholderProvider:      "",
		GitProviderPlaceholderPath:          "",
		GitProviderPlaceholderTag:           "",
		GitProviderPlaceholderTagURIEncoded: "",
	}

	// Substitute the random string for the placeholder we're testing
	substitutions[placeholder] = randomString

	// Perform substitution and check if random string appears in result
	result := v.template
	for key, value := range substitutions {
		result = strings.ReplaceAll(result, key, value)
	}

	return strings.Contains(result, randomString)
}

// GetValue returns the template with placeholders replaced
// Python reference: validators.py::GitUrlValidator::get_value()
func (v *GitUrlTemplateValidator) GetValue(namespace, module, provider, tag, path string) (string, error) {
	result := v.template

	// URL encode the tag for tag_uri_encoded
	tagURIEncoded := url.QueryEscape(tag)

	// Substitute all placeholders
	result = strings.ReplaceAll(result, GitProviderPlaceholderNamespace, namespace)
	result = strings.ReplaceAll(result, GitProviderPlaceholderModule, module)
	result = strings.ReplaceAll(result, GitProviderPlaceholderProvider, provider)
	result = strings.ReplaceAll(result, GitProviderPlaceholderTag, tag)
	result = strings.ReplaceAll(result, GitProviderPlaceholderTagURIEncoded, tagURIEncoded)
	result = strings.ReplaceAll(result, GitProviderPlaceholderPath, path)

	// Check if any unreplaced placeholders remain
	remainingPattern := regexp.MustCompile(`\{[^}]+\}`)
	remaining := remainingPattern.FindAllString(result, -1)
	if len(remaining) > 0 {
		return "", &RepositoryUrlContainsInvalidTemplateError{
			Message: fmt.Sprintf("Template contains unknown placeholder(s): %s. Valid placeholders are: {namespace}, {module}, {provider}, {path}, {tag} and {tag_uri_encoded}",
				strings.Join(remaining, ", ")),
		}
	}

	return result, nil
}

// RepositoryUrlParseError is returned when URL validation fails
// Python reference: errors.py::RepositoryUrlParseError
type RepositoryUrlParseError struct {
	Message string
}

func (e *RepositoryUrlParseError) Error() string {
	return e.Message
}

// RepositoryUrlContainsInvalidTemplateError is returned when template contains invalid placeholders
// Python reference: errors.py::RepositoryUrlContainsInvalidTemplateError
type RepositoryUrlContainsInvalidTemplateError struct {
	Message string
}

func (e *RepositoryUrlContainsInvalidTemplateError) Error() string {
	return e.Message
}

// InvalidGitProviderConfigError is returned when git provider config is invalid
// Python reference: errors.py::InvalidGitProviderConfigError
type InvalidGitProviderConfigError struct {
	Message string
}

func (e *InvalidGitProviderConfigError) Error() string {
	return e.Message
}
