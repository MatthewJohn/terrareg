package dto

import (
	"fmt"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// ProviderListResponse represents the provider list API response
// Endpoints:
//   - GET /v1/providers (Terraform v1 provider list)
//   - GET /v1/providers/search (Terraform v1 provider search)
// Python reference: ApiProviderList, ApiProviderSearch
type ProviderListResponse struct {
	Meta      PaginationMeta `json:"meta"`
	Providers []ProviderData `json:"providers"`
	Count     *int           `json:"count,omitempty"` // Only included when include_count=true (matching Python)
}

// ProviderData represents a single provider in list responses
// Matches Python's get_api_outline() response structure from ProviderVersion
type ProviderData struct {
	ID          string  `json:"id"`           // Version ID (not provider ID) - matching Python
	Owner       string  `json:"owner"`        // Repository owner - matching Python
	Namespace   string  `json:"namespace"`    // Namespace name - matching Python
	Name        string  `json:"name"`         // Provider name - matching Python
	Alias       *string `json:"alias"`        // Always null in Python - matching Python
	Version     string  `json:"version"`      // Version string - matching Python
	Tag         *string `json:"tag"`          // Git tag - matching Python
	Description *string `json:"description"`  // Repository description - matching Python
	Source      *string `json:"source"`       // Repository source URL - matching Python
	PublishedAt *string `json:"published_at"` // Version published date - matching Python
	Downloads   int64   `json:"downloads"`    // Total downloads - matching Python
	Tier        string  `json:"tier"`         // Provider tier - matching Python
	LogoURL     *string `json:"logo_url"`     // Provider logo URL - matching Python
}

// ProviderDetailResponse represents a single provider detail response
// Endpoints:
//   - GET /v1/providers/{namespace}/{provider} (Terraform v1 provider details)
// Python reference: ApiProvider - provider_version_model.py get_api_details()
type ProviderDetailResponse struct {
	ID          string   `json:"id"`
	Owner       string   `json:"owner"`
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	Alias       *string  `json:"alias"`
	Version     string   `json:"version"`       // Latest version
	Tag         *string  `json:"tag"`
	Description *string  `json:"description,omitempty"`
	Source      *string  `json:"source,omitempty"`
	PublishedAt *string  `json:"published_at,omitempty"`
	Downloads   int64    `json:"downloads"`
	Tier        string   `json:"tier"`
	LogoURL     *string  `json:"logo_url,omitempty"`
	Versions    []string `json:"versions,omitempty"` // All version strings - CRITICAL for frontend
	Docs        []Doc    `json:"docs,omitempty"`     // Documentation array
}

// Doc represents documentation for a provider version
// Python reference: provider_version_documentation_model.py get_api_outline()
type Doc struct {
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	Title    *string `json:"title,omitempty"`
	Category *string `json:"category,omitempty"`
}

// ProviderVersionsResponse represents the versions list for a provider
// Endpoints:
//   - GET /v1/providers/{namespace}/{provider}/versions (Terraform v1 versions list)
// Python reference: ApiProviderVersions
type ProviderVersionsResponse struct {
	ID       string            `json:"id"`
	Versions []ProviderVersion `json:"versions"`
}

// ProviderVersion represents a single provider version
type ProviderVersion struct {
	Version   string   `json:"version"`
	Protocols []string `json:"protocols"`
	Platforms []string `json:"platforms,omitempty"`
}

// ProviderVersionDetailResponse represents a single provider version detail response
// Endpoints:
//   - GET /v1/providers/{namespace}/{provider}/{version} (Terraform v1 version details)
// Python reference: ApiProvider (with version path parameter)
type ProviderVersionDetailResponse struct {
	ID       string            `json:"id"`
	Version  string            `json:"version"`
	Beta     bool              `json:"beta"`
	Protocols []string          `json:"protocols"`
	Binaries []ProviderBinary `json:"binaries,omitempty"`
}

// ProviderBinary represents provider binary information for download
type ProviderBinary struct {
	Filename    string `json:"filename"`
	SHASUM      string `json:"shasum"`
	DownloadURL string `json:"download_url"`
}

// ProviderDownloadResponse represents download metadata for a provider binary
// Endpoints:
//   - GET /v1/providers/{namespace}/{provider}/{version}/download/{os}/{arch}
//   - GET /v2/providers/{namespace}/{provider}/{version}/download/{os}/{arch}
// Python reference: ApiProviderVersionDownload
type ProviderDownloadResponse struct {
	DownloadURL string `json:"download_url"`
	Filename     string `json:"filename"`
	SHASUM       string `json:"shasum"`
}

// NewProviderListResponse creates a provider list response from domain models
// Matches Python structure: returns version outline, not provider info
func NewProviderListResponse(providers []*provider.Provider, namespaceNames map[int]string, versionDataMap map[int]providerRepo.VersionData, total, offset, limit int, includeCount bool) ProviderListResponse {
	providerData := make([]ProviderData, 0, len(providers))
	for _, p := range providers {
		// Get namespace name from provider's namespace entity
		// The namespace should be populated by the repository
		nsEntity := p.Namespace()
		if nsEntity == nil {
			panic(fmt.Sprintf("CRITICAL: Provider %d (%s) has nil namespace entity! Provider ptr=%p", p.ID(), p.Name(), p))
		}
		namespace := string(nsEntity.Name())

		// Validate provider name
		providerName := p.Name()
		if providerName == "" {
			panic(fmt.Sprintf("CRITICAL: Provider %d has empty name!", p.ID()))
		}

		// Get version data for this provider
		versionData, hasVersion := versionDataMap[p.ID()]

		// Skip providers without version data (matching Python behavior)
		// Python: "if module_provider.get_latest_version()"
		if !hasVersion {
			continue
		}

		// Build response matching Python's get_api_outline()
		// ID is in format {namespace}/{provider}/{version} (Python: ProviderVersion.id property)
		// Python reference: /app/terrareg/provider_version_model.py - ProviderVersion.id
		// Use domain method Provider.VersionID() to generate the formatted ID
		id := p.VersionID(namespace, versionData.Version)

		data := ProviderData{
			ID:          id,
			Owner:       derefString(versionData.RepositoryOwner),
			Namespace:   namespace,
			Name:        providerName,
			Alias:       nil, // Always null in Python
			Version:     versionData.Version,
			Tag:         versionData.GitTag,
			Description: versionData.RepositoryDescription,                  // Repository description, not provider description
			Source:      getPublicSourceURL(versionData.RepositoryCloneURL), // Convert clone URL to public source URL
			PublishedAt: versionData.PublishedAt,
			Downloads:   versionData.Downloads,
			Tier:        p.Tier(),
			LogoURL:     versionData.RepositoryLogoURL,
		}
		providerData = append(providerData, data)
	}

	// Build pagination meta matching Python ResultData.meta
	meta := PaginationMeta{
		Limit:         limit,
		CurrentOffset: offset,
	}

	// Add prev_offset if current offset > 0 (matching Python)
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		meta.PrevOffset = &prevOffset
	}

	// Add next_offset if there are more results (matching Python)
	if total > offset+limit {
		nextOffset := offset + limit
		meta.NextOffset = &nextOffset
	}

	response := ProviderListResponse{
		Meta:      meta,
		Providers: providerData,
	}

	// Only include Count when include_count=true (matching Python behavior)
	if includeCount {
		response.Count = &total
	}

	return response
}

