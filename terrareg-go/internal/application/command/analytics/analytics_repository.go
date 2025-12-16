package analytics

import (
	"context"
	"time"
)

// AnalyticsRepository defines the interface for analytics persistence
type AnalyticsRepository interface {
	// RecordDownload records a module download event
	RecordDownload(ctx context.Context, event AnalyticsEvent) error

	// RecordProviderDownload records a provider download event
	RecordProviderDownload(ctx context.Context, event ProviderDownloadEvent) error

	// GetDownloadStats retrieves download statistics for a module provider
	GetDownloadStats(ctx context.Context, namespace, module, provider string) (*DownloadStats, error)

	// GetDownloadsByVersionID retrieves download count for a specific module version ID
	GetDownloadsByVersionID(ctx context.Context, moduleVersionID int) (int, error)

	// GetMostRecentlyPublished retrieves the most recently published module version
	GetMostRecentlyPublished(ctx context.Context) (*ModuleVersionInfo, error)

	// GetMostDownloadedThisWeek retrieves the most downloaded module provider this week
	GetMostDownloadedThisWeek(ctx context.Context) (*ModuleProviderInfo, error)

	// GetModuleProviderID retrieves the ID for a module provider
	GetModuleProviderID(ctx context.Context, namespace, module, provider string) (int, error)

	// GetLatestTokenVersions retrieves the latest analytics entry for each token for a module provider
	GetLatestTokenVersions(ctx context.Context, moduleProviderID int) (map[string]TokenVersionInfo, error)
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

// ProviderDownloadEvent represents a provider download analytics event
type ProviderDownloadEvent struct {
	ProviderVersionID     int
	Timestamp             *time.Time
	TerraformVersion      *string
	AnalyticsToken        *string
	AuthToken             *string
	Environment           *string
	NamespaceName         *string
	ProviderName          *string
	Version               *string
	OS                    *string
	Architecture          *string
	UserAgent             *string
}

// DownloadStats represents download statistics
type DownloadStats struct {
	TotalDownloads  int
	RecentDownloads int // Last 30 days
}

// ModuleVersionInfo represents information about a module version
type ModuleVersionInfo struct {
	ID          string  `json:"id"`        // Format: "provider_id/version" (from Python)
	Namespace   string  `json:"namespace"` // From ModuleProvider.get_api_outline()
	Module      string  `json:"name"`      // Python uses "name" not "module"
	Provider    string  `json:"provider"`  // From ModuleProvider.get_api_outline()
	Version     string  `json:"version"`   // Version-specific
	Owner       *string `json:"owner,omitempty"`
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`       // From get_source_base_url()
	PublishedAt *string `json:"published_at,omitempty"` // ISO format from .isoformat()
	Downloads   int     `json:"downloads"`
	Internal    bool    `json:"internal"`
	Trusted     bool    `json:"trusted"`  // From ModuleProvider.get_api_outline()
	Verified    bool    `json:"verified"` // From ModuleProvider.get_api_outline()
}

// ModuleProviderInfo represents information about a module provider with download count
type ModuleProviderInfo struct {
	Namespace     string
	Module        string
	Provider      string
	DownloadCount int
}

// TokenVersionInfo represents information about a token's latest usage
type TokenVersionInfo struct {
	TerraformVersion string  `json:"terraform_version"`
	ModuleVersion    string  `json:"module_version"`
	Environment      *string `json:"environment"`
}
