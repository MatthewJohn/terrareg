package service

import (
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
)

// FileProcessingServiceAdapter adapts MarkdownService to implement FileProcessingService interface
type FileProcessingServiceAdapter struct {
	markdownService *MarkdownService
}

// NewFileProcessingServiceAdapter creates a new adapter that wraps MarkdownService
func NewFileProcessingServiceAdapter(markdownService *MarkdownService) model.FileProcessingService {
	return &FileProcessingServiceAdapter{
		markdownService: markdownService,
	}
}

// ProcessMarkdownContent implements FileProcessingService interface
func (a *FileProcessingServiceAdapter) ProcessMarkdownContent(content string) (string, error) {
	if a.markdownService == nil {
		return content, nil
	}

	if content == "" {
		return "", nil
	}

	// Convert markdown to HTML
	html := a.markdownService.ConvertToHTML(content)

	// Sanitize the resulting HTML
	return a.markdownService.SanitizeHTML(html), nil
}

// FormatCodeContent implements FileProcessingService interface
func (a *FileProcessingServiceAdapter) FormatCodeContent(content string, language string) (string, error) {
	// For now, return content as-is. In a full implementation, we could use
	// a syntax highlighter like chroma to format code content
	return content, nil
}

// SanitizeHTML implements FileProcessingService interface
func (a *FileProcessingServiceAdapter) SanitizeHTML(html string) (string, error) {
	if a.markdownService == nil {
		return html, nil
	}

	// Use the existing SanitizeHTML method and add error return for interface compatibility
	return a.markdownService.SanitizeHTML(html), nil
}