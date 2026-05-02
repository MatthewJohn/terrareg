package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// NamespaceRepository defines the interface for namespace persistence
type NamespaceRepository interface {
	// Save persists a namespace
	Save(ctx context.Context, namespace *model.Namespace) error

	// FindByID retrieves a namespace by ID
	FindByID(ctx context.Context, id int) (*model.Namespace, error)

	// FindByName retrieves a namespace by name
	FindByName(ctx context.Context, name types.NamespaceName) (*model.Namespace, error)

	// List retrieves namespaces with optional pagination
	// If opts is nil or opts.Limit is 0, returns all namespaces
	// Returns: namespaces, total count (for pagination meta), error
	List(ctx context.Context, opts *query.ListOptions) ([]*model.Namespace, int, error)

	// Delete removes a namespace
	Delete(ctx context.Context, id int) error

	// Exists checks if a namespace exists
	Exists(ctx context.Context, name types.NamespaceName) (bool, error)
}
