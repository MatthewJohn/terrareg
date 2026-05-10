//go:build !integration && !selenium

package provider_source_test

// Python reference: /app/test/unit/terrareg/test_provider_source_hierarchy.py

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modulemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	providersourcemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	providersourceservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// MockProviderSourceFactory is a mock implementation of ProviderSourceFactory for testing
type MockProviderSourceFactory struct {
	providerSources map[string]*providersourcemodel.ProviderSource
}

// NewMockProviderSourceFactory creates a new mock factory
func NewMockProviderSourceFactory() *MockProviderSourceFactory {
	return &MockProviderSourceFactory{
		providerSources: make(map[string]*providersourcemodel.ProviderSource),
	}
}

// AddProviderSource adds a provider source to the mock factory
func (m *MockProviderSourceFactory) AddProviderSource(ps *providersourcemodel.ProviderSource) {
	m.providerSources[ps.Name()] = ps
}

// GetProviderSourceByName returns a provider source by name
// Returns a ProviderSourceInstance interface wrapper for testing
func (m *MockProviderSourceFactory) GetProviderSourceByName(_ context.Context, name string) (providersourceservice.ProviderSourceInstance, error) {
	ps, ok := m.providerSources[name]
	if !ok {
		return nil, nil
	}
	return &mockProviderSourceInstance{source: ps}, nil
}

// mockProviderSourceInstance adapts ProviderSource model to ProviderSourceInstance interface for testing
type mockProviderSourceInstance struct {
	source *providersourcemodel.ProviderSource
}

func (m *mockProviderSourceInstance) Name() string {
	return m.source.Name()
}

func (m *mockProviderSourceInstance) ApiName(ctx context.Context) (string, error) {
	return m.source.ApiName(), nil
}

func (m *mockProviderSourceInstance) Type() providersourcemodel.ProviderSourceType {
	return m.source.Type()
}

func (m *mockProviderSourceInstance) GetLoginRedirectURL(ctx context.Context) (string, error) {
	return "", nil
}

func (m *mockProviderSourceInstance) GetUserAccessToken(ctx context.Context, code string) (string, error) {
	return "", nil
}

func (m *mockProviderSourceInstance) GetUsername(ctx context.Context, accessToken string) (string, error) {
	return "", nil
}

func (m *mockProviderSourceInstance) GetUserOrganizations(ctx context.Context, accessToken string) []string {
	return []string{}
}

func (m *mockProviderSourceInstance) GetUserOrganizationsList(ctx context.Context, sessionID string) ([]*providersourcemodel.Organization, error) {
	return []*providersourcemodel.Organization{}, nil
}

func (m *mockProviderSourceInstance) GetUserRepositories(ctx context.Context, sessionID string) ([]*providersourcemodel.Repository, error) {
	return []*providersourcemodel.Repository{}, nil
}

func (m *mockProviderSourceInstance) RefreshNamespaceRepositories(ctx context.Context, namespace string) error {
	return nil
}

func (m *mockProviderSourceInstance) PublishProviderFromRepository(ctx context.Context, repoID int, categoryID int, namespace string) (*providersourceservice.PublishProviderResult, error) {
	return nil, nil
}

func (m *mockProviderSourceInstance) GetReleaseArtifact(ctx context.Context, repo *sqldb.RepositoryDB, artifact *providersourcemodel.ReleaseArtifactMetadata, accessToken string) ([]byte, error) {
	return nil, nil
}

func (m *mockProviderSourceInstance) GetReleaseArchive(ctx context.Context, repo *sqldb.RepositoryDB, releaseMetadata *providersourcemodel.RepositoryReleaseMetadata, accessToken string) ([]byte, string, error) {
	return nil, "", nil
}

func (m *mockProviderSourceInstance) LoginButtonText(ctx context.Context) (string, error) {
	return "", nil
}

// TestNamespaceDefaultProviderSourcePropertyReturnsNoneWhenNotSet
// Python reference: test_namespace_default_provider_source_property_returns_none_when_not_set
func TestNamespaceDefaultProviderSourcePropertyReturnsNoneWhenNotSet(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-no-default",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	result, err := namespace.DefaultProviderSource(ctx)
	require.NoError(t, err)
	assert.Nil(t, result, "Expected nil when default_provider_source_name is not set")
}

