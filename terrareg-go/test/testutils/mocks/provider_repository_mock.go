package mocks

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	"github.com/stretchr/testify/mock"
)

// MockProviderRepository is a mock for ProviderRepository
type MockProviderRepository struct {
	mock.Mock
}

// Ensure MockProviderRepository implements the interface at compile time
var _ repository.ProviderRepository = (*MockProviderRepository)(nil)

// FindByNamespaceAndName mocks finding a provider by namespace and name
func (m *MockProviderRepository) FindByNamespaceAndName(ctx context.Context, namespace, providerName string) (*provider.Provider, error) {
	args := m.Called(ctx, namespace, providerName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Provider), args.Error(1)
}

// Save mocks saving a provider
func (m *MockProviderRepository) Save(ctx context.Context, prov *provider.Provider) error {
	args := m.Called(ctx, prov)
	return args.Error(0)
}

// FindByID mocks finding a provider by ID
func (m *MockProviderRepository) FindByID(ctx context.Context, providerID int) (*provider.Provider, error) {
	args := m.Called(ctx, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Provider), args.Error(1)
}

// FindVersionsByProvider mocks finding versions for a provider
func (m *MockProviderRepository) FindVersionsByProvider(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error) {
	args := m.Called(ctx, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*provider.ProviderVersion), args.Error(1)
}

// FindVersionByProviderAndVersion mocks finding a specific version
func (m *MockProviderRepository) FindVersionByProviderAndVersion(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error) {
	args := m.Called(ctx, providerID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.ProviderVersion), args.Error(1)
}

// FindAll mocks finding all providers
func (m *MockProviderRepository) FindAll(ctx context.Context, offset, limit int) ([]*provider.Provider, map[int]string, map[int]repository.VersionData, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(map[int]string), args.Get(2).(map[int]repository.VersionData), args.Int(3), args.Error(4)
	}
	return args.Get(0).([]*provider.Provider), args.Get(1).(map[int]string), args.Get(2).(map[int]repository.VersionData), args.Int(3), args.Error(4)
}

// Search mocks searching for providers
func (m *MockProviderRepository) Search(ctx context.Context, query repository.ProviderSearchQuery) (*repository.ProviderSearchResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ProviderSearchResult), args.Error(1)
}

// GetSearchFilters mocks getting search filters
func (m *MockProviderRepository) GetSearchFilters(ctx context.Context, query string, trustedNamespaces []string) (*repository.ProviderSearchFilters, error) {
	args := m.Called(ctx, query, trustedNamespaces)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ProviderSearchFilters), args.Error(1)
}

// FindVersionByID mocks finding a version by ID
func (m *MockProviderRepository) FindVersionByID(ctx context.Context, versionID int) (*provider.ProviderVersion, error) {
	args := m.Called(ctx, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.ProviderVersion), args.Error(1)
}

// FindBinariesByVersion mocks finding binaries for a version
func (m *MockProviderRepository) FindBinariesByVersion(ctx context.Context, versionID int) ([]*provider.ProviderBinary, error) {
	args := m.Called(ctx, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*provider.ProviderBinary), args.Error(1)
}

// FindBinaryByPlatform mocks finding a binary by platform
func (m *MockProviderRepository) FindBinaryByPlatform(ctx context.Context, versionID int, os, arch string) (*provider.ProviderBinary, error) {
	args := m.Called(ctx, versionID, os, arch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.ProviderBinary), args.Error(1)
}

// FindDocumentationByID mocks finding documentation by ID
func (m *MockProviderRepository) FindDocumentationByID(ctx context.Context, id int) (*provider.ProviderVersionDocumentation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.ProviderVersionDocumentation), args.Error(1)
}

// FindDocumentationByVersion mocks finding documentation for a version
func (m *MockProviderRepository) FindDocumentationByVersion(ctx context.Context, versionID int) ([]*provider.ProviderVersionDocumentation, error) {
	args := m.Called(ctx, versionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*provider.ProviderVersionDocumentation), args.Error(1)
}

// FindDocumentationByTypeSlugAndLanguage mocks finding documentation by type, slug, and language
func (m *MockProviderRepository) FindDocumentationByTypeSlugAndLanguage(ctx context.Context, versionID int, docType, slug, language string) (*provider.ProviderVersionDocumentation, error) {
	args := m.Called(ctx, versionID, docType, slug, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.ProviderVersionDocumentation), args.Error(1)
}

// SearchDocumentation mocks searching for documentation
func (m *MockProviderRepository) SearchDocumentation(ctx context.Context, versionID int, category, slug, language string) ([]*provider.ProviderVersionDocumentation, error) {
	args := m.Called(ctx, versionID, category, slug, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*provider.ProviderVersionDocumentation), args.Error(1)
}

// SaveDocumentation mocks saving documentation
func (m *MockProviderRepository) SaveDocumentation(ctx context.Context, documentation *provider.ProviderVersionDocumentation) error {
	args := m.Called(ctx, documentation)
	return args.Error(0)
}

// FindGPGKeysByProvider mocks finding GPG keys for a provider
func (m *MockProviderRepository) FindGPGKeysByProvider(ctx context.Context, providerID int) ([]*provider.GPGKey, error) {
	args := m.Called(ctx, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*provider.GPGKey), args.Error(1)
}

// FindGPGKeyByKeyID mocks finding a GPG key by key ID
func (m *MockProviderRepository) FindGPGKeyByKeyID(ctx context.Context, keyID string) (*provider.GPGKey, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.GPGKey), args.Error(1)
}

// SaveVersion mocks saving a provider version
func (m *MockProviderRepository) SaveVersion(ctx context.Context, version *provider.ProviderVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

// SaveBinary mocks saving a provider binary
func (m *MockProviderRepository) SaveBinary(ctx context.Context, binary *provider.ProviderBinary) error {
	args := m.Called(ctx, binary)
	return args.Error(0)
}

// SaveGPGKey mocks saving a GPG key
func (m *MockProviderRepository) SaveGPGKey(ctx context.Context, gpgKey *provider.GPGKey) error {
	args := m.Called(ctx, gpgKey)
	return args.Error(0)
}

// DeleteVersion mocks deleting a provider version
func (m *MockProviderRepository) DeleteVersion(ctx context.Context, versionID int) error {
	args := m.Called(ctx, versionID)
	return args.Error(0)
}

// DeleteBinary mocks deleting a provider binary
func (m *MockProviderRepository) DeleteBinary(ctx context.Context, binaryID int) error {
	args := m.Called(ctx, binaryID)
	return args.Error(0)
}

// DeleteGPGKey mocks deleting a GPG key
func (m *MockProviderRepository) DeleteGPGKey(ctx context.Context, keyID int) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

// SetLatestVersion mocks setting the latest version for a provider
func (m *MockProviderRepository) SetLatestVersion(ctx context.Context, providerID, versionID int) error {
	args := m.Called(ctx, providerID, versionID)
	return args.Error(0)
}

// GetProviderVersionCount mocks getting the version count for a provider
func (m *MockProviderRepository) GetProviderVersionCount(ctx context.Context, providerID int) (int, error) {
	args := m.Called(ctx, providerID)
	return args.Int(0), args.Error(1)
}

// GetBinaryDownloadCount mocks getting the download count for a binary
func (m *MockProviderRepository) GetBinaryDownloadCount(ctx context.Context, binaryID int) (int64, error) {
	args := m.Called(ctx, binaryID)
	return args.Get(0).(int64), args.Error(1)
}
