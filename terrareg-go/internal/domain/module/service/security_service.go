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
func (s *SecurityService) SanitizeContent(content *string) error {
	if content == nil {
		return nil
	}

	// Basic HTML sanitization - remove dangerous tags and attributes
	dangerousTags := []string{
		"<script", "</script",
		"<iframe", "</iframe",
		"<object", "</object",
		"<embed", "</embed",
		"<form", "</form",
		"<input", "<button",
		"javascript:", "vbscript:",
		"data:", "file:",
		"onload=", "onerror=", "onclick=",
	}

	sanitized := *content
	for _, tag := range dangerousTags {
		sanitized = strings.ReplaceAll(strings.ToLower(sanitized), strings.ToLower(tag), "")
	}

	// Remove script event handlers
	eventHandlers := []string{
		"onload", "onerror", "onclick", "onmouseover", "onmouseout",
		"onfocus", "onblur", "onchange", "onsubmit", "onreset",
	}

	for _, handler := range eventHandlers {
		// Remove any attribute with these event handlers
		re := regexp.MustCompile(`\b` + handler + `\s*=\s*["'][^"']*["']`)
		sanitized = re.ReplaceAllString(sanitized, "")
	}

	*content = sanitized
	return nil
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