// TestNamespaceDefaultProviderSourcePropertyReturnsProviderSourceWhenSet
// Python reference: test_namespace_default_provider_source_property_returns_provider_source_when_set
func TestNamespaceDefaultProviderSourcePropertyReturnsProviderSourceWhenSet(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	providerSourceName := "test-default-ps"
	providerSource := providersourcemodel.NewProviderSource(
		providerSourceName,
		"test-default-ps",
		providersourcemodel.ProviderSourceTypeGithub,
		&providersourcemodel.ProviderSourceConfig{
			BaseURL:        "https://github.example.com",
			ApiURL:         "https://api.github.example.com",
			ClientID:       "test-client-id",
			PrivateKeyPath: "test-key",
		},
	)
	factory.AddProviderSource(providerSource)

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-with-default",
		nil,
		modulemodel.NamespaceTypeGithubOrg,
		&providerSourceName,
		factory,
	)

	result, err := namespace.DefaultProviderSource(ctx)
	require.NoError(t, err)
	require.NotNil(t, result, "Expected provider source when default_provider_source_name is set")
	assert.Equal(t, providerSourceName, result.Name())
}

// TestModuleProviderGetEffectiveProviderSourceReturnsModuleProviderSource
// Python reference: test_module_provider_get_effective_provider_source_returns_module_provider_source
func TestModuleProviderGetEffectiveProviderSourceReturnsModuleProviderSource(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	providerSourceName := "test-module-ps"
	providerSource := providersourcemodel.NewProviderSource(
		providerSourceName,
		"test-module-ps",
		providersourcemodel.ProviderSourceTypeGithub,
		&providersourcemodel.ProviderSourceConfig{},
	)
	factory.AddProviderSource(providerSource)

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-module-priority",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		&providerSourceName,
		false,
		factory,
		time.Now(),
		time.Now(),
	)

	result, err := moduleProvider.GetEffectiveProviderSource(ctx)
	require.NoError(t, err)
	require.NotNil(t, result, "Expected provider source from module provider")
	assert.Equal(t, providerSourceName, result.Name())
}

// TestModuleProviderGetEffectiveProviderSourceFallsBackToNamespace
// Python reference: test_module_provider_get_effective_provider_source_falls_back_to_namespace
func TestModuleProviderGetEffectiveProviderSourceFallsBackToNamespace(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	providerSourceName := "test-namespace-fallback-ps"
	providerSource := providersourcemodel.NewProviderSource(
		providerSourceName,
		"test-namespace-fallback-ps",
		providersourcemodel.ProviderSourceTypeGithub,
		&providersourcemodel.ProviderSourceConfig{},
	)
	factory.AddProviderSource(providerSource)

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-namespace-fallback",
		nil,
		modulemodel.NamespaceTypeNone,
		&providerSourceName,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		nil, // No provider source at module level
		false,
		factory,
		time.Now(),
		time.Now(),
	)

	result, err := moduleProvider.GetEffectiveProviderSource(ctx)
	require.NoError(t, err)
	require.NotNil(t, result, "Expected provider source from namespace default")
	assert.Equal(t, providerSourceName, result.Name())
}

// TestModuleProviderGetEffectiveProviderSourceReturnsNoneWhenNoSource
// Python reference: test_module_provider_get_effective_provider_source_returns_none_when_no_source
func TestModuleProviderGetEffectiveProviderSourceReturnsNoneWhenNoSource(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-no-source",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		nil,
		false,
		factory,
		time.Now(),
		time.Now(),
	)

	result, err := moduleProvider.GetEffectiveProviderSource(ctx)
	require.NoError(t, err)
	assert.Nil(t, result, "Expected nil when no provider source is configured")
}

// TestModuleProviderGetEffectiveProviderSourceModulePriorityOverNamespace
// Python reference: test_module_provider_get_effective_provider_source_module_priority_over_namespace
func TestModuleProviderGetEffectiveProviderSourceModulePriorityOverNamespace(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	namespacePsName := "test-priority-namespace-ps"
	namespacePs := providersourcemodel.NewProviderSource(
		namespacePsName,
		"test-priority-namespace-ps",
		providersourcemodel.ProviderSourceTypeGithub,
		&providersourcemodel.ProviderSourceConfig{},
	)

	modulePsName := "test-priority-module-ps"
	modulePs := providersourcemodel.NewProviderSource(
		modulePsName,
		"test-priority-module-ps",
		providersourcemodel.ProviderSourceTypeGithub,
		&providersourcemodel.ProviderSourceConfig{},
	)
	factory.AddProviderSource(namespacePs)
	factory.AddProviderSource(modulePs)

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-priority-namespace",
		nil,
		modulemodel.NamespaceTypeNone,
		&namespacePsName,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		&modulePsName,
		false,
		factory,
		time.Now(),
		time.Now(),
	)

	result, err := moduleProvider.GetEffectiveProviderSource(ctx)
	require.NoError(t, err)
	require.NotNil(t, result, "Expected provider source")
	assert.Equal(t, modulePsName, result.Name(), "Should return module provider's source")
	assert.NotEqual(t, namespacePsName, result.Name(), "Should not return namespace's source")
}

