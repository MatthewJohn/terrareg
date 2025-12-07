package presenter

import (
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockNamespace(name string) *model.Namespace {
	namespace, _ := model.NewNamespace(name, nil, "NONE")
	return namespace
}

func createMockModuleProvider(verified bool, versions []*model.ModuleVersion) *model.ModuleProvider {
	namespace := createMockNamespace("example")
	mp, _ := model.NewModuleProvider(namespace, "aws", "aws")

	if verified {
		mp.Verify()
	}

	// Add versions
	for _, v := range versions {
		mp.AddVersion(v)
	}

	return mp
}

func createMockModuleVersion(version string, owner, description *string, publishedAt *time.Time) *model.ModuleVersion {
	details := model.NewModuleDetails(nil)

	// Use ReconstructModuleVersion to create with owner and description
	mv, _ := model.ReconstructModuleVersion(
		1, // id
		version,
		details,
		false,              // beta
		false,              // internal
		publishedAt != nil, // published
		publishedAt,
		nil,   // gitSHA
		nil,   // gitPath
		false, // archiveGitPath
		nil,   // repoBaseURLTemplate
		nil,   // repoCloneURLTemplate
		nil,   // repoBrowseURLTemplate
		owner,
		description,
		nil,        // variableTemplate
		nil,        // extractionVersion
		time.Now(), // createdAt
		time.Now(), // updatedAt
	)

	return mv
}

func TestModulePresenter_ToDTO_ProviderBaseMapping(t *testing.T) {
	// Arrange
	publishedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	owner := "team-a"
	description := "Initial version"
	version := createMockModuleVersion("1.0.0", &owner, &description, &publishedAt)

	mp := createMockModuleProvider(true, []*model.ModuleVersion{version})
	presenter := NewModulePresenter()

	// Act
	result := presenter.ToDTO(mp)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, "example/aws/aws", result.ID)
	assert.Equal(t, "example", result.Namespace)
	assert.Equal(t, "aws", result.Name)
	assert.Equal(t, "aws", result.Provider)
	assert.True(t, result.Verified)
	assert.False(t, result.Trusted) // TODO: When namespace service is integrated
}

func TestModulePresenter_ToDTO_WithLatestVersion(t *testing.T) {
	// Arrange
	publishedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	owner := "team-a"
	description := "Initial version"
	version := createMockModuleVersion("1.0.0", &owner, &description, &publishedAt)

	mp := createMockModuleProvider(true, []*model.ModuleVersion{version})
	presenter := NewModulePresenter()

	// Act
	result := presenter.ToDTO(mp)

	// Assert
	require.NotNil(t, result)
	assert.NotNil(t, result.Owner)
	assert.Equal(t, "team-a", *result.Owner)
	assert.NotNil(t, result.Description)
	assert.Equal(t, "Initial version", *result.Description)
	assert.NotNil(t, result.PublishedAt)
	assert.Equal(t, "2023-01-01T12:00:00Z", *result.PublishedAt)
}

func TestModulePresenter_ToDTO_WithoutLatestVersion(t *testing.T) {
	// Arrange
	mp := createMockModuleProvider(true, []*model.ModuleVersion{}) // No versions
	presenter := NewModulePresenter()

	// Act
	result := presenter.ToDTO(mp)

	// Assert
	require.NotNil(t, result)
	assert.Nil(t, result.Owner)
	assert.Nil(t, result.Description)
	assert.Nil(t, result.PublishedAt)
}

func TestModulePresenter_ToListDTO_Conversion(t *testing.T) {
	// Arrange
	publishedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	ownerA := "team-a"
	descA := "Module A"
	versionA := createMockModuleVersion("1.0.0", &ownerA, &descA, &publishedAt)

	ownerB := "team-b"
	descB := "Module B"
	versionB := createMockModuleVersion("2.0.0", &ownerB, &descB, &publishedAt)

	modules := []*model.ModuleProvider{
		createMockModuleProvider(true, []*model.ModuleVersion{versionA}),
		createMockModuleProvider(true, []*model.ModuleVersion{versionB}),
	}
	presenter := NewModulePresenter()

	// Act
	result := presenter.ToListDTO(modules)

	// Assert
	require.NotNil(t, result)
	assert.Len(t, result.Modules, 2)
	assert.Equal(t, "example/aws/aws", result.Modules[0].ID)
	assert.Equal(t, "example/aws/aws", result.Modules[1].ID)
}

func TestModulePresenter_ToSearchDTO_WithPagination(t *testing.T) {
	// Arrange
	modules := []*model.ModuleProvider{
		createMockModuleProvider(true, []*model.ModuleVersion{}),
	}
	presenter := NewModulePresenter()

	// Act
	result := presenter.ToSearchDTO(modules, 100, 20, 0)

	// Assert
	require.NotNil(t, result)
	assert.Len(t, result.Modules, 1)
	assert.Equal(t, 20, result.Meta.Limit)
	assert.Equal(t, 0, result.Meta.Offset)
	assert.Equal(t, 100, result.Meta.TotalCount)
}

func TestModulePresenter_ToListDTO_Empty(t *testing.T) {
	// Arrange
	var modules []*model.ModuleProvider
	presenter := NewModulePresenter()

	// Act
	result := presenter.ToListDTO(modules)

	// Assert
	require.NotNil(t, result)
	assert.Empty(t, result.Modules)
}

