package namespace

import (
	"context"
	"fmt"

	namespaceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// NamespaceDetailsQuery handles getting namespace details
type NamespaceDetailsQuery struct {
	namespaceRepo namespaceRepo.NamespaceRepository
}

// NewNamespaceDetailsQuery creates a new namespace details query
func NewNamespaceDetailsQuery(namespaceRepo namespaceRepo.NamespaceRepository) *NamespaceDetailsQuery {
	return &NamespaceDetailsQuery{
		namespaceRepo: namespaceRepo,
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
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	if namespace == nil {
		return nil, nil
	}

	// TODO: Implement is_auto_verified and trusted based on config
	// For now, return false for both
	details := &NamespaceDetails{
		Name:           namespace.Name(),
		DisplayName:    namespace.DisplayName(),
		IsAutoVerified: false, // TODO: Check against VERIFIED_MODULE_NAMESPACES config
		Trusted:        false, // TODO: Check against TRUSTED_NAMESPACES config
	}

	return details, nil
}