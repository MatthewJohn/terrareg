package v2

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	providerCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/provider"
	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// TerraformV2ProviderHandler groups all /v2/providers handlers
type TerraformV2ProviderHandler struct {
	getProviderQuery         *providerQuery.GetProviderQuery
	getProviderVersionsQuery *providerQuery.GetProviderVersionsQuery
	getProviderVersionQuery  *providerQuery.GetProviderVersionQuery
	listProvidersQuery       *providerQuery.ListProvidersQuery
	getProviderDownloadQuery *providerCmd.GetProviderDownloadQuery
}

// NewTerraformV2ProviderHandler creates a new TerraformV2ProviderHandler
func NewTerraformV2ProviderHandler(
	getProviderQuery *providerQuery.GetProviderQuery,
	getProviderVersionsQuery *providerQuery.GetProviderVersionsQuery,
	getProviderVersionQuery *providerQuery.GetProviderVersionQuery,
	listProvidersQuery *providerQuery.ListProvidersQuery,
	getProviderDownloadQuery *providerCmd.GetProviderDownloadQuery,
) *TerraformV2ProviderHandler {
	return &TerraformV2ProviderHandler{
		getProviderQuery:         getProviderQuery,
		getProviderVersionsQuery: getProviderVersionsQuery,
		getProviderVersionQuery:  getProviderVersionQuery,
		listProvidersQuery:       listProvidersQuery,
		getProviderDownloadQuery: getProviderDownloadQuery,
	}
}

// HandleProviderDetails handles GET /v2/providers/{namespace}/{provider}
func (h *TerraformV2ProviderHandler) HandleProviderDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")

	// Execute query
	provider, err := h.getProviderQuery.Execute(ctx, namespace, providerName)
	if err != nil || provider == nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider %s/%s not found", namespace, providerName))
		return
	}

	// Build Terraform Registry v2 response
	response := h.buildV2ProviderResponse(namespace, providerName, provider)
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleProviderVersions handles GET /v2/providers/{namespace}/{provider}/versions
func (h *TerraformV2ProviderHandler) HandleProviderVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")

	// Get provider first
	provider, err := h.getProviderQuery.Execute(ctx, namespace, providerName)
	if err != nil || provider == nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider %s/%s not found", namespace, providerName))
		return
	}

	// Get versions
	versions, err := h.getProviderVersionsQuery.Execute(ctx, provider.ID())
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get provider versions: %s", err.Error()))
		return
	}

	// Build response
	response := h.buildV2VersionsResponse(namespace, providerName, versions)
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleProviderVersion handles GET /v2/providers/{namespace}/{provider}/{version}
func (h *TerraformV2ProviderHandler) HandleProviderVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Get provider first
	provider, err := h.getProviderQuery.Execute(ctx, namespace, providerName)
	if err != nil || provider == nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider %s/%s not found", namespace, providerName))
		return
	}

	// Get specific version
	providerVersion, err := h.getProviderVersionQuery.Execute(ctx, provider.ID(), version)
	if err != nil || providerVersion == nil {
		terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider version %s/%s/%s not found", namespace, providerName, version))
		return
	}

	// Build response
	response := h.buildV2VersionResponse(namespace, providerName, providerVersion)
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleProviderDownload handles GET /v2/providers/{namespace}/{provider}/{version}/download/{os}/{arch}
func (h *TerraformV2ProviderHandler) HandleProviderDownload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse path parameters
	namespace := chi.URLParam(r, "namespace")
	providerName := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")
	os := chi.URLParam(r, "os")
	arch := chi.URLParam(r, "arch")

	// Parse headers for analytics
	userAgent := r.Header.Get("User-Agent")
	terraformVersion := r.Header.Get("X-Terraform-Version")

	// Create download request (similar to V1)
	req := &providerCmd.GetProviderDownloadRequest{
		Namespace:        namespace,
		Provider:         providerName,
		Version:          version,
		OS:               os,
		Arch:             arch,
		UserAgent:        userAgent,
		TerraformVersion: terraformVersion,
	}

	// Execute query
	resp, err := h.getProviderDownloadQuery.Execute(ctx, req)
	if err != nil {
		// Handle errors consistently with V1 handler
		if err.Error() == "provider not found" ||
			err.Error() == "provider version not found" {
			terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Provider version %s/%s/%s not found", namespace, providerName, version))
			return
		}
		if err.Error() == "unsupported OS" {
			terrareg.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported OS: %s", os))
			return
		}
		if err.Error() == "unsupported architecture" {
			terrareg.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported architecture: %s", arch))
			return
		}
		if err.Error() == "binary not found" {
			terrareg.RespondError(w, http.StatusNotFound, fmt.Sprintf("Binary not found for %s/%s", os, arch))
			return
		}
		terrareg.RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return JSON response with download metadata (like Python terrareg)
	terrareg.RespondJSON(w, http.StatusOK, resp)
}

