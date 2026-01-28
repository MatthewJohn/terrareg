package analytics

import (
	"context"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// GetDownloadSummaryQuery handles retrieving download summary
type GetDownloadSummaryQuery struct {
	analyticsRepo analyticsCmd.AnalyticsRepository
}

// NewGetDownloadSummaryQuery creates a new get download summary query
func NewGetDownloadSummaryQuery(analyticsRepo analyticsCmd.AnalyticsRepository) *GetDownloadSummaryQuery {
	return &GetDownloadSummaryQuery{
		analyticsRepo: analyticsRepo,
	}
}

// Execute retrieves the download summary for a module provider
func (q *GetDownloadSummaryQuery) Execute(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (*analyticsCmd.DownloadStats, error) {
	return q.analyticsRepo.GetDownloadStats(ctx, namespace, module, provider)
}
