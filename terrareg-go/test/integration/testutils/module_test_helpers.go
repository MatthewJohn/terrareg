package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// SetupModuleVersion creates a complete module hierarchy (namespace, module, provider, version)
// Returns the database models for all created entities
func SetupModuleVersion(t *testing.T, db *sqldb.Database, namespace, moduleName, provider, version string) (*sqldb.NamespaceDB, *sqldb.ModuleProviderDB, *sqldb.ModuleVersionDB) {
	t.Helper()

	// Create namespace
	ns := CreateNamespace(t, db, namespace)

	// Create module provider
	modProvider := CreateModuleProvider(t, db, ns.ID, moduleName, provider)

	// Create module version
	modVersion := CreateModuleVersion(t, db, modProvider.ID, version)

	return &ns, &modProvider, &modVersion
}

// SetupPublishedModuleVersion creates a complete module hierarchy with a published version
func SetupPublishedModuleVersion(t *testing.T, db *sqldb.Database, namespace, moduleName, provider, version string) (*sqldb.NamespaceDB, *sqldb.ModuleProviderDB, *sqldb.ModuleVersionDB) {
	t.Helper()

	// Create namespace
	ns := CreateNamespace(t, db, namespace)

	// Create module provider
	modProvider := CreateModuleProvider(t, db, ns.ID, moduleName, provider)

	// Create published module version
	published := true
	modVersion := sqldb.ModuleVersionDB{
		ModuleProviderID:      modProvider.ID,
		Version:               version,
		Beta:                  false,
		Internal:              false,
		Published:             &published,
		PublishedAt:           nil,
		GitSHA:                nil,
		GitPath:               nil,
		ArchiveGitPath:        false,
		RepoBaseURLTemplate:   nil,
		RepoCloneURLTemplate:  nil,
		RepoBrowseURLTemplate: nil,
		Owner:                 nil,
		Description:           nil,
		VariableTemplate:      nil,
		ExtractionVersion:     nil,
		ModuleDetailsID:       nil,
	}

	err := db.DB.Create(&modVersion).Error
	require.NoError(t, err)

	return &ns, &modProvider, &modVersion
}

// SetupModuleVersionWithDetails creates a module version with module details
func SetupModuleVersionWithDetails(t *testing.T, db *sqldb.Database, namespace, moduleName, provider, version string, readmeContent string) (*sqldb.NamespaceDB, *sqldb.ModuleProviderDB, *sqldb.ModuleVersionDB, *sqldb.ModuleDetailsDB) {
	t.Helper()

	// Create module details
	details := CreateModuleDetails(t, db, readmeContent)

	// Create module hierarchy
	ns, modProvider, modVersion := SetupModuleVersion(t, db, namespace, moduleName, provider, version)

	// Link version to details
	modVersion.ModuleDetailsID = &details.ID
	err := db.DB.Save(&modVersion).Error
	require.NoError(t, err)

	return ns, modProvider, modVersion, &details
}

// SetupMultipleVersions creates a module provider with multiple versions
func SetupMultipleVersions(t *testing.T, db *sqldb.Database, namespace, moduleName, provider string, versions []string) (*sqldb.NamespaceDB, *sqldb.ModuleProviderDB, []*sqldb.ModuleVersionDB) {
	t.Helper()

	// Create namespace
	ns := CreateNamespace(t, db, namespace)

	// Create module provider
	modProvider := CreateModuleProvider(t, db, ns.ID, moduleName, provider)

	// Create versions
	modVersions := make([]*sqldb.ModuleVersionDB, len(versions))
	for i, version := range versions {
		modVersion := CreateModuleVersion(t, db, modProvider.ID, version)
		modVersions[i] = &modVersion
	}

	return &ns, &modProvider, modVersions
}

