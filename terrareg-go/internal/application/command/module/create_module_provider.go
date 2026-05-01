package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	gitModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

// CreateModuleProviderCommand handles creating a new module provider
type CreateModuleProviderCommand struct {
	namespaceRepo      repository.NamespaceRepository
	moduleProviderRepo repository.ModuleProviderRepository
	auditService       service.ModuleAuditServiceInterface
}

// NewCreateModuleProviderCommand creates a new create module provider command
func NewCreateModuleProviderCommand(
	namespaceRepo repository.NamespaceRepository,
	moduleProviderRepo repository.ModuleProviderRepository,
	auditService service.ModuleAuditServiceInterface,
) *CreateModuleProviderCommand {
	return &CreateModuleProviderCommand{
		namespaceRepo:      namespaceRepo,
		moduleProviderRepo: moduleProviderRepo,
		auditService:       auditService,
	}
}

// CreateModuleProviderRequest represents the request to create a module provider
// Python reference: /app/server/api/terrareg_module_provider_create.py
type CreateModuleProviderRequest struct {
	// Path parameters
	Namespace types.NamespaceName
	Module    types.ModuleName
	Provider  types.ModuleProviderName

	// JSON body fields (optional)
	GitProviderID         *int    `json:"git_provider_id"`
	RepoBaseURLTemplate   *string `json:"repo_base_url_template"`
	RepoCloneURLTemplate  *string `json:"repo_clone_url_template"`
	RepoBrowseURLTemplate *string `json:"repo_browse_url_template"`
	GitTagFormat          *string `json:"git_tag_format"`
	GitPath               *string `json:"git_path"`
	ArchiveGitPath        *bool   `json:"archive_git_path"`
}

// Execute executes the command
func (c *CreateModuleProviderCommand) Execute(ctx context.Context, req CreateModuleProviderRequest) (*model.ModuleProvider, error) {
	// Validate git_tag_format if provided
	// Python reference: /app/server/api/terrareg_module_provider_create.py:161-164
	if req.GitTagFormat != nil {
		gitTagFormat := gitModel.NewGitTagFormat(*req.GitTagFormat)
		if err := gitTagFormat.Validate(); err != nil {
			return nil, err
		}
	}
	// Find the namespace
	namespace, err := c.namespaceRepo.FindByName(ctx, req.Namespace)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			// Python reference: /app/server/api/terrareg_module_provider_create.py:94
			return nil, fmt.Errorf("%w: Namespace does not exist", shared.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to find namespace: %w", err)
	}

	// Check if module provider already exists
	existing, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, req.Namespace, req.Module, req.Provider)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, fmt.Errorf("failed to check module provider existence: %w", err)
	}
	if existing != nil {
		// Python reference: /app/server/api/terrareg_module_provider_create.py:100
		return nil, fmt.Errorf("%w: Module provider already exists", shared.ErrAlreadyExists)
	}

	// Create module provider domain model
	moduleProvider, err := model.NewModuleProvider(namespace, req.Module, req.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create module provider: %w", err)
	}

	// Set Git configuration if provided
	// Python reference: /app/server/api/terrareg_module_provider_create.py:114-164
	if req.GitProviderID != nil || req.RepoBaseURLTemplate != nil || req.RepoCloneURLTemplate != nil ||
		req.RepoBrowseURLTemplate != nil || req.GitTagFormat != nil || req.GitPath != nil || req.ArchiveGitPath != nil {

		archiveGitPath := false
		if req.ArchiveGitPath != nil {
			archiveGitPath = *req.ArchiveGitPath
		}

		moduleProvider.SetGitConfiguration(
			req.GitProviderID,
			req.RepoBaseURLTemplate,
			req.RepoCloneURLTemplate,
			req.RepoBrowseURLTemplate,
			req.GitTagFormat,
			req.GitPath,
			archiveGitPath,
		)
	}

	// Persist to repository
	if err := c.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return nil, fmt.Errorf("failed to save module provider: %w", err)
	}

	// Log audit event (synchronous)
	username := "system"
	// Try to get username from auth context if available
	if authCtx := middleware.GetAuthContext(ctx); authCtx.IsAuthenticated() {
		username = authCtx.GetUsername()
	}

	_ = c.auditService.LogModuleProviderCreate(
		ctx,
		types.NamespaceName(username),
		req.Namespace,
		req.Module,
		req.Provider,
	)

	return moduleProvider, nil
}
