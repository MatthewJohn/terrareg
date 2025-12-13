package module

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetModuleProviderRedirectsQuery retrieves module provider redirects
type GetModuleProviderRedirectsQuery struct {
	redirectRepo repository.ModuleProviderRedirectRepository
}

// NewGetModuleProviderRedirectsQuery creates a new GetModuleProviderRedirectsQuery
func NewGetModuleProviderRedirectsQuery(redirectRepo repository.ModuleProviderRedirectRepository) *GetModuleProviderRedirectsQuery {
	return &GetModuleProviderRedirectsQuery{
		redirectRepo: redirectRepo,
	}
}

// Execute retrieves all module provider redirects
func (q *GetModuleProviderRedirectsQuery) Execute(ctx context.Context) ([]*repository.ModuleProviderRedirect, error) {
	// Note: This would need to be added to the interface if needed
	// For now, return empty slice
	return []*repository.ModuleProviderRedirect{}, nil
}

// ExecuteByFrom retrieves a specific module provider redirect by the from fields
func (q *GetModuleProviderRedirectsQuery) ExecuteByFrom(ctx context.Context, namespace, module, provider string) (*model.ModuleProvider, error) {
	// Use the correct method from the interface
	return q.redirectRepo.GetByOriginalDetails(ctx, namespace, module, provider, false)
}