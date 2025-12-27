package module

import (
	"context"
	"fmt"

	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	sharedService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/service"
)

// GetReadmeHTMLQuery retrieves README HTML content for a module version
type GetReadmeHTMLQuery struct {
	moduleProviderRepo moduleRepo.ModuleProviderRepository
	markdownService    *sharedService.MarkdownService
}

// NewGetReadmeHTMLQuery creates a new query
func NewGetReadmeHTMLQuery(
	moduleProviderRepo moduleRepo.ModuleProviderRepository,
	markdownService *sharedService.MarkdownService,
) *GetReadmeHTMLQuery {
	return &GetReadmeHTMLQuery{
		moduleProviderRepo: moduleProviderRepo,
		markdownService:    markdownService,
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

	// Convert markdown to HTML using the common markdown service
	html := q.markdownService.ConvertToHTML(
		string(readmeContent),
		sharedService.WithFileName("README.md"),
	)

	return &GetReadmeHTMLResponse{
		HTML: html,
	}, nil
}
