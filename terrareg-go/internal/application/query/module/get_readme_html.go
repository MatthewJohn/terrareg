package module

import (
	"context"
	"fmt"
	"strings"

	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetReadmeHTMLQuery retrieves README HTML content for a module version
type GetReadmeHTMLQuery struct {
	moduleProviderRepo moduleRepo.ModuleProviderRepository
}

// NewGetReadmeHTMLQuery creates a new query
func NewGetReadmeHTMLQuery(moduleProviderRepo moduleRepo.ModuleProviderRepository) *GetReadmeHTMLQuery {
	return &GetReadmeHTMLQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// GetReadmeHTMLRequest represents a request to get README HTML
type GetReadmeHTMLRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   string
}

// GetReadmeHTMLResponse represents the response for getting README HTML
type GetReadmeHTMLResponse struct {
	HTML string `json:"html"`
}

// Execute retrieves README HTML for a module version
func (q *GetReadmeHTMLQuery) Execute(ctx context.Context, req *GetReadmeHTMLRequest) (*GetReadmeHTMLResponse, error) {
	// Get module provider
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx,
		req.Namespace,
		req.Module,
		req.Provider,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find module provider: %w", err)
	}

	// Get module version
	moduleVersion, err := moduleProvider.GetVersion(req.Version)
	if err != nil || moduleVersion == nil {
		return nil, fmt.Errorf("version %s not found: %w", req.Version, err)
	}

	// Get README content from module details
	readmeContent := moduleVersion.Details().ReadmeContent()
	if len(readmeContent) == 0 {
		return nil, fmt.Errorf("no README content found")
	}

	// Convert markdown to HTML
	html := q.convertMarkdownToHTML(string(readmeContent))

	return &GetReadmeHTMLResponse{
		HTML: html,
	}, nil
}

// convertMarkdownToHTML converts markdown content to HTML
// For now, this is a basic implementation - in a full system, would use a proper markdown library
func (q *GetReadmeHTMLQuery) convertMarkdownToHTML(markdown string) string {
	if markdown == "" {
		return ""
	}

	// Basic markdown to HTML conversion
	// In a full implementation, this would use a library like blackfriday
	html := markdown

	// Convert headers
	html = strings.ReplaceAll(html, "# ", "<h1>")
	html = strings.ReplaceAll(html, "## ", "<h2>")
	html = strings.ReplaceAll(html, "### ", "<h3>")

	// Convert line breaks to <br>
	html = strings.ReplaceAll(html, "\n", "<br>")

	// Wrap in a div for styling
	return fmt.Sprintf(`<div class="markdown-content">%s</div>`, html)
}