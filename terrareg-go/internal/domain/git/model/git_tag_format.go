package model

import (
	"fmt"
	"regexp"
	"strings"
)

// GitTagFormat represents a template for converting module versions to git tags.
// This is distinct from URL templates - it's specifically for version/tag conversion.
// Python reference: /app/terrareg/models.py - ModuleProvider.git_tag_format property
//
// Examples:
//   - "v{version}" with version "1.2.3" → git tag "v1.2.3"
//   - "releases/v{major}.{minor}" with version "1.2.3" → git tag "releases/v1.2"
//   - "{major}.{patch}" with version "1.2.3" → git tag "1.3" (minor defaults to 0)
type GitTagFormat struct {
	Raw     string
	IsValid bool
}

// Git tag format placeholder constants
const (
	GitTagPlaceholderVersion = "{version}"
	GitTagPlaceholderMajor   = "{major}"
	GitTagPlaceholderMinor   = "{minor}"
	GitTagPlaceholderPatch   = "{patch}"
)

// ValidGitTagFormatPlaceholders defines all valid placeholders for git tag format
var ValidGitTagFormatPlaceholders = map[string]bool{
	GitTagPlaceholderVersion: true,
	GitTagPlaceholderMajor:   true,
	GitTagPlaceholderMinor:   true,
	GitTagPlaceholderPatch:   true,
}

// GitTagFormatRegex matches git tag format placeholders
var GitTagFormatRegex = regexp.MustCompile(`\{(version|major|minor|patch)\}`)

// NewGitTagFormat creates a new git tag format from raw string
// Returns a GitTagFormat with IsValid set based on validation
func NewGitTagFormat(raw string) *GitTagFormat {
	return &GitTagFormat{
		Raw:     raw,
		IsValid: validateGitTagFormat(raw),
	}
}

// Validate validates the git tag format and returns an error if invalid
// This is the preferred way to validate - it returns a proper error message
// Python reference: /app/terrareg/models.py:2694-2710 (update_git_tag_format)
func (f *GitTagFormat) Validate() error {
	// Empty format is valid (uses default)
	if f.Raw == "" {
		return nil
	}

	// Check for malformed braces
	openBraces := strings.Count(f.Raw, "{")
	closeBraces := strings.Count(f.Raw, "}")
	if openBraces != closeBraces {
		return &InvalidGitTagFormatError{
			Message: "Invalid git tag format. Must contain one placeholder: {version}, {major}, {minor}, {patch}.",
		}
	}

	// Find all placeholder patterns in the format
	allPlaceholders := regexp.MustCompile(`\{[^}]+\}`).FindAllString(f.Raw, -1)

	// Check if all placeholders are valid git tag format placeholders
	for _, placeholder := range allPlaceholders {
		if !ValidGitTagFormatPlaceholders[placeholder] {
			return &InvalidGitTagFormatError{
				Message: "Invalid git tag format. Must contain one placeholder: {version}, {major}, {minor}, {patch}.",
			}
		}
	}

	// Ensure at least one valid placeholder is present
	if len(allPlaceholders) == 0 {
		return &InvalidGitTagFormatError{
			Message: "Invalid git tag format. Must contain one placeholder: {version}, {major}, {minor}, {patch}.",
		}
	}

	return nil
}

// validateGitTagFormat validates a git tag format string
// Python reference: /app/terrareg/models.py:2694-2710 (update_git_tag_format)
func validateGitTagFormat(format string) bool {
	if format == "" {
		// Empty format is valid (will use default "{version}")
		return true
	}

	// Check for malformed braces (unmatched { or })
	openBraces := strings.Count(format, "{")
	closeBraces := strings.Count(format, "}")
	if openBraces != closeBraces {
		return false
	}

	// Find all placeholder patterns in the format
	allPlaceholders := regexp.MustCompile(`\{[^}]+\}`).FindAllString(format, -1)

	// Check if all placeholders are valid git tag format placeholders
	for _, placeholder := range allPlaceholders {
		if !ValidGitTagFormatPlaceholders[placeholder] {
			return false
		}
	}

	// Ensure at least one valid placeholder is present
	// (except empty string, which will use default)
	if len(allPlaceholders) == 0 {
		return false
	}

	return true
}

