package analytics

import (
	"context"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// AnalyticsRepository defines the interface for analytics persistence
type AnalyticsRepository interface {
	// RecordDownload records a module download event
	RecordDownload(ctx context.Context, event AnalyticsEvent) error

	// RecordProviderDownload records a provider download event
	RecordProviderDownload(ctx context.Context, event ProviderDownloadEvent) error

	// GetDownloadStats retrieves download statistics for a module provider
	GetDownloadStats(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (*DownloadStats, error)

	// GetDownloadsByVersionID retrieves download count for a specific module version ID
	GetDownloadsByVersionID(ctx context.Context, moduleVersionID int) (int, error)

	// GetTotalDownloads retrieves the total count of all analytics records
	// Matches Python: AnalyticsEngine.get_total_downloads()
	GetTotalDownloads(ctx context.Context) (int, error)

	// GetMostRecentlyPublished retrieves the most recently published module version
	GetMostRecentlyPublished(ctx context.Context) (*ModuleVersionInfo, error)

	// GetMostDownloadedThisWeek retrieves the most downloaded module provider this week
	GetMostDownloadedThisWeek(ctx context.Context) (*ModuleProviderInfo, error)

	// GetModuleProviderID retrieves the ID for a module provider
	GetModuleProviderID(ctx context.Context, namespace types.NamespaceName, module types.ModuleName, provider types.ModuleProviderName) (int, error)

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
	NamespaceName         *types.NamespaceName
	ModuleName            *types.ModuleName
	ProviderName          *types.ModuleProviderName
}

// ProviderDownloadEvent represents a provider download analytics event
type ProviderDownloadEvent struct {
	ProviderVersionID int
	Timestamp         *time.Time
	TerraformVersion  *string
	AnalyticsToken    *string
	AuthToken         *string
	Environment       *string
	NamespaceName     *types.NamespaceName
	ProviderName      *types.ModuleProviderName
	Version           *types.ModuleVersion
	OS                *string
	Architecture      *string
	UserAgent         *string
}

// DownloadStats represents download statistics
// Matches Python: AnalyticsEngine.get_module_provider_download_stats()
type DownloadStats struct {
	TotalDownloads int
	Week           int // Last 7 days
	Month          int // Last 31 days
	Year           int // Last 365 days
}

// ModuleVersionInfo represents information about a module version
type ModuleVersionInfo struct {
	ID          string                   `json:"id"`        // Format: "provider_id/version" (from Python)
	Namespace   types.NamespaceName      `json:"namespace"` // From ModuleProvider.get_api_outline()
	Module      types.ModuleName         `json:"name"`      // Python uses "name" not "module"
	Provider    types.ModuleProviderName `json:"provider"`  // From ModuleProvider.get_api_outline()
	Version     types.ModuleVersion      `json:"version"`   // Version-specific
	Owner       *string                  `json:"owner,omitempty"`
	Description *string                  `json:"description,omitempty"`
	Source      *string                  `json:"source,omitempty"`       // From get_source_base_url()
	PublishedAt *string                  `json:"published_at,omitempty"` // ISO format from .isoformat()
	Downloads   int                      `json:"downloads"`
	Internal    bool                     `json:"internal"`
	Trusted     bool                     `json:"trusted"`  // From ModuleProvider.get_api_outline()
	Verified    bool                     `json:"verified"` // From ModuleProvider.get_api_outline()
}

// ModuleProviderInfo represents information about a module provider with download count
type ModuleProviderInfo struct {
	Namespace     types.NamespaceName
	Module        types.ModuleName
	Provider      types.ModuleProviderName
	DownloadCount int
}

// TokenVersionInfo represents information about a token's latest usage
type TokenVersionInfo struct {
	TerraformVersion string              `json:"terraform_version"`
	ModuleVersion    types.ModuleVersion `json:"module_version"`
	Environment      *string             `json:"environment"`
}
