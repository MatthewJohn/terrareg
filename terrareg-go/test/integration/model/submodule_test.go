package model

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestSubmodule_Save tests saving a submodule
func TestSubmodule_Save(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data: namespace, module provider, module version
	namespace := testutils.CreateNamespace(t, db, "test-submodule-save", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create a submodule
	submoduleType := "module"
	submoduleName := "Test Submodule"
	submodule := &sqldb.SubmoduleDB{
		Type: &submoduleType,
		Path: "path/to/submodule",
		Name: &submoduleName,
	}

	// Save the submodule
	saved, err := repo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)
	require.NotNil(t, saved)

	// Verify the submodule was saved
	assert.Greater(t, saved.ID, 0)
	assert.Equal(t, version.ID, saved.ParentModuleVersion)
	assert.Equal(t, submoduleType, *saved.Type)
	assert.Equal(t, "path/to/submodule", saved.Path)
	assert.Equal(t, submoduleName, *saved.Name)
}

// TestSubmodule_SaveWithDetails tests saving a submodule with module details
func TestSubmodule_SaveWithDetails(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-submodule-details", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create module details
	details := sqldb.ModuleDetailsDB{
		ReadmeContent: []byte("# Test README"),
	}
	err = db.DB.Create(&details).Error
	require.NoError(t, err)

	// Create a submodule
	submoduleType := "module"
	submoduleName := "Test Submodule"
	submodule := &sqldb.SubmoduleDB{
		Type: &submoduleType,
		Path: "path/to/submodule",
		Name: &submoduleName,
	}

	// Save the submodule with details
	saved, err := repo.SaveWithDetails(ctx, version.ID, submodule, details.ID)
	require.NoError(t, err)
	require.NotNil(t, saved)

	// Verify the submodule was saved with details
	assert.Greater(t, saved.ID, 0)
	assert.Equal(t, version.ID, saved.ParentModuleVersion)
	assert.Equal(t, details.ID, *saved.ModuleDetailsID)
	assert.Equal(t, submoduleType, *saved.Type)
	assert.Equal(t, "path/to/submodule", saved.Path)
	assert.Equal(t, submoduleName, *saved.Name)
}

// TestSubmodule_FindByParentModuleVersion tests finding submodules by parent module version
func TestSubmodule_FindByParentModuleVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-submodule-find", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version1 := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")
	version2 := testutils.CreateModuleVersion(t, db, provider.ID, "2.0.0")

	// Create submodules for version 1
	submoduleType1 := "module"
	submoduleName1 := "Submodule 1"
	submodule1 := &sqldb.SubmoduleDB{
		Type: &submoduleType1,
		Path: "submodule1",
		Name: &submoduleName1,
	}
	_, err = repo.Save(ctx, version1.ID, submodule1)
	require.NoError(t, err)

	submoduleType2 := "module"
	submoduleName2 := "Submodule 2"
	submodule2 := &sqldb.SubmoduleDB{
		Type: &submoduleType2,
		Path: "submodule2",
		Name: &submoduleName2,
	}
	_, err = repo.Save(ctx, version1.ID, submodule2)
	require.NoError(t, err)

	// Create a submodule for version 2
	submoduleType3 := "module"
	submoduleName3 := "Submodule 3"
	submodule3 := &sqldb.SubmoduleDB{
		Type: &submoduleType3,
		Path: "submodule3",
		Name: &submoduleName3,
	}
	_, err = repo.Save(ctx, version2.ID, submodule3)
	require.NoError(t, err)

	// Find submodules for version 1
	submodules, err := repo.FindByParentModuleVersion(ctx, version1.ID)
	require.NoError(t, err)
	assert.Len(t, submodules, 2)

	// Verify the submodules
	paths := make([]string, 0, 2)
	for _, sub := range submodules {
		paths = append(paths, sub.Path)
	}
	assert.Contains(t, paths, "submodule1")
	assert.Contains(t, paths, "submodule2")

	// Find submodules for version 2
	submodulesV2, err := repo.FindByParentModuleVersion(ctx, version2.ID)
	require.NoError(t, err)
	assert.Len(t, submodulesV2, 1)
	assert.Equal(t, "submodule3", submodulesV2[0].Path)
}