// TestGetEffectiveProviderSourceWithInheritanceDisabled
// Python reference: test_get_effective_provider_source_with_inheritance_disabled
func TestGetEffectiveProviderSourceWithInheritanceDisabled(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	namespacePsName := "test-namespace-ps-effective"
	namespacePs := providersourcemodel.NewProviderSource(
		namespacePsName,
		"test-namespace-ps-effective",
		providersourcemodel.ProviderSourceTypeGithub,
		&providersourcemodel.ProviderSourceConfig{},
	)
	factory.AddProviderSource(namespacePs)

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-inherit-effective",
		nil,
		modulemodel.NamespaceTypeNone,
		&namespacePsName,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		nil,
		false, // Inheritance enabled
		factory,
		time.Now(),
		time.Now(),
	)

	// Without inheritance disabled, should get namespace default
	result, err := moduleProvider.GetEffectiveProviderSource(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, namespacePsName, result.Name())

	// Disable inheritance
	moduleProvider.SetProviderSourceInheritanceDisabled(true)

	// With inheritance disabled, should return nil
	result, err = moduleProvider.GetEffectiveProviderSource(ctx)
	require.NoError(t, err)
	assert.Nil(t, result, "Should return nil when inheritance is disabled and no module-level source")
}

// TestNamespaceUpdateDefaultProviderSourceSetValid
// Python reference: test_namespace_update_default_provider_source_set_valid
func TestNamespaceUpdateDefaultProviderSourceSetValid(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	providerSourceName := "test-ps-valid"
	providerSource := providersourcemodel.NewProviderSource(
		providerSourceName,
		"test-ps-valid",
		providersourcemodel.ProviderSourceTypeGithub,
		&providersourcemodel.ProviderSourceConfig{},
	)
	factory.AddProviderSource(providerSource)

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-update-valid",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	err := namespace.UpdateDefaultProviderSource(ctx, &providerSourceName)
	require.NoError(t, err)

	result, err := namespace.DefaultProviderSource(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, providerSourceName, result.Name())
}

// TestNamespaceUpdateDefaultProviderSourceUnsetEmptyString
// Python reference: test_namespace_update_default_provider_source_unset_empty_string
func TestNamespaceUpdateDefaultProviderSourceUnsetEmptyString(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	providerSourceName := "test-ps-unset"
	providerSource := providersourcemodel.NewProviderSource(
		providerSourceName,
		"test-ps-unset",
		providersourcemodel.ProviderSourceTypeGithub,
		&providersourcemodel.ProviderSourceConfig{},
	)
	factory.AddProviderSource(providerSource)

	// Set provider source first
	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-unset",
		nil,
		modulemodel.NamespaceTypeNone,
		&providerSourceName,
		factory,
	)

	// Verify it's set
	result, _ := namespace.DefaultProviderSource(ctx)
	assert.NotNil(t, result)

	// Unset with empty string
	emptyStr := ""
	err := namespace.UpdateDefaultProviderSource(ctx, &emptyStr)
	require.NoError(t, err)

	// Verify it was unset
	result, err = namespace.DefaultProviderSource(ctx)
	require.NoError(t, err)
	assert.Nil(t, result)
}

// TestNamespaceUpdateDefaultProviderSourceInvalidProvider
// Python reference: test_namespace_update_default_provider_source_invalid_provider
func TestNamespaceUpdateDefaultProviderSourceInvalidProvider(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-invalid",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	// Try to set invalid provider source
	invalidName := "nonexistent-provider"
	err := namespace.UpdateDefaultProviderSource(ctx, &invalidName)

	// Should return InvalidProviderSourceNameError
	require.Error(t, err)
	var invalidErr *modulemodel.InvalidProviderSourceNameError
	require.ErrorAs(t, err, &invalidErr)
	assert.Equal(t, invalidName, invalidErr.Name)
}

// TestNamespaceUpdateDefaultProviderSourceNoChange
// Python reference: test_namespace_update_default_provider_source_no_change
func TestNamespaceUpdateDefaultProviderSourceNoChange(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-no-change",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	// Initially nil
	result, _ := namespace.DefaultProviderSource(ctx)
	assert.Nil(t, result)

	// Call with nil - should not raise and should not change
	err := namespace.UpdateDefaultProviderSource(ctx, nil)
	require.NoError(t, err)

	// Verify still nil
	result, err = namespace.DefaultProviderSource(ctx)
	require.NoError(t, err)
	assert.Nil(t, result)
}

// TestModuleProviderUpdateProviderSourceSetValid
// Python reference: test_module_provider_update_provider_source_set_valid
func TestModuleProviderUpdateProviderSourceSetValid(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	providerSourceName := "test-mp-ps-valid"
	providerSource := providersourcemodel.NewProviderSource(
		providerSourceName,
		"test-mp-ps-valid",
		providersourcemodel.ProviderSourceTypeGithub,
		&providersourcemodel.ProviderSourceConfig{},
	)
	factory.AddProviderSource(providerSource)

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-mp-valid",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		nil,
		false,
		factory,
		time.Now(),
		time.Now(),
	)

	err := moduleProvider.UpdateProviderSource(ctx, &providerSourceName)
	require.NoError(t, err)

	result, err := moduleProvider.ProviderSource(ctx)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, providerSourceName, result.Name())
}

