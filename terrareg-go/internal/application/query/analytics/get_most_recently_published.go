package analytics

import (
	"context"
)

// GetMostRecentlyPublishedQuery handles retrieving the most recently published module version
type GetMostRecentlyPublishedQuery struct {
	analyticsRepo AnalyticsRepository
}

// NewGetMostRecentlyPublishedQuery creates a new query
func NewGetMostRecentlyPublishedQuery(analyticsRepo AnalyticsRepository) *GetMostRecentlyPublishedQuery {
	return &GetMostRecentlyPublishedQuery{
		analyticsRepo: analyticsRepo,
	}
}

// Execute retrieves the most recently published module version
func (q *GetMostRecentlyPublishedQuery) Execute(ctx context.Context) (*ModuleVersionInfo, error) {
	return q.analyticsRepo.GetMostRecentlyPublished(ctx)
}
