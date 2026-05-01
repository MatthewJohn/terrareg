package service

import "context"

// InfracostService handles cost analysis of module examples using Infracost
type InfracostService interface {
	// AnalyzeExample runs infracost on a module example and returns the JSON results
	// Returns (nil, nil) if infracost is not configured (not an error)
	// Returns (results, nil) on success
	// Returns (nil, error) on execution failure
	AnalyzeExample(ctx context.Context, examplePath string) ([]byte, error)

	// IsAvailable returns true if infracost API key is configured
	IsAvailable() bool
}
