package model

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestModuleDetails_Create tests creating a ModuleDetails row/object
func TestModuleDetails_Create(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo := module.NewModuleDetailsRepository(db.DB)

	// Create new module details object (without any content initially)
	details := model.NewModuleDetails(nil)

	// Use SaveAndReturnID to get the ID directly
	detailsID, err := repo.SaveAndReturnID(ctx, details)
	require.NoError(t, err)

	// Verify ID is assigned (will be non-zero)
	assert.Greater(t, detailsID, 0, "ID should be assigned and greater than 0")

	// Verify we can find the details by ID
	savedDetails, err := repo.FindByID(ctx, detailsID)
	require.NoError(t, err)
	require.NotNil(t, savedDetails)

	// Verify all fields are empty/unset for a new ModuleDetails created with nil
	assert.Nil(t, savedDetails.ReadmeContent())
	assert.Nil(t, savedDetails.TerraformDocs())
	assert.Nil(t, savedDetails.Tfsec())
	assert.Nil(t, savedDetails.Infracost())
	assert.Nil(t, savedDetails.TerraformGraph())
	assert.Nil(t, savedDetails.TerraformModules())
	assert.Empty(t, savedDetails.TerraformVersion())
}

// getModuleDetailsID retrieves the database ID for a module details entity
// This is a helper since the domain model doesn't expose the ID
func getModuleDetailsID(t *testing.T, db *sqldb.Database, details *model.ModuleDetails) int {
	t.Helper()

	var dbDetails sqldb.ModuleDetailsDB
	err := db.DB.Where("readme_content = ?", details.ReadmeContent()).
		Or("terraform_docs = ?", details.TerraformDocs()).
		First(&dbDetails).Error
	require.NoError(t, err)
	return dbDetails.ID
}

// TestModuleDetails_UpdateAttributes tests update_attributes method of ModuleDetails
func TestModuleDetails_UpdateAttributes(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo := module.NewModuleDetailsRepository(db.DB)

	testReadmeContent := []byte("test readme content")
	testTerraformDocs := []byte(`{"test": "output"}`)
	testTfsec := []byte(`{"results": [{"test_result": 0}]}`)
	testInfracost := []byte(`{"totalMonthlyCost": "123.321"}`)

	// Create new module details object
	details := model.NewModuleDetails(testReadmeContent)
	savedDetails, err := repo.Save(ctx, details)
	require.NoError(t, err)

	// Get the ID for updates
	detailsID := getModuleDetailsID(t, db, savedDetails)

	// Update attributes using immutable updates
	updatedDetails := savedDetails.
		WithTerraformDocs(testTerraformDocs).
		WithTfsec(testTfsec).
		WithInfracost(testInfracost)

	updated, err := repo.Update(ctx, detailsID, updatedDetails)
	require.NoError(t, err)

	// Verify all attributes were updated
	assert.Equal(t, testReadmeContent, updated.ReadmeContent())
	assert.Equal(t, testTerraformDocs, updated.TerraformDocs())
	assert.Equal(t, testTfsec, updated.Tfsec())
	assert.Equal(t, testInfracost, updated.Infracost())
}

// TestModuleDetails_Delete tests delete method of ModuleDetails
func TestModuleDetails_Delete(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo := module.NewModuleDetailsRepository(db.DB)

	// Create new module details object
	details := model.NewModuleDetails([]byte("test readme"))
	savedDetails, err := repo.Save(ctx, details)
	require.NoError(t, err)

	detailsID := getModuleDetailsID(t, db, savedDetails)

	// Ensure the row can be found in the database
	var dbDetails sqldb.ModuleDetailsDB
	err = db.DB.First(&dbDetails, detailsID).Error
	require.NoError(t, err, "Row should exist before deletion")
	assert.NotNil(t, dbDetails)

	// Delete module details
	err = repo.Delete(ctx, detailsID)
	require.NoError(t, err)

	// Ensure the row is no longer present in DB
	err = db.DB.First(&dbDetails, detailsID).Error
	assert.Error(t, err, "Row should not exist after deletion")
}

