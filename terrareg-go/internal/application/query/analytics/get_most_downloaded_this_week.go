package analytics

import (
	"context"
)

// GetMostDownloadedThisWeekQuery handles retrieving the most downloaded module provider this week
type GetMostDownloadedThisWeekQuery struct {
	analyticsRepo AnalyticsRepository
}

// NewGetMostDownloadedThisWeekQuery creates a new query
func NewGetMostDownloadedThisWeekQuery(analyticsRepo AnalyticsRepository) *GetMostDownloadedThisWeekQuery {
	return &GetMostDownloadedThisWeekQuery{
		analyticsRepo: analyticsRepo,
	}
}

// Execute retrieves the most downloaded module provider this week
func (q *GetMostDownloadedThisWeekQuery) Execute(ctx context.Context) (*ModuleProviderInfo, error) {
	return q.analyticsRepo.GetMostDownloadedThisWeek(ctx)
}
