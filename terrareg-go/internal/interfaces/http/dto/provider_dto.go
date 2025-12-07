package dto

import (
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
)

// ProviderListResponse represents the provider list API response
type ProviderListResponse struct {
	Meta      PaginationMeta `json:"meta"`
	Providers []ProviderData `json:"providers"`
}

// ProviderData represents a single provider in list responses
type ProviderData struct {
	ID          string  `json:"id"`
	Namespace   string  `json:"namespace"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Tier        string  `json:"tier"`
	Alias       *string `json:"alias,omitempty"`
	Source      *string `json:"source,omitempty"`
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
func NewProviderListResponse(providers []*provider.Provider, total, offset, limit int) ProviderListResponse {
	providerData := make([]ProviderData, len(providers))
	for i, p := range providers {
		providerData[i] = ProviderData{
			ID:          fmt.Sprintf("%d", p.ID()),
			Namespace:   fmt.Sprintf("namespace-%d", p.NamespaceID()), // Placeholder - need to join namespace
			Name:        p.Name(),
			Description: p.Description(),
			Tier:        p.Tier(),
		}
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
