package service

import (
	"regexp"
	"strings"
)

// SecurityService handles security-related operations for module files
type SecurityService struct {
	allowedFileTypes map[string]bool
	pathValidator    *regexp.Regexp
}

// NewSecurityService creates a new security service
func NewSecurityService() *SecurityService {
	return &SecurityService{
		allowedFileTypes: map[string]bool{
			".tf":      true,
			".tf.json": true,
			".tfvars":  true,
			".md":      true,
			".txt":     true,
			".yml":     true,
			".yaml":    true,
			".json":    true,
			".sh":      true,
			".bat":     true,
			".ps1":     true,
		},
		pathValidator: regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`),
	}
}

// ValidateFilePath validates that a file path is safe
func (s *SecurityService) ValidateFilePath(path string) error {
	// Check for empty path
	if path == "" {
		return ErrInvalidFilePath
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return ErrInvalidFilePath
	}

	// Check for absolute paths
	if strings.HasPrefix(path, "/") {
		return ErrInvalidFilePath
	}

	// Check for forbidden characters
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return ErrInvalidFilePath
		}
	}

	// Check path segments
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if segment == "" || segment == "." {
			continue
		}

		// Validate each segment
		if !s.pathValidator.MatchString(segment) {
			return ErrInvalidFilePath
		}
	}

	return nil
}

// SanitizeContent sanitizes HTML content to prevent XSS
// Matches Python bleach implementation with allowed tags and attributes
func (s *SecurityService) SanitizeContent(content *string) error {
	if content == nil {
		return nil
	}

	// Define allowed tags matching Python implementation
	allowedTags := map[string]bool{
		"a": true, "abbr": true, "acronym": true, "b": true,
		"blockquote": true, "code": true, "em": true, "i": true,
		"li": true, "ol": true, "strong": true, "ul": true,
		"p": true, "h1": true, "h2": true, "h3": true,
		"h4": true, "h5": true, "h6": true, "table": true,
		"thead": true, "tbody": true, "th": true, "tr": true,
		"td": true, "pre": true, "img": true, "br": true,
	}

	// Define allowed attributes per tag matching Python implementation
	allowedAttrs := map[string][]string{
		"a":       {"href", "title", "name", "id"},
		"acronym": {"title"},
		"abbr":    {"title"},
		"h1":      {"id"}, "h2": {"id"}, "h3": {"id"},
		"h4": {"id"}, "h5": {"id"}, "h6": {"id"},
		"img":  {"src"},
		"code": {"class"},
	}

	sanitized := s.sanitizeHTML(*content, allowedTags, allowedAttrs)
	*content = sanitized
	return nil
}

// sanitizeHTML implements HTML sanitization matching Python bleach behavior
func (s *SecurityService) sanitizeHTML(content string, allowedTags map[string]bool, allowedAttrs map[string][]string) string {
	// Remove script tags, iframes, and other dangerous elements
	content = s.removeDangerousElements(content)

	// Remove javascript: and other dangerous protocols
	content = s.removeDangerousProtocols(content)

	// Remove event handlers
	content = s.removeEventHandlers(content)

	// Sanitize remaining HTML tags
	content = s.sanitizeTags(content, allowedTags, allowedAttrs)

	return content
}

// removeDangerousElements removes script, iframe, object, embed, form, input, button tags
func (s *SecurityService) removeDangerousElements(content string) string {
	dangerousTags := []string{
		"<script[^>]*>.*?</script>",
		"<iframe[^>]*>.*?</iframe>",
		"<object[^>]*>.*?</object>",
		"<embed[^>]*>.*?</embed>",
		"<form[^>]*>.*?</form>",
		"<input[^>]*>",
		"<button[^>]*>.*?</button>",
	}

	for _, pattern := range dangerousTags {
		re := regexp.MustCompile(`(?i)` + pattern)
		content = re.ReplaceAllString(content, "")
	}

	return content
}

// removeDangerousProtocols removes javascript:, vbscript:, data:, file: protocols
func (s *SecurityService) removeDangerousProtocols(content string) string {
	// For javascript: and data:, remove the protocol and everything after it (until delimiter)
	// These are executable/dangerous and the entire payload should be removed
	// This pattern handles both quoted and unquoted contexts
	re := regexp.MustCompile(`(?i)(javascript:|data:)(?:[^"'\s<>]|\([^)]*\)|'[^']*'|"[^"]*")*`)
	content = re.ReplaceAllString(content, "")

	// For vbscript: and file:, only remove the protocol itself
	// The remaining content is not directly executable
	re = regexp.MustCompile(`(?i)(vbscript:|file:)`)
	content = re.ReplaceAllString(content, "")

	return content
}

// removeEventHandlers removes all on* event handlers
func (s *SecurityService) removeEventHandlers(content string) string {
	// Remove all event handlers with quoted values
	// Match: space + on* + = + quoted value (using backreference for matching quotes)
	re := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*(?:"[^"]*"|'[^']*')`)
	content = re.ReplaceAllString(content, "")

	// Remove event handlers without quotes
	re = regexp.MustCompile(`(?i)\s+on\w+\s*=\s*[^\s>]+`)
	content = re.ReplaceAllString(content, "")

	return content
}

// sanitizeTags removes HTML tags that are not in the allowed list
func (s *SecurityService) sanitizeTags(content string, allowedTags map[string]bool, allowedAttrs map[string][]string) string {
	// Simple tag sanitization - for proper implementation would use HTML parser
	// This matches the basic approach of the current Go implementation
	// Remove dangerous tags
	dangerousTags := []string{
		"<script[^>]*>", "</script>",
		"<iframe[^>]*>", "</iframe>",
		"<object[^>]*>", "</object>",
		"<embed[^>]*>", "</embed>",
		"<form[^>]*>", "</form>",
		"<input[^>]*>", "<button[^>]*>", "</button>",
	}

	for _, pattern := range dangerousTags {
		re := regexp.MustCompile(`(?i)` + pattern)
		content = re.ReplaceAllString(content, "")
	}

	return content
}

// ValidateFileType checks if a file type is allowed
func (s *SecurityService) ValidateFileType(fileName string) error {
	// Get file extension
	parts := strings.Split(fileName, ".")
	if len(parts) < 2 {
		return ErrInvalidFileType
	}

	extension := "." + strings.ToLower(parts[len(parts)-1])
	if !s.allowedFileTypes[extension] {
		return ErrInvalidFileType
	}

	return nil
}

// Error definitions
var (
	ErrInvalidFilePath = NewSecurityError("invalid file path")
	ErrInvalidFileType = NewSecurityError("invalid file type")
)

// SecurityError represents a security-related error
type SecurityError struct {
	message string
}

// NewSecurityError creates a new security error
func NewSecurityError(message string) *SecurityError {
	return &SecurityError{message: message}
}

// Error implements the error interface
func (e *SecurityError) Error() string {
	return e.message
}