// TestModuleProviderUpdateProviderSourceUnsetEmptyString
// Python reference: test_module_provider_update_provider_source_unset_empty_string
func TestModuleProviderUpdateProviderSourceUnsetEmptyString(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	providerSourceName := "test-mp-ps-unset"
	providerSource := providersourcemodel.NewProviderSource(
		providerSourceName,
		"test-mp-ps-unset",
		providersourcemodel.ProviderSourceTypeGithub,
		&providersourcemodel.ProviderSourceConfig{},
	)
	factory.AddProviderSource(providerSource)

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-mp-unset",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		&providerSourceName,
		false,
		factory,
		time.Now(),
		time.Now(),
	)

	// Verify it's set
	result, _ := moduleProvider.ProviderSource(ctx)
	assert.NotNil(t, result)

	// Unset with empty string
	emptyStr := ""
	err := moduleProvider.UpdateProviderSource(ctx, &emptyStr)
	require.NoError(t, err)

	// Verify it was unset
	result, err = moduleProvider.ProviderSource(ctx)
	require.NoError(t, err)
	assert.Nil(t, result)
}

// TestModuleProviderUpdateProviderSourceInvalidProvider
// Python reference: test_module_provider_update_provider_source_invalid_provider
func TestModuleProviderUpdateProviderSourceInvalidProvider(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-mp-invalid",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		nil,
		false,
		factory,
		time.Now(),
		time.Now(),
	)

	// Try to set invalid provider source
	invalidName := "nonexistent-provider"
	err := moduleProvider.UpdateProviderSource(ctx, &invalidName)

	// Should return InvalidProviderSourceNameError
	require.Error(t, err)
	var invalidErr *modulemodel.InvalidProviderSourceNameError
	require.ErrorAs(t, err, &invalidErr)
	assert.Equal(t, invalidName, invalidErr.Name)
}

// TestModuleProviderUpdateInheritanceDisabledEnable
// Python reference: test_module_provider_update_inheritance_disabled_enable
func TestModuleProviderUpdateInheritanceDisabledEnable(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-inherit-enable",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		nil,
		true, // Set to disabled first
		factory,
		time.Now(),
		time.Now(),
	)

	// Verify it's disabled
	assert.True(t, moduleProvider.ProviderSourceInheritanceDisabled())

	// Enable inheritance (set to false)
	falseVal := false
	err := moduleProvider.UpdateProviderSourceInheritanceDisabled(ctx, &falseVal)
	require.NoError(t, err)

	// Verify it's enabled
	assert.False(t, moduleProvider.ProviderSourceInheritanceDisabled())
}

// TestModuleProviderUpdateInheritanceDisabledDisable
// Python reference: test_module_provider_update_inheritance_disabled_disable
func TestModuleProviderUpdateInheritanceDisabledDisable(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-inherit-disable",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		nil,
		false,
		factory,
		time.Now(),
		time.Now(),
	)

	// Verify it's enabled by default
	assert.False(t, moduleProvider.ProviderSourceInheritanceDisabled())

	// Disable inheritance
	trueVal := true
	err := moduleProvider.UpdateProviderSourceInheritanceDisabled(ctx, &trueVal)
	require.NoError(t, err)

	// Verify it's disabled
	assert.True(t, moduleProvider.ProviderSourceInheritanceDisabled())
}

// TestModuleProviderUpdateInheritanceDisabledNoChange
// Python reference: test_module_provider_update_inheritance_disabled_no_change
func TestModuleProviderUpdateInheritanceDisabledNoChange(t *testing.T) {
	ctx := context.Background()
	factory := NewMockProviderSourceFactory()

	namespace := modulemodel.ReconstructNamespace(
		1,
		"test-inherit-no-change",
		nil,
		modulemodel.NamespaceTypeNone,
		nil,
		factory,
	)

	moduleProvider := modulemodel.ReconstructModuleProvider(
		1,
		namespace,
		"test-module",
		"aws",
		false,
		nil, nil, nil, nil,
		nil, nil, false,
		nil,
		false,
		factory,
		time.Now(),
		time.Now(),
	)

	// Verify default is false
	assert.False(t, moduleProvider.ProviderSourceInheritanceDisabled())

	// Call with nil - should not change
	err := moduleProvider.UpdateProviderSourceInheritanceDisabled(ctx, nil)
	require.NoError(t, err)

	// Verify still false
	assert.False(t, moduleProvider.ProviderSourceInheritanceDisabled())
}