// SetupModuleWithSubmodules creates a module version with submodules
func SetupModuleWithSubmodules(t *testing.T, db *sqldb.Database, namespace, moduleName, provider, version string) (*sqldb.NamespaceDB, *sqldb.ModuleProviderDB, *sqldb.ModuleVersionDB, []*sqldb.SubmoduleDB) {
	t.Helper()

	// Create module hierarchy
	ns, modProvider, modVersion := SetupModuleVersion(t, db, namespace, moduleName, provider, version)

	// Create module details for root
	rootDetails := CreateModuleDetails(t, db, "Root README")
	modVersion.ModuleDetailsID = &rootDetails.ID
	err := db.DB.Save(&modVersion).Error
	require.NoError(t, err)

	// Create submodules
	submodules := make([]*sqldb.SubmoduleDB, 0)

	submodule1Details := CreateModuleDetails(t, db, "Submodule 1 README")
	submodule1 := CreateSubmodule(t, db, modVersion.ID, "submodules/submodule1", "", "", &submodule1Details.ID)
	submodules = append(submodules, &submodule1)

	submodule2Details := CreateModuleDetails(t, db, "Submodule 2 README")
	submodule2 := CreateSubmodule(t, db, modVersion.ID, "submodules/submodule2", "Submodule Two", "network", &submodule2Details.ID)
	submodules = append(submodules, &submodule2)

	return ns, modProvider, modVersion, submodules
}

// SetupModuleWithExamples creates a module version with examples
func SetupModuleWithExamples(t *testing.T, db *sqldb.Database, namespace, moduleName, provider, version string) (*sqldb.NamespaceDB, *sqldb.ModuleProviderDB, *sqldb.ModuleVersionDB, []*sqldb.SubmoduleDB) {
	t.Helper()

	// Create module hierarchy
	ns, modProvider, modVersion := SetupModuleVersion(t, db, namespace, moduleName, provider, version)

	// Create module details for root
	rootDetails := CreateModuleDetails(t, db, "Root README")
	modVersion.ModuleDetailsID = &rootDetails.ID
	err := db.DB.Save(&modVersion).Error
	require.NoError(t, err)

	// Create examples (examples are stored as submodules)
	examples := make([]*sqldb.SubmoduleDB, 0)

	example1Details := CreateModuleDetails(t, db, "Example 1 README")
	example1 := CreateSubmodule(t, db, modVersion.ID, "examples/simple", "Simple Example", "", &example1Details.ID)
	examples = append(examples, &example1)

	// Add example files
	_ = CreateExampleFile(t, db, example1.ID, "main.tf", "terraform { ... }")
	_ = CreateExampleFile(t, db, example1.ID, "variables.tf", "variable \"example\" {}")

	example2Details := CreateModuleDetails(t, db, "Example 2 README")
	example2 := CreateSubmodule(t, db, modVersion.ID, "examples/advanced", "Advanced Example", "", &example2Details.ID)
	examples = append(examples, &example2)

	_ = CreateExampleFile(t, db, example2.ID, "main.tf", "terraform { ... }")

	return ns, modProvider, modVersion, examples
}

// ReconstructNamespaceDomainModel reconstructs the domain model from database
func ReconstructNamespaceDomainModel(t *testing.T, db *sqldb.Database, nsDB *sqldb.NamespaceDB) *model.Namespace {
	t.Helper()

	var nsType model.NamespaceType
	switch nsDB.NamespaceType {
	case sqldb.NamespaceTypeGithubOrg:
		nsType = model.NamespaceTypeGithubOrg
	case sqldb.NamespaceTypeGithubUser:
		nsType = model.NamespaceTypeGithubUser
	default:
		nsType = model.NamespaceTypeNone
	}

	return model.ReconstructNamespace(nsDB.ID, nsDB.Namespace, nsDB.DisplayName, nsType)
}

// SetLatestVersionForProvider sets the latest version for a module provider
func SetLatestVersionForProvider(t *testing.T, db *sqldb.Database, providerID int, versionID int) {
	t.Helper()

	err := db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", providerID).
		Update("latest_version_id", versionID).Error
	require.NoError(t, err)
}

