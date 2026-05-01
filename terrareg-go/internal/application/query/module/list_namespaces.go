package module

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// ListNamespacesQuery handles listing all namespaces
type ListNamespacesQuery struct {
	namespaceRepo repository.NamespaceRepository
}

// NewListNamespacesQuery creates a new list namespaces query
func NewListNamespacesQuery(namespaceRepo repository.NamespaceRepository) *ListNamespacesQuery {
	return &ListNamespacesQuery{
		namespaceRepo: namespaceRepo,
	}
}

// Execute executes the query with optional pagination
// Returns: namespaces, total count (for pagination meta), error
func (q *ListNamespacesQuery) Execute(ctx context.Context, opts *query.ListOptions) ([]*model.Namespace, int, error) {
	return q.namespaceRepo.List(ctx, opts)
}
