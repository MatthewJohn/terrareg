package analytics

import (
	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
)

// Type aliases for analytics types from command package
type (
	AnalyticsRepository = analyticsCmd.AnalyticsRepository
	ModuleVersionInfo   = analyticsCmd.ModuleVersionInfo
	ModuleProviderInfo  = analyticsCmd.ModuleProviderInfo
)
