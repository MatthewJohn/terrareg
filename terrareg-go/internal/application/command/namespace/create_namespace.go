package namespace

import (
	"context"
	"errors"
	"fmt"

	"github.com/terrareg/terrareg/internal/domain/module/model"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
	"github.com/terrareg/terrareg/internal/domain/shared"
)

// CreateNamespaceCommand handles creating a new namespace
type CreateNamespaceCommand struct {
	namespaceRepo repository.NamespaceRepository
}

// NewCreateNamespaceCommand creates a new create namespace command
func NewCreateNamespaceCommand(namespaceRepo repository.NamespaceRepository) *CreateNamespaceCommand {
	return &CreateNamespaceCommand{
		namespaceRepo: namespaceRepo,
	}
}

// CreateNamespaceRequest represents the request to create a namespace
type CreateNamespaceRequest struct {
	Name        string
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

	return namespace, nil
}