// CreateBetaVersion creates a beta version of a module
func CreateBetaVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string) sqldb.ModuleVersionDB {
	t.Helper()

	betaVersion := sqldb.ModuleVersionDB{
		ModuleProviderID:      moduleProviderID,
		Version:               version,
		Beta:                  true,
		Internal:              false,
		Published:             nil,
		PublishedAt:           nil,
		GitSHA:                nil,
		GitPath:               nil,
		ArchiveGitPath:        false,
		RepoBaseURLTemplate:   nil,
		RepoCloneURLTemplate:  nil,
		RepoBrowseURLTemplate: nil,
		Owner:                 nil,
		Description:           nil,
		VariableTemplate:      nil,
		ExtractionVersion:     nil,
		ModuleDetailsID:       nil,
	}

	err := db.DB.Create(&betaVersion).Error
	require.NoError(t, err)

	return betaVersion
}

// CreateInternalVersion creates an internal version of a module
func CreateInternalVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string) sqldb.ModuleVersionDB {
	t.Helper()

	internalVersion := sqldb.ModuleVersionDB{
		ModuleProviderID:      moduleProviderID,
		Version:               version,
		Beta:                  false,
		Internal:              true,
		Published:             nil,
		PublishedAt:           nil,
		GitSHA:                nil,
		GitPath:               nil,
		ArchiveGitPath:        false,
		RepoBaseURLTemplate:   nil,
		RepoCloneURLTemplate:  nil,
		RepoBrowseURLTemplate: nil,
		Owner:                 nil,
		Description:           nil,
		VariableTemplate:      nil,
		ExtractionVersion:     nil,
		ModuleDetailsID:       nil,
	}

	err := db.DB.Create(&internalVersion).Error
	require.NoError(t, err)

	return internalVersion
}

// CreateVersionWithGitInfo creates a module version with git information
func CreateVersionWithGitInfo(t *testing.T, db *sqldb.Database, moduleProviderID int, version string, gitSHA, gitPath string) sqldb.ModuleVersionDB {
	t.Helper()

	modVersion := sqldb.ModuleVersionDB{
		ModuleProviderID:      moduleProviderID,
		Version:               version,
		Beta:                  false,
		Internal:              false,
		Published:             nil,
		PublishedAt:           nil,
		GitSHA:                &gitSHA,
		GitPath:               &gitPath,
		ArchiveGitPath:        false,
		RepoBaseURLTemplate:   nil,
		RepoCloneURLTemplate:  nil,
		RepoBrowseURLTemplate: nil,
		Owner:                 nil,
		Description:           nil,
		VariableTemplate:      nil,
		ExtractionVersion:     nil,
		ModuleDetailsID:       nil,
	}

	err := db.DB.Create(&modVersion).Error
	require.NoError(t, err)

	return modVersion
}

// PublishModuleVersion publishes an existing module version
func PublishModuleVersion(t *testing.T, db *sqldb.Database, versionID int) {
	t.Helper()

	now := time.Now()
	published := true

	err := db.DB.Model(&sqldb.ModuleVersionDB{}).
		Where("id = ?", versionID).
		Updates(map[string]interface{}{
			"published":    &published,
			"published_at": &now,
		}).Error
	require.NoError(t, err)
}

// UnpublishModuleVersion unpublishes an existing module version
func UnpublishModuleVersion(t *testing.T, db *sqldb.Database, versionID int) {
	t.Helper()

	err := db.DB.Model(&sqldb.ModuleVersionDB{}).
		Where("id = ?", versionID).
		Updates(map[string]interface{}{
			"published":    nil,
			"published_at": nil,
		}).Error
	require.NoError(t, err)
}

// VerifyModuleProvider marks a module provider as verified
func VerifyModuleProvider(t *testing.T, db *sqldb.Database, providerID int) {
	t.Helper()

	err := db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", providerID).
		Update("verified", true).Error
	require.NoError(t, err)
}

