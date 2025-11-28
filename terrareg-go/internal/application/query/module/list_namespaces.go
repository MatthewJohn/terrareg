package module

import (
	"context"

	"github.com/terrareg/terrareg/internal/domain/module/model"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
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

// Execute executes the query
func (q *ListNamespacesQuery) Execute(ctx context.Context) ([]*model.Namespace, error) {
	return q.namespaceRepo.List(ctx)
}
