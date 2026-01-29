package namespace

import (
	"context"
	"errors"
	"fmt"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// CreateNamespaceCommand handles creating a new namespace
type CreateNamespaceCommand struct {
	namespaceRepo       repository.NamespaceRepository
	namespaceAuditService *auditservice.NamespaceAuditService
}

// NewCreateNamespaceCommand creates a new create namespace command
func NewCreateNamespaceCommand(namespaceRepo repository.NamespaceRepository, namespaceAuditService *auditservice.NamespaceAuditService) *CreateNamespaceCommand {
	return &CreateNamespaceCommand{
		namespaceRepo:       namespaceRepo,
		namespaceAuditService: namespaceAuditService,
	}
}

// CreateNamespaceRequest represents the request to create a namespace
type CreateNamespaceRequest struct {
	Name        types.NamespaceName
	DisplayName *string
	Type        string // "github_organization", "gitlab_group", or empty
}

// Execute executes the command
func (c *CreateNamespaceCommand) Execute(ctx context.Context, req CreateNamespaceRequest) (*model.Namespace, error) {
	// Check if namespace already exists
	existing, err := c.namespaceRepo.FindByName(ctx, req.Name)
	if err != nil && !errors.Is(err, shared.ErrNotFound) {
		return nil, fmt.Errorf("failed to check namespace existence: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("namespace %s already exists", req.Name)
	}

	// Determine namespace type
	var nsType model.NamespaceType
	if req.Type == "" {
		nsType = model.NamespaceTypeNone
	} else {
		nsType = model.NamespaceType(req.Type)
	}

	// Create namespace domain model
	namespace, err := model.NewNamespace(req.Name, req.DisplayName, nsType)
	if err != nil {
		return nil, fmt.Errorf("failed to create namespace: %w", err)
	}

	// Persist to repository
	if err := c.namespaceRepo.Save(ctx, namespace); err != nil {
		return nil, fmt.Errorf("failed to save namespace: %w", err)
	}

	// Log audit event (synchronous)
	// Python reference: /app/terrareg/models.py:1144 - AuditAction.NAMESPACE_CREATE
	if c.namespaceAuditService != nil {
		c.namespaceAuditService.LogNamespaceCreate(ctx, req.Name)
	}

	return namespace, nil
}
