package namespace

import (
	"context"
	"fmt"

	appConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	namespaceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// NamespaceDetailsQuery handles getting namespace details
type NamespaceDetailsQuery struct {
	namespaceRepo namespaceRepo.NamespaceRepository
	config         *appConfig.Config
}

// NewNamespaceDetailsQuery creates a new namespace details query
func NewNamespaceDetailsQuery(namespaceRepo namespaceRepo.NamespaceRepository, config *appConfig.Config) *NamespaceDetailsQuery {
	return &NamespaceDetailsQuery{
		namespaceRepo: namespaceRepo,
		config:         config,
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

	// Check if namespace is in verified list
	isAutoVerified := false
	for _, ns := range q.config.VerifiedModuleNamespaces {
		if ns == namespaceName {
			isAutoVerified = true
			break
		}
	}

	// Check if namespace is in trusted list
	trusted := false
	for _, ns := range q.config.TrustedNamespaces {
		if ns == namespaceName {
			trusted = true
			break
		}
	}

	details := &NamespaceDetails{
		Name:           namespace.Name(),
		DisplayName:    namespace.DisplayName(),
		IsAutoVerified: isAutoVerified,
		Trusted:        trusted,
	}

	return details, nil
}