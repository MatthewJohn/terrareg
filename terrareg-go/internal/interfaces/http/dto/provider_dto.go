package dto

import (
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// ProviderListResponse represents the provider list API response
type ProviderListResponse struct {
	Meta      PaginationMeta `json:"meta"`
	Providers []ProviderData `json:"providers"`
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
type ProviderDetailResponse struct {
	ID          string   `json:"id"`
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Tier        string   `json:"tier"`
	Source      *string  `json:"source,omitempty"`
	Versions    []string `json:"versions,omitempty"`
}

// ProviderVersionsResponse represents the versions list for a provider
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

// NewProviderListResponse creates a provider list response from domain models
// Matches Python structure: returns version outline, not provider info
func NewProviderListResponse(providers []*provider.Provider, namespaceNames map[int]string, versionDataMap map[int]providerRepo.VersionData, total, offset, limit int) ProviderListResponse {
	providerData := make([]ProviderData, 0, len(providers))
	for _, p := range providers {
		namespace := ""
		if ns, ok := namespaceNames[p.ID()]; ok {
			namespace = ns
		}

		// Get version data for this provider
		versionData, hasVersion := versionDataMap[p.ID()]

		// Skip providers without version data (matching Python behavior)
		// Python: "if module_provider.get_latest_version()"
		if !hasVersion {
			continue
		}

		// Build response matching Python's get_api_outline()
		// ID is version ID, not provider ID
		data := ProviderData{
			ID:          fmt.Sprintf("%d", versionData.VersionID),
			Owner:       derefString(versionData.RepositoryOwner),
			Namespace:   namespace,
			Name:        p.Name(),
			Alias:       nil, // Always null in Python
			Version:     versionData.Version,
			Tag:         versionData.GitTag,
			Description: versionData.RepositoryDescription, // Repository description, not provider description
			Source:      versionData.RepositoryCloneURL,
			PublishedAt: versionData.PublishedAt,
			Downloads:   0, // TODO: Calculate total downloads from binaries
			Tier:        p.Tier(),
			LogoURL:     nil, // TODO: Add logo URL to domain model
		}
		providerData = append(providerData, data)
	}

	return ProviderListResponse{
		Meta: PaginationMeta{
			Limit:      limit,
			Offset:     offset,
			TotalCount: total,
		},
		Providers: providerData,
	}
}

// derefString safely dereferences a string pointer
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// NewProviderDetailResponse creates a provider detail response from domain model
func NewProviderDetailResponse(p *provider.Provider) ProviderDetailResponse {
	return ProviderDetailResponse{
		ID:          fmt.Sprintf("%d", p.ID()),
		Namespace:   fmt.Sprintf("namespace-%d", p.NamespaceID()), // Placeholder
		Name:        p.Name(),
		Description: p.Description(),
		Tier:        p.Tier(),
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
