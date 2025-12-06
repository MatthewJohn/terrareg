package terrareg

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	analyticsQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/analytics"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// AnalyticsHandler handles analytics-related requests
type AnalyticsHandler struct {
	globalStatsQuery               *analyticsQuery.GlobalStatsQuery
	getDownloadSummaryQuery        *analyticsQuery.GetDownloadSummaryQuery
	recordModuleDownloadCmd        *analyticsCmd.RecordModuleDownloadCommand
	getMostRecentlyPublishedQuery  *analyticsQuery.GetMostRecentlyPublishedQuery
	getMostDownloadedThisWeekQuery *analyticsQuery.GetMostDownloadedThisWeekQuery
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(
	globalStatsQuery *analyticsQuery.GlobalStatsQuery,
	getDownloadSummaryQuery *analyticsQuery.GetDownloadSummaryQuery,
	recordModuleDownloadCmd *analyticsCmd.RecordModuleDownloadCommand,
	getMostRecentlyPublishedQuery *analyticsQuery.GetMostRecentlyPublishedQuery,
	getMostDownloadedThisWeekQuery *analyticsQuery.GetMostDownloadedThisWeekQuery,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		globalStatsQuery:               globalStatsQuery,
		getDownloadSummaryQuery:        getDownloadSummaryQuery,
		recordModuleDownloadCmd:        recordModuleDownloadCmd,
		getMostRecentlyPublishedQuery:  getMostRecentlyPublishedQuery,
		getMostDownloadedThisWeekQuery: getMostDownloadedThisWeekQuery,
	}
}

// HandleGlobalStatsSummary handles GET /v1/terrareg/analytics/global/stats_summary
func (h *AnalyticsHandler) HandleGlobalStatsSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.globalStatsQuery.Execute(ctx)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, stats)
}

// HandleMostRecentlyPublished handles GET /v1/terrareg/analytics/global/most_recently_published_module_version
func (h *AnalyticsHandler) HandleMostRecentlyPublished(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	info, err := h.getMostRecentlyPublishedQuery.Execute(ctx)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return 404 if no module found (matching Python behavior)
	if info == nil {
		RespondJSON(w, http.StatusNotFound, map[string]interface{}{})
		return
	}

	// Build response matching Python's get_api_outline()
	response := map[string]interface{}{
		"id":           info.ID,
		"namespace":    info.Namespace,
		"name":         info.Module,  // Python uses "name" not "module"
		"provider":     info.Provider,
		"version":      info.Version,
		"owner":        info.Owner,
		"description":  info.Description,
		"source":       info.Source,
		"published_at": info.PublishedAt,
		"downloads":    info.Downloads,
		"internal":     info.Internal,
		"trusted":      info.Trusted,
		"verified":     info.Verified,
	}

	RespondJSON(w, http.StatusOK, response)
}

// HandleMostDownloadedThisWeek handles GET /v1/terrareg/analytics/global/most_downloaded_module_provider_this_week
func (h *AnalyticsHandler) HandleMostDownloadedThisWeek(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	info, err := h.getMostDownloadedThisWeekQuery.Execute(ctx)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return 404 if no module found (matching Python behavior)
	if info == nil {
		RespondJSON(w, http.StatusNotFound, map[string]interface{}{})
		return
	}

	// Build response
	response := map[string]interface{}{
		"namespace": info.Namespace,
		"module":    info.Module,
		"provider":  info.Provider,
		"downloads": info.DownloadCount,
	}

	RespondJSON(w, http.StatusOK, response)
}

// HandleModuleDownloadsSummary handles GET /v1/modules/{namespace}/{name}/{provider}/downloads/summary
func (h *AnalyticsHandler) HandleModuleDownloadsSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")

	// Execute query
	stats, err := h.getDownloadSummaryQuery.Execute(ctx, namespace, name, provider)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Build response in Terraform Registry format
	id := fmt.Sprintf("%s/%s/%s", namespace, name, provider)
	response := dto.DownloadSummaryResponse{
		Data: dto.DownloadData{
			Type: "module-downloads",
			ID:   id,
			Attributes: dto.DownloadAttributes{
				Total: stats.TotalDownloads,
			},
		},
	}

	RespondJSON(w, http.StatusOK, response)
}

// RecordModuleDownload records a module download (called during download)
func (h *AnalyticsHandler) RecordModuleDownload(ctx context.Context, namespace, module, provider, version string, r *http.Request) {
	// Extract optional analytics parameters from headers
	terraformVersion := r.Header.Get("X-Terraform-Version")
	var tfVersionPtr *string
	if terraformVersion != "" {
		tfVersionPtr = &terraformVersion
	}

	// Execute command to record the download
	req := analyticsCmd.RecordModuleDownloadRequest{
		Namespace:        namespace,
		Module:           module,
		Provider:         provider,
		Version:          version,
		TerraformVersion: tfVersionPtr,
	}

	// Don't fail the download if analytics fails - just log silently
	_ = h.recordModuleDownloadCmd.Execute(ctx, req)
}