// derefString safely dereferences a string pointer
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// getPublicSourceURL converts a git clone URL to a public source URL
// by removing the .git suffix if present (matching Python's get_public_source_url)
func getPublicSourceURL(cloneURL *string) *string {
	if cloneURL == nil {
		return nil
	}
	url := *cloneURL
	// Remove .git suffix if present (matching Python behavior)
	if len(url) > 4 && url[len(url)-4:] == ".git" {
		trimmed := url[:len(url)-4]
		return &trimmed
	}
	return cloneURL
}

// NewProviderDetailResponse creates a provider detail response from domain model
// Matches Python's provider_version_model.py get_api_details() response structure
func NewProviderDetailResponse(p *provider.Provider, versions []*provider.ProviderVersion, docs []*provider.ProviderVersionDocumentation, totalDownloads int64) ProviderDetailResponse {
	// Get namespace name from provider's namespace entity
	// The namespace should be populated by the repository's toDomainProvider conversion
	// Data integrity: provider without namespace indicates database corruption
	ns := p.Namespace()
	if ns == nil {
		panic(fmt.Sprintf("data integrity error: provider %d has nil namespace - ensure repository populates namespace", p.ID()))
	}

	// Build versions array (all version strings)
	// Python reference: get_api_details() - "versions": [version.version for version in self.provider.get_all_versions()]
	versionStrings := make([]string, len(versions))
	for i, v := range versions {
		versionStrings[i] = v.Version()
	}

	// Get latest version for detailed fields
	// Versions are ordered DESC by ID (highest/latest first)
	var latestVersion *provider.ProviderVersion
	var tag *string
	var publishedAt *string
	if len(versions) > 0 {
		latestVersion = versions[0]
		tag = latestVersion.GitTag()
		if latestVersion.PublishedAt() != nil {
			formatted := latestVersion.PublishedAt().Format(time.RFC3339)
			publishedAt = &formatted
		}
	}

	// Build latest version string
	latestVersionString := ""
	if latestVersion != nil {
		latestVersionString = latestVersion.Version()
	}

	// Build docs array
	// Python reference: get_api_details() - "docs": [doc.get_api_outline() for doc in ProviderVersionDocumentation.get_by_provider_version(self)]
	docsArray := make([]Doc, 0, len(docs))
	for _, doc := range docs {
		docsArray = append(docsArray, Doc{
			Name:     doc.Name(),
			Slug:     doc.Slug(),
			Title:    doc.Title(),
			Category: doc.Subcategory(), // Python uses subcategory as the category in API
		})
	}

	return ProviderDetailResponse{
		ID:          fmt.Sprintf("%d", p.ID()),
		Owner:       p.Owner(), // From repository owner
		Namespace:   string(ns.Name()),
		Name:        p.Name(),
		Alias:       nil, // Always null in Python
		Version:     latestVersionString,
		Tag:         tag,
		Description: p.Description(),
		Source:      p.SourceURL(), // From repository clone URL (with .git removed)
		PublishedAt: publishedAt,
		Downloads:   totalDownloads, // From analytics
		Tier:        p.Tier(),
		LogoURL:     p.LogoURL(), // From repository logo URL
		Versions:    versionStrings, // CRITICAL - enables frontend version dropdown
		Docs:        docsArray, // From provider version documentation
	}
}

// NewProviderVersionsResponse creates a versions response from domain models
func NewProviderVersionsResponse(namespace, providerName string, versions []*provider.ProviderVersion) ProviderVersionsResponse {
	versionData := make([]ProviderVersion, len(versions))
	for i, v := range versions {
		versionData[i] = ProviderVersion{
			Version:   v.Version(),
			Protocols: v.ProtocolVersions(),
		}
	}

	return ProviderVersionsResponse{
		ID:       fmt.Sprintf("%s/%s", namespace, providerName),
		Versions: versionData,
	}
}
