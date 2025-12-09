package module

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
)

// GetModuleVersionFileQuery handles retrieving module version files
type GetModuleVersionFileQuery struct {
	moduleFileService *service.ModuleFileService
}

// NewGetModuleVersionFileQuery creates a new get module version file query
func NewGetModuleVersionFileQuery(
	moduleFileService *service.ModuleFileService,
) *GetModuleVersionFileQuery {
	return &GetModuleVersionFileQuery{
		moduleFileService: moduleFileService,
	}
}

// GetModuleVersionFileRequest represents a request to get a module version file
type GetModuleVersionFileRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   string
	Path      string
}

// GetModuleVersionFileResponse represents the response for getting a module version file
type GetModuleVersionFileResponse struct {
	File         *ModuleFileResponse
	Content      string
	ContentType  string
	ContentHTML  string // For display purposes (processed markdown, etc.)
}

// Execute executes the query to get a module version file
func (q *GetModuleVersionFileQuery) Execute(ctx context.Context, req *GetModuleVersionFileRequest) (*GetModuleVersionFileResponse, error) {
	// Convert to service request
	serviceReq := &service.GetModuleFileRequest{
		Namespace: req.Namespace,
		Module:    req.Module,
		Provider:  req.Provider,
		Version:   req.Version,
		Path:      req.Path,
	}

	// Execute the service operation
	serviceResp, err := q.moduleFileService.GetModuleFile(ctx, serviceReq)
	if err != nil {
		return nil, err
	}

	// Convert to query response
	return &GetModuleVersionFileResponse{
		File: &ModuleFileResponse{
			ID:          serviceResp.File.ID(),
			Path:        serviceResp.File.Path(),
			FileName:    serviceResp.File.FileName(),
			ContentType: serviceResp.File.ContentType(),
			IsMarkdown:  serviceResp.File.IsMarkdown(),
		},
		Content:     serviceResp.Content,
		ContentType: serviceResp.ContentType,
		ContentHTML: serviceResp.ContentHTML,
	}, nil
}

// ModuleFileResponse represents a module file response (simplified for DTO)
type ModuleFileResponse struct {
	ID          int    `json:"id"`
	Path        string `json:"path"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	IsMarkdown  bool   `json:"is_markdown"`
}