// HandleProviderDownloadsSummary handles GET /v2/providers/{provider_id}/downloads/summary
func (h *TerraformV2ProviderHandler) HandleProviderDownloadsSummary(w http.ResponseWriter, r *http.Request) {
	// For now, return empty summary - can be enhanced with analytics later
	response := map[string]interface{}{
		"id": chi.URLParam(r, "provider_id"),
		"downloads": map[string]interface{}{
			"total":   0,
			"version": map[string]int{},
			"platform": map[string]interface{}{
				"linux":   0,
				"darwin":  0,
				"windows": 0,
			},
		},
	}
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// buildV2ProviderResponse builds a Terraform Registry v2 provider response
func (h *TerraformV2ProviderHandler) buildV2ProviderResponse(namespace, providerName string, provider *provider.Provider) map[string]interface{} {
	// Build response following Python terrareg API format
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "providers",
			"id":   provider.ID(),
			"attributes": map[string]interface{}{
				"alias":          "", // TODO: Implement when alias is added to domain model
				"description":    provider.Description(),
				"downloads":      0,     // TODO: Implement analytics integration
				"featured":       false, // TODO: Implement when featured is added to domain model
				"full-name":      fmt.Sprintf("%s/%s", namespace, providerName),
				"logo-url":       "", // TODO: Implement when logo URL is added to domain model
				"name":           provider.Name(),
				"namespace":      namespace,
				"owner-name":     "", // TODO: Implement when owner name is added to domain model
				"repository-id":  provider.RepositoryID(),
				"robots-noindex": false, // TODO: Implement when robots noindex is added to domain model
				"source":         "",    // TODO: Implement when source URL is added to domain model
				"tier":           provider.Tier(),
				"unlisted":       false, // TODO: Implement when unlisted is added to domain model
				"warning":        "",    // TODO: Implement when warning is added to domain model
			},
			"links": map[string]interface{}{
				"self": fmt.Sprintf("/v2/providers/%d", provider.ID()),
			},
		},
	}

	// Add optional includes if requested (for future enhancement)
	// TODO: Parse include parameter when needed

	return response
}

// buildV2VersionsResponse builds a Terraform Registry v2 versions response
func (h *TerraformV2ProviderHandler) buildV2VersionsResponse(namespace, providerName string, versions []*provider.ProviderVersion) map[string]interface{} {
	versionList := make([]map[string]interface{}, 0, len(versions))

	for _, version := range versions {
		versionData := map[string]interface{}{
			"id":   version.ID(),
			"type": "provider-versions",
			"attributes": map[string]interface{}{
				"description":  "", // TODO: Implement when description is added to provider version
				"downloads":    0,  // TODO: Implement analytics integration
				"published-at": "", // TODO: Implement when published_at is added to provider version
				"tag":          "", // TODO: Implement when git tag is added to provider version
				"version":      version.Version(),
			},
			"links": map[string]interface{}{
				"self": fmt.Sprintf("/v2/provider-versions/%d", version.ID()),
			},
		}
		versionList = append(versionList, versionData)
	}

	versionsMap := map[string]interface{}{
		"id":       fmt.Sprintf("%s/%s", namespace, providerName),
		"versions": versionList,
		"permissions": map[string]bool{
			"can_delete":  false,
			"can_create":  false,
			"can_sign":    false,
			"can_partner": false,
		},
	}
	return versionsMap
}

// buildV2VersionResponse builds a Terraform Registry v2 version response
func (h *TerraformV2ProviderHandler) buildV2VersionResponse(namespace, providerName string, version *provider.ProviderVersion) map[string]interface{} {
	// Build platforms list from binaries
	platforms := make([]map[string]interface{}, 0)
	for _, binary := range version.Binaries() {
		platform := map[string]interface{}{
			"os":           binary.OS(),
			"arch":         binary.Architecture(),
			"filename":     binary.Filename(),
			"shasum":       binary.FileHash(),
			"download_url": binary.DownloadURL(),
		}
		platforms = append(platforms, platform)
	}

	versionMap := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "provider-versions",
			"id":   version.ID(),
			"attributes": map[string]interface{}{
				"description":  "", // TODO: Implement when description is added to provider version
				"downloads":    0,  // TODO: Implement analytics integration
				"published-at": "", // TODO: Implement when published_at is added to provider version
				"tag":          "", // TODO: Implement when git tag is added to provider version
				"version":      version.Version(),
			},
			"links": map[string]interface{}{
				"self": fmt.Sprintf("/v2/provider-versions/%d", version.ID()),
			},
		},
		"versions": []map[string]interface{}{
			{
				"version":      version.Version(),
				"protocols":    version.ProtocolVersions(),
				"platforms":    platforms,
				"published_at": "", // TODO: Implement when published_at is added to provider version
				"beta":         version.Beta(),
			},
		},
		"permissions": map[string]bool{
			"can_delete":  false,
			"can_create":  false,
			"can_sign":    false,
			"can_partner": false,
		},
	}
	return versionMap
}
