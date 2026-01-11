package model

import (
	"regexp"
	"strings"
)

// URLTemplate represents a Git URL template with placeholders
type URLTemplate struct {
	Raw     string `json:"raw"`
	IsValid bool   `json:"is_valid"`
}

// GitCredentials represents credentials for Git operations
type GitCredentials struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	SSHKey   string `json:"ssh_key,omitempty"`
}

// Placeholder definitions for URL templates
const (
	PlaceholderNamespace = "{namespace}"
	PlaceholderModule    = "{module}"
	PlaceholderProvider  = "{provider}"
	PlaceholderGitTag    = "{git_tag}"
	PlaceholderVersion   = "{version}"
)

// ValidPlaceholders defines all supported placeholders
var ValidPlaceholders = map[string]bool{
	PlaceholderNamespace: true,
	PlaceholderModule:    true,
	PlaceholderProvider:  true,
	PlaceholderGitTag:    true,
	PlaceholderVersion:   true,
}

// TemplateRegex matches placeholder patterns in templates
var TemplateRegex = regexp.MustCompile(`\{(namespace|module|provider|git_tag|version)\}`)

// NewURLTemplate creates a new URL template from raw string
func NewURLTemplate(raw string) *URLTemplate {
	return &URLTemplate{
		Raw:     raw,
		IsValid: validateTemplate(raw),
	}
}

// validateTemplate checks if a URL template contains valid placeholders
func validateTemplate(template string) bool {
	if template == "" {
		return false
	}

	// Check for malformed braces
	openBraces := strings.Count(template, "{")
	closeBraces := strings.Count(template, "}")
	if openBraces != closeBraces {
		return false
	}

	// Find all placeholder patterns in the template
	allPlaceholders := regexp.MustCompile(`\{[^}]+\}`).FindAllString(template, -1)

	// Check if all placeholders are valid
	for _, placeholder := range allPlaceholders {
		if !ValidPlaceholders[placeholder] {
			return false
		}
	}

	return true
}

// GetPlaceholders returns all placeholders used in the template
func (t *URLTemplate) GetPlaceholders() []string {
	var placeholders []string

	matches := TemplateRegex.FindAllStringSubmatch(t.Raw, -1)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			placeholder := "{" + match[1] + "}"
			if !seen[placeholder] {
				placeholders = append(placeholders, placeholder)
				seen[placeholder] = true
			}
		}
	}

	// Return empty array instead of nil for consistency
	if len(placeholders) == 0 {
		return []string{}
	}

	return placeholders
}

// RequiresCredentials checks if the template requires credentials (HTTPS URL)
func (t *URLTemplate) RequiresCredentials() bool {
	return strings.HasPrefix(strings.ToLower(t.Raw), "https://") ||
		strings.HasPrefix(strings.ToLower(t.Raw), "http://")
}

// IsSSH checks if the template is for SSH URL
func (t *URLTemplate) IsSSH() bool {
	lowerRaw := strings.ToLower(t.Raw)
	return strings.HasPrefix(t.Raw, "git@") ||
		strings.HasPrefix(lowerRaw, "ssh://")
}
