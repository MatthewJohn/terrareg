package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// ModuleFileService handles module file operations following DDD principles
type ModuleFileService struct {
	moduleProviderRepo    moduleRepo.ModuleProviderRepository
	moduleVersionFileRepo model.ModuleVersionFileRepository
	namespaceService      *NamespaceService
	securityService       *SecurityService
}

// NewModuleFileService creates a new module file service
func NewModuleFileService(
	moduleProviderRepo moduleRepo.ModuleProviderRepository,
	moduleVersionFileRepo model.ModuleVersionFileRepository,
	namespaceService *NamespaceService,
	securityService *SecurityService,
) *ModuleFileService {
	return &ModuleFileService{
		moduleProviderRepo:    moduleProviderRepo,
		moduleVersionFileRepo: moduleVersionFileRepo,
		namespaceService:      namespaceService,
		securityService:       securityService,
	}
}

// GetModuleFileRequest represents a request to get a module file
type GetModuleFileRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   string
	Path      string
}

// GetModuleFileResponse represents the response for getting a module file
type GetModuleFileResponse struct {
	File        *model.ModuleVersionFile
	Content     string
	ContentType string
	ContentHTML string // For display purposes (processed markdown, etc.)
}

// GetModuleFile retrieves a module file with proper security and processing
func (s *ModuleFileService) GetModuleFile(ctx context.Context, req *GetModuleFileRequest) (*GetModuleFileResponse, error) {
	// Validate path to prevent directory traversal
	if err := s.validateFilePath(req.Path); err != nil {
		return nil, err
	}

	// Get module provider first
	moduleProvider, err := s.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx,
		req.Namespace,
		req.Module,
		req.Provider,
	)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, fmt.Errorf("module version not found: %w", model.ErrFileNotFound)
		}
		return nil, fmt.Errorf("failed to get module provider: %w", err)
	}

	// Get module version from module provider
	moduleVersion, err := moduleProvider.GetVersion(req.Version)
	if err != nil || moduleVersion == nil {
		return nil, fmt.Errorf("version %s not found: %w", req.Version, model.ErrFileNotFound)
	}

	// Check if module is accessible
	if err := s.checkModuleAccess(ctx, moduleProvider); err != nil {
		return nil, err
	}

	// Get module version file from repository
	file, err := s.moduleVersionFileRepo.FindByPath(ctx, moduleVersion.ID(), req.Path)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, fmt.Errorf("file not found: %w", model.ErrFileNotFound)
		}
		return nil, fmt.Errorf("failed to get module file: %w", err)
	}

	// Validate file path again with the specific file
	if err := file.ValidatePath(); err != nil {
		return nil, err
	}

	// Process content based on file type
	contentHTML := file.Content()
	if file.IsMarkdown() {
		// Convert markdown to HTML and sanitize
		contentHTML = s.processMarkdownContent(file.Content())
	} else if file.ContentType() == "text/plain" {
		// Wrap plain text in pre tags
		contentHTML = fmt.Sprintf("<pre>%s</pre>", file.Content())
	}

	// Sanitize content for security
	if err := s.securityService.SanitizeContent(&contentHTML); err != nil {
		return nil, fmt.Errorf("failed to sanitize content: %w", err)
	}

	return &GetModuleFileResponse{
		File:        file,
		Content:     file.Content(),
		ContentType: file.ContentType(),
		ContentHTML: contentHTML,
	}, nil
}

// validateFilePath validates that a file path is safe
func (s *ModuleFileService) validateFilePath(path string) error {
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return model.ErrInvalidFilePath
	}

	// Ensure path is relative and doesn't start with /
	if strings.HasPrefix(path, "/") {
		return model.ErrInvalidFilePath
	}

	// Check for empty path
	if path == "" {
		return model.ErrInvalidFilePath
	}

	// Additional path validation using security service
	if err := s.securityService.ValidateFilePath(path); err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	return nil
}

// checkModuleAccess checks if the user has access to the module
func (s *ModuleFileService) checkModuleAccess(ctx context.Context, moduleProvider *model.ModuleProvider) error {
	// Get namespace from module provider
	namespace := moduleProvider.Namespace()

	// Check if namespace is trusted
	if !s.namespaceService.IsTrusted(namespace) {
		return model.ErrUnauthorized
	}

	// Additional security checks can be added here
	// For example, checking if the module is published, user permissions, etc.

	return nil
}

// processMarkdownContent processes markdown content into HTML
func (s *ModuleFileService) processMarkdownContent(content string) string {
	// For now, return basic HTML. In a full implementation,
	// this would use a markdown library like blackfriday
	// and HTML sanitization
	return fmt.Sprintf(
		`<div class="markdown-content">
			<pre>%s</pre>
		</div>`,
		content,
	)
}

// ListModuleFiles lists all files for a module version
func (s *ModuleFileService) ListModuleFiles(ctx context.Context, namespace, moduleName, provider, version string) ([]*model.ModuleVersionFile, error) {
	// Get module provider first
	moduleProvider, err := s.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx,
		namespace,
		moduleName,
		provider,
	)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, fmt.Errorf("module version not found: %w", shared.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get module provider: %w", err)
	}

	// Get module version from module provider
	moduleVersion, err := moduleProvider.GetVersion(version)
	if err != nil || moduleVersion == nil {
		return nil, fmt.Errorf("version %s not found: %w", version, shared.ErrNotFound)
	}

	// Check module access
	if err := s.checkModuleAccess(ctx, moduleProvider); err != nil {
		return nil, err
	}

	// Get all files for the module version
	files, err := s.moduleVersionFileRepo.FindByModuleVersionID(ctx, moduleVersion.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to list module files: %w", err)
	}

	return files, nil
}