// CreateModuleProviderWithGitConfig creates a module provider with git configuration
func CreateModuleProviderWithGitConfig(t *testing.T, db *sqldb.Database, namespaceID int, moduleName, providerName string, gitProviderID int, repoBaseURL, repoCloneURL, repoBrowseURL, gitTagFormat, gitPath string) sqldb.ModuleProviderDB {
	t.Helper()

	moduleProvider := sqldb.ModuleProviderDB{
		NamespaceID:           namespaceID,
		Module:                moduleName,
		Provider:              providerName,
		Verified:              nil,
		GitProviderID:         &gitProviderID,
		RepoBaseURLTemplate:   &repoBaseURL,
		RepoCloneURLTemplate:  &repoCloneURL,
		RepoBrowseURLTemplate: &repoBrowseURL,
		GitTagFormat:          &gitTagFormat,
		GitPath:               &gitPath,
		ArchiveGitPath:        false,
		LatestVersionID:       nil,
	}

	err := db.DB.Create(&moduleProvider).Error
	require.NoError(t, err)

	return moduleProvider
}

// AssertModuleVersionExists asserts that a module version exists in the database
func AssertModuleVersionExists(t *testing.T, db *sqldb.Database, moduleProviderID int, version string) {
	t.Helper()

	var count int64
	err := db.DB.Model(&sqldb.ModuleVersionDB{}).
		Where("module_provider_id = ? AND version = ?", moduleProviderID, version).
		Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(1), count, "Module version should exist")
}

// AssertModuleVersionNotExists asserts that a module version does not exist in the database
func AssertModuleVersionNotExists(t *testing.T, db *sqldb.Database, moduleProviderID int, version string) {
	t.Helper()

	var count int64
	err := db.DB.Model(&sqldb.ModuleVersionDB{}).
		Where("module_provider_id = ? AND version = ?", moduleProviderID, version).
		Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(0), count, "Module version should not exist")
}

// AssertModuleVersionPublished asserts that a module version is published
func AssertModuleVersionPublished(t *testing.T, db *sqldb.Database, versionID int) {
	t.Helper()

	var version sqldb.ModuleVersionDB
	err := db.DB.First(&version, versionID).Error
	require.NoError(t, err)
	require.NotNil(t, version.Published, "Version should be published")
	require.True(t, *version.Published, "Version should be published")
}

// AssertModuleVersionUnpublished asserts that a module version is not published
func AssertModuleVersionUnpublished(t *testing.T, db *sqldb.Database, versionID int) {
	t.Helper()

	var version sqldb.ModuleVersionDB
	err := db.DB.First(&version, versionID).Error
	require.NoError(t, err)
	require.Nil(t, version.Published, "Version should not be published")
}

// GetModuleVersion retrieves a module version from the database
func GetModuleVersion(t *testing.T, db *sqldb.Database, versionID int) sqldb.ModuleVersionDB {
	t.Helper()

	var version sqldb.ModuleVersionDB
	err := db.DB.First(&version, versionID).Error
	require.NoError(t, err)
	return version
}

// GetModuleProvider retrieves a module provider from the database
func GetModuleProvider(t *testing.T, db *sqldb.Database, providerID int) sqldb.ModuleProviderDB {
	t.Helper()

	var provider sqldb.ModuleProviderDB
	err := db.DB.First(&provider, providerID).Error
	require.NoError(t, err)
	return provider
}

// CountVersionsForProvider counts the number of versions for a module provider
func CountVersionsForProvider(t *testing.T, db *sqldb.Database, providerID int) int {
	t.Helper()

	var count int64
	err := db.DB.Model(&sqldb.ModuleVersionDB{}).
		Where("module_provider_id = ?", providerID).
		Count(&count).Error
	require.NoError(t, err)
	return int(count)
}

// GetVersionsForProvider retrieves all versions for a module provider
func GetVersionsForProvider(t *testing.T, db *sqldb.Database, providerID int) []sqldb.ModuleVersionDB {
	t.Helper()

	var versions []sqldb.ModuleVersionDB
	err := db.DB.Where("module_provider_id = ?", providerID).
		Order("version ASC").
		Find(&versions).Error
	require.NoError(t, err)
	return versions
}