// FormatVersion converts a version to a git tag using this format
// Returns the git tag string, or error if the format is invalid
// Python reference: /app/terrareg/models.py:2370 (get_tag_regex usage)
func (f *GitTagFormat) FormatVersion(version string) (string, error) {
	if !f.IsValid {
		return "", fmt.Errorf("invalid git tag format: %s", f.Raw)
	}

	// Use default format if empty
	format := f.Raw
	if format == "" {
		format = "{version}"
	}

	// Substitute version placeholder
	result := strings.ReplaceAll(format, GitTagPlaceholderVersion, version)

	// For formats without {version}, we'd need to parse the version
	// and substitute major/minor/patch. This is handled by FormatVersionWithComponents.
	if !strings.Contains(format, GitTagPlaceholderVersion) {
		// This version requires version components - use FormatVersionWithComponents instead
		return "", fmt.Errorf("git tag format %s requires version components (major/minor/patch), use FormatVersionWithComponents", f.Raw)
	}

	// Check for any remaining placeholders (shouldn't happen if IsValid is true)
	remaining := GitTagFormatRegex.FindAllString(result, -1)
	if len(remaining) > 0 {
		return "", fmt.Errorf("git tag format contains placeholders that couldn't be substituted: %v", remaining)
	}

	return result, nil
}

// FormatVersionWithComponents converts version components to a git tag
// This is used for formats that use {major}, {minor}, or {patch} placeholders
// Python reference: /app/terrareg/models.py - ModuleProvider.get_git_tag() logic
func (f *GitTagFormat) FormatVersionWithComponents(major, minor, patch string) (string, error) {
	if !f.IsValid {
		return "", fmt.Errorf("invalid git tag format: %s", f.Raw)
	}

	// Use default format if empty
	format := f.Raw
	if format == "" {
		format = "{version}"
	}

	// If format uses {version}, construct version from components
	if strings.Contains(format, GitTagPlaceholderVersion) {
		version := major + "." + minor + "." + patch
		return strings.ReplaceAll(format, GitTagPlaceholderVersion, version), nil
	}

	// Substitute individual components
	result := format
	result = strings.ReplaceAll(result, GitTagPlaceholderMajor, major)
	result = strings.ReplaceAll(result, GitTagPlaceholderMinor, minor)
	result = strings.ReplaceAll(result, GitTagPlaceholderPatch, patch)

	// Check for any remaining placeholders
	remaining := GitTagFormatRegex.FindAllString(result, -1)
	if len(remaining) > 0 {
		return "", fmt.Errorf("git tag format contains placeholders that couldn't be substituted: %v", remaining)
	}

	return result, nil
}

// GetPlaceholders returns all placeholders used in this git tag format
func (f *GitTagFormat) GetPlaceholders() []string {
	var placeholders []string

	matches := GitTagFormatRegex.FindAllStringSubmatch(f.Raw, -1)
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

	if len(placeholders) == 0 {
		return []string{}
	}

	return placeholders
}

// RequiresVersionComponents returns true if this format requires individual version components
// (major, minor, patch) rather than just the full version string
func (f *GitTagFormat) RequiresVersionComponents() bool {
	if f.Raw == "" {
		return false // Default format uses {version}
	}
	return !strings.Contains(f.Raw, GitTagPlaceholderVersion)
}

// InvalidGitTagFormatError is returned when a git tag format is invalid
// Python reference: /app/terrareg/errors.py::InvalidGitTagFormatError
type InvalidGitTagFormatError struct {
	Message string
}

func (e *InvalidGitTagFormatError) Error() string {
	return e.Message
}
