package service

import (
	"context"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
)

// AnalyticsService provides domain services for analytics
type AnalyticsService struct {
	analyticsRepository analyticsCmd.AnalyticsRepository
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(analyticsRepository analyticsCmd.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepository: analyticsRepository,
	}
}

// GetDownloadsByVersionID retrieves download count for a specific module version ID
func (s *AnalyticsService) GetDownloadsByVersionID(ctx context.Context, moduleVersionID int) (int, error) {
	return s.analyticsRepository.GetDownloadsByVersionID(ctx, moduleVersionID)
}