// TestSubmodule_FindByPath tests finding a submodule by path
func TestSubmodule_FindByPath(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-submodule-findpath", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create submodules
	submoduleType1 := "module"
	submoduleName1 := "Submodule 1"
	submodule1 := &sqldb.SubmoduleDB{
		Type: &submoduleType1,
		Path: "path/to/submodule1",
		Name: &submoduleName1,
	}
	_, err = repo.Save(ctx, version.ID, submodule1)
	require.NoError(t, err)

	submoduleType2 := "module"
	submoduleName2 := "Submodule 2"
	submodule2 := &sqldb.SubmoduleDB{
		Type: &submoduleType2,
		Path: "another/path",
		Name: &submoduleName2,
	}
	_, err = repo.Save(ctx, version.ID, submodule2)
	require.NoError(t, err)

	// Find submodule by path
	found, err := repo.FindByPath(ctx, version.ID, "path/to/submodule1")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "path/to/submodule1", found.Path)
	assert.Equal(t, submoduleName1, *found.Name)

	// Find another submodule by path
	found2, err := repo.FindByPath(ctx, version.ID, "another/path")
	require.NoError(t, err)
	require.NotNil(t, found2)
	assert.Equal(t, "another/path", found2.Path)
	assert.Equal(t, submoduleName2, *found2.Name)
}

// TestSubmodule_FindByPath_NotFound tests finding a non-existent submodule by path
func TestSubmodule_FindByPath_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-submodule-notfound", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Try to find non-existent submodule
	found, err := repo.FindByPath(ctx, version.ID, "nonexistent/path")
	assert.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "submodule not found")
}

// TestSubmodule_DeleteByParentModuleVersion tests deleting submodules by parent module version
func TestSubmodule_DeleteByParentModuleVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-submodule-delete", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create submodules
	for i := 1; i <= 3; i++ {
		submoduleType := "module"
		submoduleName := "Submodule"
		path := "submodule"
		submodule := &sqldb.SubmoduleDB{
			Type: &submoduleType,
			Path: path,
			Name: &submoduleName,
		}
		_, err := repo.Save(ctx, version.ID, submodule)
		require.NoError(t, err)
	}

	// Verify submodules exist
	submodules, err := repo.FindByParentModuleVersion(ctx, version.ID)
	require.NoError(t, err)
	assert.Len(t, submodules, 3)

	// Delete submodules
	err = repo.DeleteByParentModuleVersion(ctx, version.ID)
	require.NoError(t, err)

	// Verify submodules were deleted
	submodules, err = repo.FindByParentModuleVersion(ctx, version.ID)
	require.NoError(t, err)
	assert.Len(t, submodules, 0)
}

// TestSubmodule_NilSubmodule tests handling nil submodule
func TestSubmodule_NilSubmodule(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-submodule-nil", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Try to save nil submodule
	_, err = repo.Save(ctx, version.ID, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "submodule cannot be nil")

	// Try to save nil submodule with details
	_, err = repo.SaveWithDetails(ctx, version.ID, nil, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "submodule cannot be nil")
}

// TestSubmodule_OptionalFields tests submodules with optional fields
func TestSubmodule_OptionalFields(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-submodule-optional", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create submodule without type and name (they're optional in some cases)
	submodule := &sqldb.SubmoduleDB{
		Path: "simple/submodule",
		Type: nil,
		Name: nil,
	}

	// Save the submodule
	saved, err := repo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)
	require.NotNil(t, saved)

	// Verify the submodule was saved
	assert.Greater(t, saved.ID, 0)
	assert.Equal(t, version.ID, saved.ParentModuleVersion)
	assert.Equal(t, "simple/submodule", saved.Path)
	assert.Nil(t, saved.Type)
	assert.Nil(t, saved.Name)
}

