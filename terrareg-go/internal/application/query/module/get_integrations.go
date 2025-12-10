package module

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// GetIntegrationsQuery retrieves integrations for a module provider
type GetIntegrationsQuery struct {
	moduleProviderRepo repository.ModuleProviderRepository
}

// NewGetIntegrationsQuery creates a new query
func NewGetIntegrationsQuery(moduleProviderRepo repository.ModuleProviderRepository) *GetIntegrationsQuery {
	return &GetIntegrationsQuery{
		moduleProviderRepo: moduleProviderRepo,
	}
}

// Integration represents an available integration for a module provider
type Integration struct {
	Description string  `json:"description"`
	Method      *string `json:"method"`
	URL         string  `json:"url"`
	Notes       string  `json:"notes"`
	ComingSoon  *bool   `json:"coming_soon,omitempty"`
}

// Execute retrieves integrations for a module provider
func (q *GetIntegrationsQuery) Execute(ctx context.Context, namespace, name, provider string) ([]Integration, error) {
	// Find the module provider to get its ID
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, namespace, name, provider)
	if err != nil {
		return nil, err
	}

	// Build integrations list based on module provider
	var integrations []Integration
	moduleProviderID := moduleProvider.ID()

	// Import integration
	importMethod := "POST"
	integrations = append(integrations, Integration{
		Description: "Trigger module version import",
		Method:      &importMethod,
		URL:         fmt.Sprintf("/v1/terrareg/modules/%d/${version}/import", moduleProviderID),
		Notes:       "",
	})

	// Upload integration (if module hosting is allowed)
	uploadMethod := "POST"
	integrations = append(integrations, Integration{
		Description: "Create module version using source archive",
		Method:      &uploadMethod,
		URL:         fmt.Sprintf("/v1/terrareg/modules/%d/${version}/upload", moduleProviderID),
		Notes:       "",
	})

	// Publish integration
	publishMethod := "POST"
	integrations = append(integrations, Integration{
		Description: "Mark a module version as published",
		Method:      &publishMethod,
		URL:         fmt.Sprintf("/v1/terrareg/modules/%d/${version}/publish", moduleProviderID),
		Notes:       "",
	})

	// Webhooks
	// GitHub hook
	integrations = append(integrations, Integration{
		Description: "GitHub hook trigger",
		Method:      nil, // Webhook endpoint, no method specified
		URL:         fmt.Sprintf("/v1/terrareg/modules/%d/hooks/github", moduleProviderID),
		Notes:       "",
	})

	// Bitbucket hook
	integrations = append(integrations, Integration{
		Description: "Bitbucket hook trigger",
		Method:      nil,
		URL:         fmt.Sprintf("/v1/terrareg/modules/%d/hooks/bitbucket", moduleProviderID),
		Notes:       "",
	})

	// GitLab hook (coming soon)
	comingSoon := true
	integrations = append(integrations, Integration{
		Description: "Gitlab hook trigger",
		Method:      nil,
		URL:         fmt.Sprintf("/v1/terrareg/modules/%d/hooks/gitlab", moduleProviderID),
		Notes:       "",
		ComingSoon:  &comingSoon,
	})

	return integrations, nil
}