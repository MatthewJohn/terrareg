package analytics

import (
	"context"
	"time"

	"gorm.io/gorm"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// AnalyticsRepositoryImpl implements the analytics repository
type AnalyticsRepositoryImpl struct {
	db *gorm.DB
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepositoryImpl {
	return &AnalyticsRepositoryImpl{db: db}
}

// RecordDownload records a module download event
func (r *AnalyticsRepositoryImpl) RecordDownload(ctx context.Context, event analyticsCmd.AnalyticsEvent) error {
	analytics := sqldb.AnalyticsDB{
		ParentModuleVersion: event.ParentModuleVersionID,
		Timestamp:           event.Timestamp,
		TerraformVersion:    event.TerraformVersion,
		AnalyticsToken:      event.AnalyticsToken,
		AuthToken:           event.AuthToken,
		Environment:         event.Environment,
		NamespaceName:       event.NamespaceName,
		ModuleName:          event.ModuleName,
		ProviderName:        event.ProviderName,
	}

	return r.db.WithContext(ctx).Create(&analytics).Error
}

// GetDownloadStats retrieves download statistics for a module provider
func (r *AnalyticsRepositoryImpl) GetDownloadStats(ctx context.Context, namespace, module, provider string) (*analyticsCmd.DownloadStats, error) {
	var totalCount int64
	var recentCount int64

	// Get total downloads
	err := r.db.WithContext(ctx).
		Model(&sqldb.AnalyticsDB{}).
		Where("namespace_name = ? AND module_name = ? AND provider_name = ?", namespace, module, provider).
		Count(&totalCount).Error
	if err != nil {
		return nil, err
	}

	// Get recent downloads (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err = r.db.WithContext(ctx).
		Model(&sqldb.AnalyticsDB{}).
		Where("namespace_name = ? AND module_name = ? AND provider_name = ? AND timestamp >= ?",
			namespace, module, provider, thirtyDaysAgo).
		Count(&recentCount).Error
	if err != nil {
		return nil, err
	}

	return &analyticsCmd.DownloadStats{
		TotalDownloads:  int(totalCount),
		RecentDownloads: int(recentCount),
	}, nil
}

// GetMostRecentlyPublished retrieves the most recently published module version
func (r *AnalyticsRepositoryImpl) GetMostRecentlyPublished(ctx context.Context) (*analyticsCmd.ModuleVersionInfo, error) {
	var result struct {
		Namespace string
		Module    string
		Provider  string
		Version   string
	}

	err := r.db.WithContext(ctx).
		Table("module_version").
		Select("namespace.namespace AS namespace, module.module AS module, module_provider.provider AS provider, module_version.version AS version").
		Joins("JOIN module_provider ON module_version.module_provider_id = module_provider.id").
		Joins("JOIN module ON module_provider.module_id = module.id").
		Joins("JOIN namespace ON module.namespace_id = namespace.id").
		Where("module_version.published_at IS NOT NULL").
		Order("module_version.published_at DESC").
		Limit(1).
		Scan(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// Return nil if no result found
	if result.Namespace == "" {
		return nil, nil
	}

	return &analyticsCmd.ModuleVersionInfo{
		Namespace: result.Namespace,
		Module:    result.Module,
		Provider:  result.Provider,
		Version:   result.Version,
	}, nil
}

// GetMostDownloadedThisWeek retrieves the most downloaded module provider this week
func (r *AnalyticsRepositoryImpl) GetMostDownloadedThisWeek(ctx context.Context) (*analyticsCmd.ModuleProviderInfo, error) {
	var result struct {
		Namespace     string
		Module        string
		Provider      string
		DownloadCount int
	}

	// Calculate the start of the current week (Sunday)
	now := time.Now()
	weekday := int(now.Weekday())
	startOfWeek := now.AddDate(0, 0, -weekday).Truncate(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Table("analytics").
		Select("namespace_name AS namespace, module_name AS module, provider_name AS provider, COUNT(*) AS download_count").
		Where("timestamp >= ?", startOfWeek).
		Where("namespace_name IS NOT NULL AND module_name IS NOT NULL AND provider_name IS NOT NULL").
		Group("namespace_name, module_name, provider_name").
		Order("download_count DESC").
		Limit(1).
		Scan(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// Return nil if no result found
	if result.Namespace == "" {
		return nil, nil
	}

	return &analyticsCmd.ModuleProviderInfo{
		Namespace:     result.Namespace,
		Module:        result.Module,
		Provider:      result.Provider,
		DownloadCount: result.DownloadCount,
	}, nil
}
