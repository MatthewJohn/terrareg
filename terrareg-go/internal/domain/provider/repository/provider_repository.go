package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
)

// ProviderSearchQuery represents search criteria for providers
type ProviderSearchQuery struct {
	Query             string
	Namespaces        []string
	Categories        []string
	TrustedNamespaces *bool // Filter for trusted namespaces only
	Contributed       *bool // Filter for contributed providers only
	Limit             int
	Offset            int
}

// ProviderSearchResult represents search results
type ProviderSearchResult struct {
	Providers       []*provider.Provider
	TotalCount      int
	NamespaceNames  map[int]string     // Maps provider ID to namespace name
	VersionData     map[int]VersionData // Maps provider ID to version data
}

// VersionData holds version information for a provider
type VersionData struct {
	VersionID             int
	Version               string
	GitTag                *string
	PublishedAt           *string
	RepositoryOwner       *string
	RepositoryDescription *string
	RepositoryCloneURL    *string
	RepositoryLogoURL     *string
	Downloads             int64
}

// ProviderSearchFilters represents available filter counts for a search query
type ProviderSearchFilters struct {
	TrustedNamespaces  int
	Contributed        int
	ProviderCategories map[string]int
	Namespaces         map[string]int
}

// ProviderRepository defines the interface for provider persistence
type ProviderRepository interface {
	// FindAll retrieves all providers
	FindAll(ctx context.Context, offset, limit int) ([]*provider.Provider, int, error)

	// Search searches for providers by query with filters
	Search(ctx context.Context, query ProviderSearchQuery) (*ProviderSearchResult, error)

	// GetSearchFilters gets available filters and counts for a search query
	GetSearchFilters(ctx context.Context, query string, trustedNamespaces []string) (*ProviderSearchFilters, error)

	// FindByNamespaceAndName retrieves a provider by namespace and name
	FindByNamespaceAndName(ctx context.Context, namespace, providerName string) (*provider.Provider, error)

	// FindByID retrieves a provider by its ID
	FindByID(ctx context.Context, providerID int) (*provider.Provider, error)

	// FindVersionsByProvider retrieves all versions for a provider
	FindVersionsByProvider(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error)

	// FindVersionByProviderAndVersion retrieves a specific version
	FindVersionByProviderAndVersion(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error)

	// FindVersionByID retrieves a version by its ID
	FindVersionByID(ctx context.Context, versionID int) (*provider.ProviderVersion, error)

	// FindBinariesByVersion retrieves all binaries for a provider version
	FindBinariesByVersion(ctx context.Context, versionID int) ([]*provider.ProviderBinary, error)

	// FindBinaryByPlatform retrieves a specific binary for a platform
	FindBinaryByPlatform(ctx context.Context, versionID int, os, arch string) (*provider.ProviderBinary, error)

	// FindDocumentationByID retrieves documentation by its ID
	FindDocumentationByID(ctx context.Context, id int) (*provider.ProviderVersionDocumentation, error)

	// FindDocumentationByVersion retrieves all documentation for a provider version
	FindDocumentationByVersion(ctx context.Context, versionID int) ([]*provider.ProviderVersionDocumentation, error)

	// FindDocumentationByTypeSlugAndLanguage retrieves documentation by type, slug, and language
	FindDocumentationByTypeSlugAndLanguage(ctx context.Context, versionID int, docType, slug, language string) (*provider.ProviderVersionDocumentation, error)

	// SearchDocumentation searches for documentation by category, slug, and language
	SearchDocumentation(ctx context.Context, versionID int, category, slug, language string) ([]*provider.ProviderVersionDocumentation, error)

	// SaveDocumentation persists provider documentation
	SaveDocumentation(ctx context.Context, documentation *provider.ProviderVersionDocumentation) error

	// FindGPGKeysByProvider retrieves all GPG keys for a provider
	FindGPGKeysByProvider(ctx context.Context, providerID int) ([]*provider.GPGKey, error)

	// FindGPGKeyByKeyID retrieves a GPG key by its key identifier
	FindGPGKeyByKeyID(ctx context.Context, keyID string) (*provider.GPGKey, error)

	// Save persists a provider aggregate to the database
	Save(ctx context.Context, provider *provider.Provider) error

	// SaveVersion persists a provider version
	SaveVersion(ctx context.Context, version *provider.ProviderVersion) error

	// SaveBinary persists a provider binary
	SaveBinary(ctx context.Context, binary *provider.ProviderBinary) error

	// SaveGPGKey persists a GPG key
	SaveGPGKey(ctx context.Context, gpgKey *provider.GPGKey) error

	// DeleteVersion removes a provider version
	DeleteVersion(ctx context.Context, versionID int) error

	// DeleteBinary removes a provider binary
	DeleteBinary(ctx context.Context, binaryID int) error

	// DeleteGPGKey removes a GPG key
	DeleteGPGKey(ctx context.Context, keyID int) error

	// SetLatestVersion updates the latest version for a provider
	SetLatestVersion(ctx context.Context, providerID, versionID int) error

	// GetProviderVersionCount returns the number of versions for a provider
	GetProviderVersionCount(ctx context.Context, providerID int) (int, error)

	// GetBinaryDownloadCount returns the download count for a binary
	GetBinaryDownloadCount(ctx context.Context, binaryID int) (int64, error)
}
