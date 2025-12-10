package analytics

import (
	"context"
	"fmt"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
)

// GetTokenVersionsQuery retrieves analytics token information for a module provider
type GetTokenVersionsQuery struct {
	analyticsRepo analyticsCmd.AnalyticsRepository
}

// NewGetTokenVersionsQuery creates a new query
func NewGetTokenVersionsQuery(analyticsRepo analyticsCmd.AnalyticsRepository) *GetTokenVersionsQuery {
	return &GetTokenVersionsQuery{
		analyticsRepo: analyticsRepo,
	}
}

// Execute retrieves token versions for a module provider
func (q *GetTokenVersionsQuery) Execute(ctx context.Context, namespace, name, provider string) (map[string]analyticsCmd.TokenVersionInfo, error) {
	// Get module provider ID first
	moduleProviderID, err := q.analyticsRepo.GetModuleProviderID(ctx, namespace, name, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get module provider ID: %w", err)
	}

	// Get latest analytics entries for each token
	tokenVersions, err := q.analyticsRepo.GetLatestTokenVersions(ctx, moduleProviderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token versions: %w", err)
	}

	return tokenVersions, nil
}