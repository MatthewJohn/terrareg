package repository

import (
	"context"

	"github.com/terrareg/terrareg/internal/domain/module/model"
)

// NamespaceRepository defines the interface for namespace persistence
type NamespaceRepository interface {
	// Save persists a namespace
	Save(ctx context.Context, namespace *model.Namespace) error

	// FindByID retrieves a namespace by ID
	FindByID(ctx context.Context, id int) (*model.Namespace, error)

	// FindByName retrieves a namespace by name
	FindByName(ctx context.Context, name string) (*model.Namespace, error)

	// List retrieves all namespaces
	List(ctx context.Context) ([]*model.Namespace, error)

	// Delete removes a namespace
	Delete(ctx context.Context, id int) error

	// Exists checks if a namespace exists
	Exists(ctx context.Context, name string) (bool, error)
}