// TestModuleDetails_FindByID tests finding ModuleDetails by ID
func TestModuleDetails_FindByID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo := module.NewModuleDetailsRepository(db.DB)

	// Create a module details with all fields populated
	testReadme := []byte("# Test README")
	testTfDocs := []byte(`{"terraform": "docs"}`)
	testTfsec := []byte(`{"tfsec": "results"}`)
	testInfracost := []byte(`{"infracost": "data"}`)
	testGraph := []byte(`{"graph": "data"}`)
	testModules := []byte(`{"modules": "list"}`)
	testVersion := "1.5.0"

	details := model.NewCompleteModuleDetails(
		testReadme,
		testTfDocs,
		testTfsec,
		testInfracost,
		testGraph,
		testModules,
		testVersion,
	)

	savedDetails, err := repo.Save(ctx, details)
	require.NoError(t, err)

	detailsID := getModuleDetailsID(t, db, savedDetails)

	// Find by ID
	found, err := repo.FindByID(ctx, detailsID)
	require.NoError(t, err)
	require.NotNil(t, found)

	// Verify all fields match
	assert.Equal(t, testReadme, found.ReadmeContent())
	assert.Equal(t, testTfDocs, found.TerraformDocs())
	assert.Equal(t, testTfsec, found.Tfsec())
	assert.Equal(t, testInfracost, found.Infracost())
	assert.Equal(t, testGraph, found.TerraformGraph())
	assert.Equal(t, testModules, found.TerraformModules())
	assert.Equal(t, testVersion, found.TerraformVersion())
}

// TestModuleDetails_FindByNonExistentID tests finding non-existent ModuleDetails
func TestModuleDetails_FindByNonExistentID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo := module.NewModuleDetailsRepository(db.DB)

	// Try to find a non-existent ID
	found, err := repo.FindByID(ctx, 99999)
	require.NoError(t, err)
	assert.Nil(t, found, "Non-existent ID should return nil")
}

// TestModuleDetails_UpdateWithAllFields tests updating all fields
func TestModuleDetails_UpdateWithAllFields(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo := module.NewModuleDetailsRepository(db.DB)

	// Create initial details
	details := model.NewModuleDetails([]byte("initial readme"))
	savedDetails, err := repo.Save(ctx, details)
	require.NoError(t, err)

	detailsID := getModuleDetailsID(t, db, savedDetails)

	// Update with all fields
	newReadme := []byte("updated readme")
	newTfDocs := []byte(`{"updated": "terraform"}`)
	newTfsec := []byte(`{"updated": "tfsec"}`)
	newInfracost := []byte(`{"updated": "infracost"}`)
	newGraph := []byte(`{"updated": "graph"}`)
	newModules := []byte(`{"updated": "modules"}`)
	newVersion := "2.0.0"

	updatedDetails := model.NewCompleteModuleDetails(
		newReadme,
		newTfDocs,
		newTfsec,
		newInfracost,
		newGraph,
		newModules,
		newVersion,
	)

	updated, err := repo.Update(ctx, detailsID, updatedDetails)
	require.NoError(t, err)

	// Verify all fields were updated
	assert.Equal(t, newReadme, updated.ReadmeContent())
	assert.Equal(t, newTfDocs, updated.TerraformDocs())
	assert.Equal(t, newTfsec, updated.Tfsec())
	assert.Equal(t, newInfracost, updated.Infracost())
	assert.Equal(t, newGraph, updated.TerraformGraph())
	assert.Equal(t, newModules, updated.TerraformModules())
	assert.Equal(t, newVersion, updated.TerraformVersion())
}

// TestModuleDetails_Equals tests the Equals method
func TestModuleDetails_Equals(t *testing.T) {
	t.Run("equal details", func(t *testing.T) {
		readme := []byte("test readme")
		details1 := model.NewModuleDetails(readme)
		details2 := model.NewModuleDetails(readme)

		assert.True(t, details1.Equals(details2))
		assert.True(t, details2.Equals(details1))
	})

	t.Run("not equal - different readme", func(t *testing.T) {
		details1 := model.NewModuleDetails([]byte("readme 1"))
		details2 := model.NewModuleDetails([]byte("readme 2"))

		assert.False(t, details1.Equals(details2))
	})

	t.Run("not equal - nil details", func(t *testing.T) {
		details := model.NewModuleDetails([]byte("test"))

		assert.False(t, details.Equals(nil))
	})

	t.Run("equal with terraform docs", func(t *testing.T) {
		readme := []byte("readme")
		tfDocs := []byte(`{"docs": "test"}`)

		details1 := model.NewModuleDetails(readme).WithTerraformDocs(tfDocs)
		details2 := model.NewModuleDetails(readme).WithTerraformDocs(tfDocs)

		assert.True(t, details1.Equals(details2))
	})

	t.Run("not equal - different terraform docs", func(t *testing.T) {
		readme := []byte("readme")

		details1 := model.NewModuleDetails(readme).WithTerraformDocs([]byte(`{"a": "1"}`))
		details2 := model.NewModuleDetails(readme).WithTerraformDocs([]byte(`{"b": "2"}`))

		assert.False(t, details1.Equals(details2))
	})

	t.Run("equal with all fields", func(t *testing.T) {
		details1 := model.NewCompleteModuleDetails(
			[]byte("readme"),
			[]byte(`{"tf": "docs"}`),
			[]byte(`{"tfsec": "results"}`),
			[]byte(`{"infracost": "data"}`),
			[]byte(`{"graph": "data"}`),
			[]byte(`{"modules": "list"}`),
			"1.0.0",
		)
		details2 := model.NewCompleteModuleDetails(
			[]byte("readme"),
			[]byte(`{"tf": "docs"}`),
			[]byte(`{"tfsec": "results"}`),
			[]byte(`{"infracost": "data"}`),
			[]byte(`{"graph": "data"}`),
			[]byte(`{"modules": "list"}`),
			"1.0.0",
		)

		assert.True(t, details1.Equals(details2))
	})
}

