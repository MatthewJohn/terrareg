package namespace

import (
	"context"
	"errors"

	namespaceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	namespaceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// NamespaceDetailsQuery handles getting namespace details
type NamespaceDetailsQuery struct {
	namespaceRepo    namespaceRepo.NamespaceRepository
	namespaceService *namespaceService.NamespaceService
}

// NewNamespaceDetailsQuery creates a new namespace details query
func NewNamespaceDetailsQuery(namespaceRepo namespaceRepo.NamespaceRepository, namespaceService *namespaceService.NamespaceService) *NamespaceDetailsQuery {
	return &NamespaceDetailsQuery{
		namespaceRepo:    namespaceRepo,
		namespaceService: namespaceService,
	}
}

// NamespaceDetails represents namespace details
type NamespaceDetails struct {
	Name           string  `json:"name"`
	DisplayName    *string `json:"display_name,omitempty"`
	IsAutoVerified bool    `json:"is_auto_verified"`
	Trusted        bool    `json:"trusted"`
}

// Execute executes the query
func (q *NamespaceDetailsQuery) Execute(ctx context.Context, namespaceName string) (*NamespaceDetails, error) {
	// Get namespace
	namespace, err := q.namespaceRepo.FindByName(ctx, namespaceName)
	if err != nil {
		if errors.Is(err, shared.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if namespace == nil {
		return nil, nil
	}

	details := &NamespaceDetails{
		Name:           namespace.Name(),
		DisplayName:    namespace.DisplayName(),
		IsAutoVerified: q.namespaceService.IsAutoVerified(namespace),
		Trusted:        q.namespaceService.IsTrusted(namespace),
	}

	return details, nil
}