// TestSubmodule_EmptyPath tests submodules with empty path
func TestSubmodule_EmptyPath(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-submodule-empty", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create submodule with root path (empty string)
	submoduleType := "root"
	submoduleName := "Root Module"
	submodule := &sqldb.SubmoduleDB{
		Type: &submoduleType,
		Path: "",
		Name: &submoduleName,
	}

	// Save the submodule
	saved, err := repo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)
	require.NotNil(t, saved)

	// Verify the submodule was saved
	assert.Greater(t, saved.ID, 0)
	assert.Equal(t, "", saved.Path)
	assert.Equal(t, submoduleType, *saved.Type)
	assert.Equal(t, submoduleName, *saved.Name)
}

// TestSubmodule_MultipleVersions tests submodules across multiple module versions
func TestSubmodule_MultipleVersions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-submodule-multiversion", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version1 := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")
	version2 := testutils.CreateModuleVersion(t, db, provider.ID, "2.0.0")
	version3 := testutils.CreateModuleVersion(t, db, provider.ID, "3.0.0")

	// Create different submodules for each version
	versions := []struct {
		versionID  int
		submodules []string
	}{
		{version1.ID, []string{"sub1", "sub2"}},
		{version2.ID, []string{"sub1", "sub2", "sub3"}},
		{version3.ID, []string{"sub1"}},
	}

	for _, v := range versions {
		for _, path := range v.submodules {
			submoduleType := "module"
			submoduleName := path + " name"
			submodule := &sqldb.SubmoduleDB{
				Type: &submoduleType,
				Path: path,
				Name: &submoduleName,
			}
			_, err := repo.Save(ctx, v.versionID, submodule)
			require.NoError(t, err)
		}
	}

	// Verify each version has correct submodules
	for _, v := range versions {
		submodules, err := repo.FindByParentModuleVersion(ctx, v.versionID)
		require.NoError(t, err)
		assert.Len(t, submodules, len(v.submodules))
	}
}

// TestSubmodule_WithModuleDetails tests submodules with module details preloaded
func TestSubmodule_WithModuleDetails(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-submodule-preload", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create module details
	details := sqldb.ModuleDetailsDB{
		ReadmeContent: []byte("# README for submodule"),
	}
	err = db.DB.Create(&details).Error
	require.NoError(t, err)

	// Create submodule with details
	submoduleType := "module"
	submoduleName := "Submodule with details"
	submodule := &sqldb.SubmoduleDB{
		Type: &submoduleType,
		Path: "path/to/submodule",
		Name: &submoduleName,
	}
	_, err = repo.SaveWithDetails(ctx, version.ID, submodule, details.ID)
	require.NoError(t, err)

	// Find by path (which preloads ModuleDetails)
	found, err := repo.FindByPath(ctx, version.ID, "path/to/submodule")
	require.NoError(t, err)
	require.NotNil(t, found)

	// Verify ModuleDetails is preloaded
	assert.NotNil(t, found.ModuleDetails)
	assert.Equal(t, []byte("# README for submodule"), found.ModuleDetails.ReadmeContent)
}

// TestSubmodule_UpdateModuleDetailsID_Success tests successfully updating module details ID for a submodule
func TestSubmodule_UpdateModuleDetailsID_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data: namespace, module provider, module version
	namespace := testutils.CreateNamespace(t, db, "test-submodule-update-details", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create a submodule
	submoduleType := "module"
	submoduleName := "Test Submodule"
	submodule := &sqldb.SubmoduleDB{
		Type: &submoduleType,
		Path: "examples/test-example",
		Name: &submoduleName,
	}
	saved, err := repo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)
	require.NotNil(t, saved)

	// Create a module details record
	moduleDetails := sqldb.ModuleDetailsDB{
		ReadmeContent: []byte("# Test README"),
	}
	err = db.DB.Create(&moduleDetails).Error
	require.NoError(t, err)

	// Update submodule with module details ID
	err = repo.UpdateModuleDetailsID(ctx, saved.ID, moduleDetails.ID)
	assert.NoError(t, err, "UpdateModuleDetailsID should succeed")

	// Verify the update
	var updatedSubmodule sqldb.SubmoduleDB
	err = db.DB.First(&updatedSubmodule, saved.ID).Error
	require.NoError(t, err, "Should be able to fetch updated submodule")

	assert.Equal(t, moduleDetails.ID, *updatedSubmodule.ModuleDetailsID, "ModuleDetailsID should be updated")
}