// TestModuleDetails_HasMethods tests the Has* methods
func TestModuleDetails_HasMethods(t *testing.T) {
	t.Run("empty details", func(t *testing.T) {
		details := model.NewModuleDetails(nil)

		assert.False(t, details.HasReadme())
		assert.False(t, details.HasTerraformDocs())
		assert.False(t, details.HasTfsec())
		assert.False(t, details.HasInfracost())
		assert.False(t, details.HasTerraformGraph())
		assert.False(t, details.HasTerraformModules())
	})

	t.Run("with readme", func(t *testing.T) {
		details := model.NewModuleDetails([]byte("readme content"))

		assert.True(t, details.HasReadme())
		assert.False(t, details.HasTerraformDocs())
	})

	t.Run("with all fields", func(t *testing.T) {
		details := model.NewCompleteModuleDetails(
			[]byte("readme"),
			[]byte(`{"tf": "docs"}`),
			[]byte(`{"tfsec": "results"}`),
			[]byte(`{"infracost": "data"}`),
			[]byte(`{"graph": "data"}`),
			[]byte(`{"modules": "list"}`),
			"1.0.0",
		)

		assert.True(t, details.HasReadme())
		assert.True(t, details.HasTerraformDocs())
		assert.True(t, details.HasTfsec())
		assert.True(t, details.HasInfracost())
		assert.True(t, details.HasTerraformGraph())
		assert.True(t, details.HasTerraformModules())
	})
}

// TestModuleDetails_ImmutableUpdateTests tests that immutable updates create new instances
func TestModuleDetails_ImmutableUpdateTests(t *testing.T) {
	original := model.NewModuleDetails([]byte("original readme"))

	// WithTerraformDocs
	withDocs := original.WithTerraformDocs([]byte(`{"docs": "test"}`))
	assert.Equal(t, []byte("original readme"), original.ReadmeContent(), "Original should not be modified")
	assert.Nil(t, original.TerraformDocs(), "Original should not have docs")
	assert.Equal(t, []byte(`{"docs": "test"}`), withDocs.TerraformDocs(), "New instance should have docs")
	assert.True(t, original.HasReadme())
	assert.False(t, original.HasTerraformDocs())
	assert.True(t, withDocs.HasTerraformDocs())

	// WithTfsec
	withTfsec := withDocs.WithTfsec([]byte(`{"tfsec": "results"}`))
	assert.Nil(t, withDocs.Tfsec(), "Intermediate instance should not have tfsec")
	assert.Equal(t, []byte(`{"tfsec": "results"}`), withTfsec.Tfsec(), "New instance should have tfsec")

	// WithInfracost
	withInfracost := withTfsec.WithInfracost([]byte(`{"infracost": "data"}`))
	assert.Nil(t, withTfsec.Infracost(), "Intermediate instance should not have infracost")
	assert.Equal(t, []byte(`{"infracost": "data"}`), withInfracost.Infracost(), "New instance should have infracost")

	// WithTerraformGraph
	withGraph := withInfracost.WithTerraformGraph([]byte(`{"graph": "data"}`))
	assert.Nil(t, withInfracost.TerraformGraph(), "Intermediate instance should not have graph")
	assert.Equal(t, []byte(`{"graph": "data"}`), withGraph.TerraformGraph(), "New instance should have graph")

	// WithTerraformModules
	withModules := withGraph.WithTerraformModules([]byte(`{"modules": "list"}`))
	assert.Nil(t, withGraph.TerraformModules(), "Intermediate instance should not have modules")
	assert.Equal(t, []byte(`{"modules": "list"}`), withModules.TerraformModules(), "New instance should have modules")

	// WithTerraformVersion
	withVersion := withModules.WithTerraformVersion("1.5.0")
	assert.Empty(t, withModules.TerraformVersion(), "Intermediate instance should not have version")
	assert.Equal(t, "1.5.0", withVersion.TerraformVersion(), "New instance should have version")
}

