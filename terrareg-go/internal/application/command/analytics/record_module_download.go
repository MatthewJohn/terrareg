package analytics

import (
	"context"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// RecordModuleDownloadCommand handles recording module downloads for analytics
type RecordModuleDownloadCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
	analyticsRepo      AnalyticsRepository
}

// NewRecordModuleDownloadCommand creates a new record module download command
func NewRecordModuleDownloadCommand(
	moduleProviderRepo repository.ModuleProviderRepository,
	analyticsRepo AnalyticsRepository,
) *RecordModuleDownloadCommand {
	return &RecordModuleDownloadCommand{
		moduleProviderRepo: moduleProviderRepo,
		analyticsRepo:      analyticsRepo,
	}
}

// RecordModuleDownloadRequest represents a request to record a module download
type RecordModuleDownloadRequest struct {
	Namespace        string
	Module           string
	Provider         string
	Version          string
	TerraformVersion *string
	AnalyticsToken   *string
	AuthToken        *string
	Environment      *string
}

// Execute records the module download
func (c *RecordModuleDownloadCommand) Execute(ctx context.Context, req RecordModuleDownloadRequest) error {
	// Find the module provider to validate it exists
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(ctx, req.Namespace, req.Module, req.Provider)
	if err != nil {
		// Silently fail for analytics - don't block downloads
		return nil
	}

	// Find the version
	version, err := moduleProvider.GetVersion(req.Version)
	if err != nil || version == nil {
		// Silently fail for analytics
		return nil
	}

	// Record the analytics event
	now := time.Now()
	event := AnalyticsEvent{
		ParentModuleVersionID: version.ID(),
		Timestamp:             &now,
		TerraformVersion:      req.TerraformVersion,
		AnalyticsToken:        req.AnalyticsToken,
		AuthToken:             req.AuthToken,
		Environment:           req.Environment,
		NamespaceName:         &req.Namespace,
		ModuleName:            &req.Module,
		ProviderName:          &req.Provider,
	}

	return c.analyticsRepo.RecordDownload(ctx, event)
}