// TestSubmodule_UpdateModuleDetailsID_ZeroModuleDetailsID tests updating with zero module details ID (clearing)
func TestSubmodule_UpdateModuleDetailsID_ZeroModuleDetailsID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data: namespace, module provider, module version
	namespace := testutils.CreateNamespace(t, db, "test-submodule-clear-details", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create module details
	moduleDetails := sqldb.ModuleDetailsDB{
		ReadmeContent: []byte("# Test README"),
	}
	err = db.DB.Create(&moduleDetails).Error
	require.NoError(t, err)

	// Create a submodule with module details ID
	submoduleType := "module"
	submoduleName := "Test Submodule"
	submodule := &sqldb.SubmoduleDB{
		Type:            &submoduleType,
		Path:            "examples/test-example",
		Name:            &submoduleName,
		ModuleDetailsID: &moduleDetails.ID,
	}
	saved, err := repo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)
	require.NotNil(t, saved)

	// Update submodule with zero module details ID
	// Note: In GORM, updating to 0 sets the field to a pointer to 0, not nil
	err = repo.UpdateModuleDetailsID(ctx, saved.ID, 0)
	assert.NoError(t, err, "UpdateModuleDetailsID should succeed with zero module details ID")

	// Verify the update
	var updatedSubmodule sqldb.SubmoduleDB
	err = db.DB.First(&updatedSubmodule, saved.ID).Error
	require.NoError(t, err, "Should be able to fetch updated submodule")

	// GORM sets to a pointer to 0, not nil
	assert.NotNil(t, updatedSubmodule.ModuleDetailsID, "ModuleDetailsID is a pointer (not nil)")
	assert.Equal(t, 0, *updatedSubmodule.ModuleDetailsID, "ModuleDetailsID should be 0")
}

// TestSubmodule_UpdateModuleDetailsID_MultipleUpdates tests updating the same submodule multiple times
func TestSubmodule_UpdateModuleDetailsID_MultipleUpdates(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo, err := module.NewSubmoduleRepository(db.DB)
	require.NoError(t, err)

	// Create test data: namespace, module provider, module version
	namespace := testutils.CreateNamespace(t, db, "test-submodule-multiple-updates", nil)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create a submodule
	submoduleType := "module"
	submoduleName := "Test Submodule"
	submodule := &sqldb.SubmoduleDB{
		Type: &submoduleType,
		Path: "examples/test-example",
		Name: &submoduleName,
	}
	saved, err := repo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)
	require.NotNil(t, saved)

	// Create multiple module details records
	moduleDetails1 := sqldb.ModuleDetailsDB{
		ReadmeContent: []byte("# README 1"),
	}
	err = db.DB.Create(&moduleDetails1).Error
	require.NoError(t, err)

	moduleDetails2 := sqldb.ModuleDetailsDB{
		ReadmeContent: []byte("# README 2"),
	}
	err = db.DB.Create(&moduleDetails2).Error
	require.NoError(t, err)

	// First update
	err = repo.UpdateModuleDetailsID(ctx, saved.ID, moduleDetails1.ID)
	assert.NoError(t, err, "First update should succeed")

	// Verify first update
	var updated1 sqldb.SubmoduleDB
	err = db.DB.First(&updated1, saved.ID).Error
	require.NoError(t, err)
	assert.Equal(t, moduleDetails1.ID, *updated1.ModuleDetailsID, "First update should set ModuleDetailsID")

	// Second update
	err = repo.UpdateModuleDetailsID(ctx, saved.ID, moduleDetails2.ID)
	assert.NoError(t, err, "Second update should succeed")

	// Verify second update
	var updated2 sqldb.SubmoduleDB
	err = db.DB.First(&updated2, saved.ID).Error
	require.NoError(t, err)
	assert.Equal(t, moduleDetails2.ID, *updated2.ModuleDetailsID, "Second update should change ModuleDetailsID")
}