func TestModulePresenter_ToSearchDTO_Empty(t *testing.T) {
	// Arrange
	var modules []*model.ModuleProvider
	presenter := NewModulePresenter()

	// Act
	result := presenter.ToSearchDTO(modules, 0, 10, 0)

	// Assert
	require.NotNil(t, result)
	assert.Empty(t, result.Modules)
	assert.Equal(t, 10, result.Meta.Limit)
	assert.Equal(t, 0, result.Meta.Offset)
	assert.Equal(t, 0, result.Meta.TotalCount)
}

// ModuleVersionPresenter tests

func TestModuleVersionPresenter_ToDTO_VersionBaseMapping(t *testing.T) {
	// Arrange
	presenter := NewModuleVersionPresenter()
	mv := createMockModuleVersion("1.0.0", nil, nil, nil)
	namespace := "example"
	moduleName := "aws"
	provider := "aws"

	// Act
	result := presenter.ToDTO(mv, namespace, moduleName, provider)

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, "example/aws/aws/1.0.0", result.ID)
	assert.Equal(t, namespace, result.Namespace)
	assert.Equal(t, moduleName, result.Name)
	assert.Equal(t, provider, result.Provider)
	assert.Equal(t, "1.0.0", result.Version)
	assert.False(t, result.Verified) // TODO: Get from module provider
	assert.False(t, result.Trusted)  // TODO: Get from namespace service
	assert.False(t, result.Internal)
}

func TestModuleVersionPresenter_ToDTO_WithOwner(t *testing.T) {
	// Arrange
	presenter := NewModuleVersionPresenter()
	owner := "team-a"
	mv := createMockModuleVersion("1.0.0", &owner, nil, nil)

	// Act
	result := presenter.ToDTO(mv, "example", "aws", "aws")

	// Assert
	require.NotNil(t, result)
	assert.NotNil(t, result.Owner)
	assert.Equal(t, owner, *result.Owner)
}

func TestModuleVersionPresenter_ToDTO_WithDescription(t *testing.T) {
	// Arrange
	presenter := NewModuleVersionPresenter()
	description := "Test module version"
	mv := createMockModuleVersion("1.0.0", nil, &description, nil)

	// Act
	result := presenter.ToDTO(mv, "example", "aws", "aws")

	// Assert
	require.NotNil(t, result)
	assert.NotNil(t, result.Description)
	assert.Equal(t, description, *result.Description)
}

func TestModuleVersionPresenter_ToDTO_WithPublishedAt(t *testing.T) {
	// Arrange
	presenter := NewModuleVersionPresenter()
	publishedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	mv := createMockModuleVersion("1.0.0", nil, nil, &publishedAt)

	// Act
	result := presenter.ToDTO(mv, "example", "aws", "aws")

	// Assert
	require.NotNil(t, result)
	assert.NotNil(t, result.PublishedAt)
	assert.Equal(t, "2023-01-01T12:00:00Z", *result.PublishedAt)
}

func TestModuleVersionPresenter_ToDTO_WithAllFields(t *testing.T) {
	// Arrange
	presenter := NewModuleVersionPresenter()
	owner := "team-a"
	description := "Complete module version"
	publishedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	mv := createMockModuleVersion("2.1.0", &owner, &description, &publishedAt)

	// Act
	result := presenter.ToDTO(mv, "example", "aws", "aws")

	// Assert
	require.NotNil(t, result)
	assert.Equal(t, "example/aws/aws/2.1.0", result.ID)
	assert.Equal(t, "example", result.Namespace)
	assert.Equal(t, "aws", result.Name)
	assert.Equal(t, "aws", result.Provider)
	assert.Equal(t, "2.1.0", result.Version)
	assert.NotNil(t, result.Owner)
	assert.Equal(t, owner, *result.Owner)
	assert.NotNil(t, result.Description)
	assert.Equal(t, description, *result.Description)
	assert.NotNil(t, result.PublishedAt)
	assert.Equal(t, "2023-01-01T12:00:00Z", *result.PublishedAt)
	assert.False(t, result.Internal)
}

func TestModuleVersionPresenter_ToDTO_InternalVersion(t *testing.T) {
	// Arrange
	presenter := NewModuleVersionPresenter()
	mv := createMockModuleVersionWithInternal("1.0.0-beta", nil, nil, nil, true)

	// Act
	result := presenter.ToDTO(mv, "example", "aws", "aws")

	// Assert
	require.NotNil(t, result)
	assert.True(t, result.Internal)
}

func TestModuleVersionPresenter_ToDTO_WithoutOptionalFields(t *testing.T) {
	// Arrange
	presenter := NewModuleVersionPresenter()
	mv := createMockModuleVersion("1.0.0", nil, nil, nil)

	// Act
	result := presenter.ToDTO(mv, "example", "aws", "aws")

	// Assert
	require.NotNil(t, result)
	assert.Nil(t, result.Owner)
	assert.Nil(t, result.Description)
	assert.Nil(t, result.PublishedAt)
}

// Helper function for creating internal module versions
func createMockModuleVersionWithInternal(version string, owner, description *string, publishedAt *time.Time, internal bool) *model.ModuleVersion {
	details := model.NewModuleDetails(nil)

	// Use ReconstructModuleVersion to create with owner and description
	mv, _ := model.ReconstructModuleVersion(
		1, // id
		version,
		details,
		false,              // beta
		internal,           // internal
		publishedAt != nil, // published
		publishedAt,
		nil,   // gitSHA
		nil,   // gitPath
		false, // archiveGitPath
		nil,   // repoBaseURLTemplate
		nil,   // repoCloneURLTemplate
		nil,   // repoBrowseURLTemplate
		owner,
		description,
		nil,        // variableTemplate
		nil,        // extractionVersion
		time.Now(), // createdAt
		time.Now(), // updatedAt
	)

	return mv
}
