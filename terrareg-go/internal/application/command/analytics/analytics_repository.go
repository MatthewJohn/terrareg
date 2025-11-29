package analytics

import (
	"context"
	"time"
)

// AnalyticsRepository defines the interface for analytics persistence
type AnalyticsRepository interface {
	// RecordDownload records a module download event
	RecordDownload(ctx context.Context, event AnalyticsEvent) error

	// GetDownloadStats retrieves download statistics for a module provider
	GetDownloadStats(ctx context.Context, namespace, module, provider string) (*DownloadStats, error)

	// GetMostRecentlyPublished retrieves the most recently published module version
	GetMostRecentlyPublished(ctx context.Context) (*ModuleVersionInfo, error)

	// GetMostDownloadedThisWeek retrieves the most downloaded module provider this week
	GetMostDownloadedThisWeek(ctx context.Context) (*ModuleProviderInfo, error)
}

// AnalyticsEvent represents a module download analytics event
type AnalyticsEvent struct {
	ParentModuleVersionID int
	Timestamp             *time.Time
	TerraformVersion      *string
	AnalyticsToken        *string
	AuthToken             *string
	Environment           *string
	NamespaceName         *string
	ModuleName            *string
	ProviderName          *string
}

// DownloadStats represents download statistics
type DownloadStats struct {
	TotalDownloads  int
	RecentDownloads int // Last 30 days
}

// ModuleVersionInfo represents information about a module version
type ModuleVersionInfo struct {
	Namespace string
	Module    string
	Provider  string
	Version   string
}

// ModuleProviderInfo represents information about a module provider with download count
type ModuleProviderInfo struct {
	Namespace      string
	Module         string
	Provider       string
	DownloadCount  int
}