// TestModuleDetails_GetGraphJson tests the GetGraphJson method
// Note: This tests the domain model method. Full integration test with module versions
// requires complex test data setup (modules with submodules, examples, etc.)
func TestModuleDetails_GetGraphJson(t *testing.T) {
	t.Run("no graph data", func(t *testing.T) {
		details := model.NewModuleDetails([]byte("readme"))

		// When no graph data is present, should return empty structure
		assert.Nil(t, details.TerraformGraph())
	})

	t.Run("with graph data", func(t *testing.T) {
		// Example graph JSON structure
		graphData := map[string]interface{}{
			"nodes": []map[string]interface{}{
				{
					"data": map[string]interface{}{
						"id": "root",
						"label": "Root Module",
						"child_count": 1,
						"parent": "",
					},
					"style": map[string]interface{}{
						"background-color": "#F8F7F9",
						"font-weight": "bold",
					},
				},
				{
					"data": map[string]interface{}{
						"id": "aws_s3_bucket.test",
						"label": "aws_s3_bucket.test",
						"child_count": 0,
						"parent": "root",
					},
					"style": map[string]interface{}{},
				},
			},
			"edges": []map[string]interface{}{
				{
					"data": map[string]interface{}{
						"id": "root.aws_s3_bucket.test",
						"source": "aws_s3_bucket.test",
						"target": "root",
					},
				},
			},
		}

		graphBytes, err := json.Marshal(graphData)
		require.NoError(t, err)

		details := model.NewModuleDetails(nil).WithTerraformGraph(graphBytes)

		// Verify graph data is stored
		assert.Equal(t, graphBytes, details.TerraformGraph())
		assert.True(t, details.HasTerraformGraph())

		// Verify the data can be unmarshaled back
		var unmarshaled map[string]interface{}
		err = json.Unmarshal(details.TerraformGraph(), &unmarshaled)
		require.NoError(t, err)
		assert.Contains(t, unmarshaled, "nodes")
		assert.Contains(t, unmarshaled, "edges")
	})

	t.Run("invalid graph json", func(t *testing.T) {
		// Invalid JSON should still be stored (validation happens at service layer)
		invalidJSON := []byte(`{invalid json}`)

		details := model.NewModuleDetails(nil).WithTerraformGraph(invalidJSON)

		assert.Equal(t, invalidJSON, details.TerraformGraph())
		assert.True(t, details.HasTerraformGraph())
	})
}

// TestModuleDetails_SaveAndReturnID tests SaveAndReturnID method
func TestModuleDetails_SaveAndReturnID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo := module.NewModuleDetailsRepository(db.DB)

	// Save and get ID in one call
	details := model.NewModuleDetails([]byte("test readme"))
	id, err := repo.SaveAndReturnID(ctx, details)
	require.NoError(t, err)
	assert.Greater(t, id, 0)

	// Verify the record exists
	found, err := repo.FindByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, []byte("test readme"), found.ReadmeContent())
}

// TestModuleDetails_FindByModuleVersionID tests finding ModuleDetails by module version ID
func TestModuleDetails_FindByModuleVersionID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	repo := module.NewModuleDetailsRepository(db.DB)

	// Create module details
	details := model.NewModuleDetails([]byte("test readme"))
	savedDetails, err := repo.Save(ctx, details)
	require.NoError(t, err)

	detailsID := getModuleDetailsID(t, db, savedDetails)

	// Create a namespace
	namespace := testutils.CreateNamespace(t, db, "test-namespace-details")

	// Create a module provider (module name is stored as a string)
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module-details", "test-provider")

	// Create a module version with the details
	version := sqldb.ModuleVersionDB{
		ModuleProviderID: provider.ID,
		Version:          "1.0.0",
		ModuleDetailsID:  &detailsID,
	}
	err = db.DB.Create(&version).Error
	require.NoError(t, err)

	// Find by module version ID
	found, err := repo.FindByModuleVersionID(ctx, version.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, []byte("test readme"), found.ReadmeContent())

	// Test with non-existent module version
	notFound, err := repo.FindByModuleVersionID(ctx, 99999)
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

// TestModuleDetails_NilHandling tests nil handling
func TestModuleDetails_NilHandling(t *testing.T) {
	t.Run("nil details - ReadmeContent", func(t *testing.T) {
		var details *model.ModuleDetails = nil
		assert.Empty(t, details.ReadmeContent())
	})

	t.Run("nil details - TerraformDocs", func(t *testing.T) {
		var details *model.ModuleDetails = nil
		assert.Nil(t, details.TerraformDocs())
	})

	t.Run("nil details - HasReadme", func(t *testing.T) {
		var details *model.ModuleDetails = nil
		assert.False(t, details.HasReadme())
	})

	t.Run("nil details - Equals", func(t *testing.T) {
		var details *model.ModuleDetails = nil
		nonNil := model.NewModuleDetails([]byte("test"))
		assert.False(t, nonNil.Equals(details))
	})
}
