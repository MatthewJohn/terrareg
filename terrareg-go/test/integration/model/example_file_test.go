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

// TestExampleFile_Save tests saving an example file
func TestExampleFile_Save(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	submoduleRepo := module.NewSubmoduleRepository(db.DB)
	exampleFileRepo := module.NewExampleFileRepository(db.DB)

	// Create test data: namespace, module provider, module version
	namespace := testutils.CreateNamespace(t, db, "test-examplefile-save")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create an example submodule
	exampleType := "example"
	exampleName := "Test Example"
	submodule := &sqldb.SubmoduleDB{
		Type: &exampleType,
		Path: "examples/testexample",
		Name: &exampleName,
	}
	savedSubmodule, err := submoduleRepo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)

	// Create an example file
	exampleFile := &sqldb.ExampleFileDB{
		SubmoduleID: savedSubmodule.ID,
		Path:        "main.tf",
		Content:     []byte(`module "test" { source = "../.." }`),
	}

	// Save the example file
	saved, err := exampleFileRepo.Save(ctx, exampleFile)
	require.NoError(t, err)
	require.NotNil(t, saved)

	// Verify the example file was saved
	assert.Greater(t, saved.ID, 0)
	assert.Equal(t, savedSubmodule.ID, saved.SubmoduleID)
	assert.Equal(t, "main.tf", saved.Path)
	assert.Equal(t, []byte(`module "test" { source = "../.." }`), saved.Content)
}

// TestExampleFile_SaveBatch tests saving multiple example files in a batch
func TestExampleFile_SaveBatch(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	submoduleRepo := module.NewSubmoduleRepository(db.DB)
	exampleFileRepo := module.NewExampleFileRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-examplefile-batch")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create an example submodule
	exampleType := "example"
	exampleName := "Batch Example"
	submodule := &sqldb.SubmoduleDB{
		Type: &exampleType,
		Path: "examples/batchexample",
		Name: &exampleName,
	}
	savedSubmodule, err := submoduleRepo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)

	// Create multiple example files
	exampleFiles := []*sqldb.ExampleFileDB{
		{
			SubmoduleID: savedSubmodule.ID,
			Path:        "main.tf",
			Content:     []byte(`variable "region" { type = string }`),
		},
		{
			SubmoduleID: savedSubmodule.ID,
			Path:        "variables.tf",
			Content:     []byte(`variable "name" { description = "Name" }`),
		},
		{
			SubmoduleID: savedSubmodule.ID,
			Path:        "outputs.tf",
			Content:     []byte(`output "id" { value = "test" }`),
		},
		{
			SubmoduleID: savedSubmodule.ID,
			Path:        "README.md",
			Content:     []byte("# Example\n\nThis is an example."),
		},
	}

	// Save the example files in a batch
	saved, err := exampleFileRepo.SaveBatch(ctx, exampleFiles)
	require.NoError(t, err)
	assert.Len(t, saved, 4)

	// Verify each file was saved and has an ID
	for _, file := range saved {
		assert.Greater(t, file.ID, 0)
		assert.Equal(t, savedSubmodule.ID, file.SubmoduleID)
	}
}

// TestExampleFile_FindBySubmoduleID tests finding example files by submodule ID
func TestExampleFile_FindBySubmoduleID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	submoduleRepo := module.NewSubmoduleRepository(db.DB)
	exampleFileRepo := module.NewExampleFileRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-examplefile-find")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create two example submodules
	exampleType := "example"
	exampleName1 := "Example 1"
	submodule1 := &sqldb.SubmoduleDB{
		Type: &exampleType,
		Path: "examples/example1",
		Name: &exampleName1,
	}
	savedSubmodule1, err := submoduleRepo.Save(ctx, version.ID, submodule1)
	require.NoError(t, err)

	exampleName2 := "Example 2"
	submodule2 := &sqldb.SubmoduleDB{
		Type: &exampleType,
		Path: "examples/example2",
		Name: &exampleName2,
	}
	savedSubmodule2, err := submoduleRepo.Save(ctx, version.ID, submodule2)
	require.NoError(t, err)

	// Create example files for each submodule
	files1 := []*sqldb.ExampleFileDB{
		{SubmoduleID: savedSubmodule1.ID, Path: "main.tf", Content: []byte("file 1 content")},
		{SubmoduleID: savedSubmodule1.ID, Path: "variables.tf", Content: []byte("file 2 content")},
	}
	_, err = exampleFileRepo.SaveBatch(ctx, files1)
	require.NoError(t, err)

	files2 := []*sqldb.ExampleFileDB{
		{SubmoduleID: savedSubmodule2.ID, Path: "main.tf", Content: []byte("file 3 content")},
	}
	_, err = exampleFileRepo.SaveBatch(ctx, files2)
	require.NoError(t, err)

	// Find files for submodule 1
	files, err := exampleFileRepo.FindBySubmoduleID(ctx, savedSubmodule1.ID)
	require.NoError(t, err)
	assert.Len(t, files, 2)

	paths := make([]string, 0, 2)
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	assert.Contains(t, paths, "main.tf")
	assert.Contains(t, paths, "variables.tf")

	// Find files for submodule 2
	files2Found, err := exampleFileRepo.FindBySubmoduleID(ctx, savedSubmodule2.ID)
	require.NoError(t, err)
	assert.Len(t, files2Found, 1)
	assert.Equal(t, "main.tf", files2Found[0].Path)
}

// TestExampleFile_DeleteBySubmoduleID tests deleting example files by submodule ID
func TestExampleFile_DeleteBySubmoduleID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	submoduleRepo := module.NewSubmoduleRepository(db.DB)
	exampleFileRepo := module.NewExampleFileRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-examplefile-delete")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create an example submodule
	exampleType := "example"
	exampleName := "Delete Example"
	submodule := &sqldb.SubmoduleDB{
		Type: &exampleType,
		Path: "examples/deleteexample",
		Name: &exampleName,
	}
	savedSubmodule, err := submoduleRepo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)

	// Create example files
	files := []*sqldb.ExampleFileDB{
		{SubmoduleID: savedSubmodule.ID, Path: "main.tf", Content: []byte("content 1")},
		{SubmoduleID: savedSubmodule.ID, Path: "outputs.tf", Content: []byte("content 2")},
		{SubmoduleID: savedSubmodule.ID, Path: "README.md", Content: []byte("content 3")},
	}
	_, err = exampleFileRepo.SaveBatch(ctx, files)
	require.NoError(t, err)

	// Verify files exist
	foundFiles, err := exampleFileRepo.FindBySubmoduleID(ctx, savedSubmodule.ID)
	require.NoError(t, err)
	assert.Len(t, foundFiles, 3)

	// Delete files by submodule ID
	err = exampleFileRepo.DeleteBySubmoduleID(ctx, savedSubmodule.ID)
	require.NoError(t, err)

	// Verify files were deleted
	foundFiles, err = exampleFileRepo.FindBySubmoduleID(ctx, savedSubmodule.ID)
	require.NoError(t, err)
	assert.Len(t, foundFiles, 0)
}

// TestExampleFile_DeleteByModuleVersion tests deleting example files by module version
func TestExampleFile_DeleteByModuleVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	submoduleRepo := module.NewSubmoduleRepository(db.DB)
	exampleFileRepo := module.NewExampleFileRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-examplefile-deletever")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version1 := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")
	version2 := testutils.CreateModuleVersion(t, db, provider.ID, "2.0.0")

	// Create example submodules for each version
	exampleType := "example"
	versions := []*sqldb.ModuleVersionDB{&version1, &version2}
	for i, version := range versions {
		exampleName := "Example"
		submodule := &sqldb.SubmoduleDB{
			Type: &exampleType,
			Path: "examples/deletebyver",
			Name: &exampleName,
		}
		savedSubmodule, err := submoduleRepo.Save(ctx, version.ID, submodule)
		require.NoError(t, err)

		files := []*sqldb.ExampleFileDB{
			{SubmoduleID: savedSubmodule.ID, Path: "main.tf", Content: []byte("content")},
		}
		_, err = exampleFileRepo.SaveBatch(ctx, files)
		require.NoError(t, err)

		if i == 0 {
			// Verify version 1 files exist
			foundFiles, err := exampleFileRepo.FindBySubmoduleID(ctx, savedSubmodule.ID)
			require.NoError(t, err)
			assert.Len(t, foundFiles, 1)
		}
	}

	// Delete files for version 1
	err := exampleFileRepo.DeleteByModuleVersion(ctx, version1.ID)
	require.NoError(t, err)

	// Verify version 1 files were deleted
	var submoduleIDs []int
	err = db.DB.Table("submodule").
		Select("id").
		Where("parent_module_version = ? AND type = ?", version1.ID, "example").
		Scan(&submoduleIDs).Error
	require.NoError(t, err)

	if len(submoduleIDs) > 0 {
		foundFiles, err := exampleFileRepo.FindBySubmoduleID(ctx, submoduleIDs[0])
		require.NoError(t, err)
		assert.Len(t, foundFiles, 0)
	}

	// Verify version 2 files still exist
	var submodule2IDs []int
	err = db.DB.Table("submodule").
		Select("id").
		Where("parent_module_version = ? AND type = ?", version2.ID, "example").
		Scan(&submodule2IDs).Error
	require.NoError(t, err)
	assert.Len(t, submodule2IDs, 1)

	foundFiles2, err := exampleFileRepo.FindBySubmoduleID(ctx, submodule2IDs[0])
	require.NoError(t, err)
	assert.Len(t, foundFiles2, 1)
}

// TestExampleFile_NilFile tests handling nil example file
func TestExampleFile_NilFile(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	exampleFileRepo := module.NewExampleFileRepository(db.DB)

	// Try to save nil example file
	_, err := exampleFileRepo.Save(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "example file cannot be nil")
}

// TestExampleFile_EmptyBatch tests saving an empty batch of example files
func TestExampleFile_EmptyBatch(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	exampleFileRepo := module.NewExampleFileRepository(db.DB)

	// Save empty batch - should succeed without error
	saved, err := exampleFileRepo.SaveBatch(ctx, []*sqldb.ExampleFileDB{})
	require.NoError(t, err)
	assert.Empty(t, saved)
}

// TestExampleFile_WithReadme tests example file with README content
func TestExampleFile_WithReadme(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	submoduleRepo := module.NewSubmoduleRepository(db.DB)
	exampleFileRepo := module.NewExampleFileRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-examplefile-readme")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create an example submodule
	exampleType := "example"
	exampleName := "Example with README"
	submodule := &sqldb.SubmoduleDB{
		Type: &exampleType,
		Path: "examples/withreadme",
		Name: &exampleName,
	}
	savedSubmodule, err := submoduleRepo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)

	// Create example files including a README (using regular string literal because backtick strings can't contain backticks)
	readmeContent := []byte("# Example README\n\nThis example demonstrates how to use the module.\n\n## Usage\n\n```hcl\nmodule \"example\" {\n  source = \"../..\"\n}\n```\n")

	exampleFiles := []*sqldb.ExampleFileDB{
		{SubmoduleID: savedSubmodule.ID, Path: "main.tf", Content: []byte(`module "test" { source = "../.." }`)},
		{SubmoduleID: savedSubmodule.ID, Path: "README.md", Content: readmeContent},
		{SubmoduleID: savedSubmodule.ID, Path: "variables.tf", Content: []byte(`variable "region" { type = string }`)},
	}

	// Save the files
	saved, err := exampleFileRepo.SaveBatch(ctx, exampleFiles)
	require.NoError(t, err)
	assert.Len(t, saved, 3)

	// Find and verify the README
	foundFiles, err := exampleFileRepo.FindBySubmoduleID(ctx, savedSubmodule.ID)
	require.NoError(t, err)
	assert.Len(t, foundFiles, 3)

	var readmeFile *sqldb.ExampleFileDB
	for _, file := range foundFiles {
		if file.Path == "README.md" {
			readmeFile = &file
			break
		}
	}
	require.NotNil(t, readmeFile)
	assert.Equal(t, readmeContent, readmeFile.Content)
	assert.Contains(t, string(readmeFile.Content), "# Example README")
	assert.Contains(t, string(readmeFile.Content), "## Usage")
}

// TestExampleFile_MultipleExamples tests multiple examples for a single module version
func TestExampleFile_MultipleExamples(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	submoduleRepo := module.NewSubmoduleRepository(db.DB)
	exampleFileRepo := module.NewExampleFileRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-examplefile-multiple")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create multiple example submodules
	examples := []struct {
		path       string
		name       string
		fileCount  int
	}{
		{"examples/basic", "Basic Example", 2},
		{"examples/advanced", "Advanced Example", 4},
		{"examples/minimal", "Minimal Example", 1},
	}

	for _, ex := range examples {
		exampleType := "example"
		submodule := &sqldb.SubmoduleDB{
			Type: &exampleType,
			Path: ex.path,
			Name: &ex.name,
		}
		savedSubmodule, err := submoduleRepo.Save(ctx, version.ID, submodule)
		require.NoError(t, err)

		files := make([]*sqldb.ExampleFileDB, ex.fileCount)
		for i := 0; i < ex.fileCount; i++ {
			files[i] = &sqldb.ExampleFileDB{
				SubmoduleID: savedSubmodule.ID,
				Path:        "file.tf",
				Content:     []byte("content"),
			}
		}
		_, err = exampleFileRepo.SaveBatch(ctx, files)
		require.NoError(t, err)
	}

	// Verify each example has the correct number of files
	for _, ex := range examples {
		var submodule sqldb.SubmoduleDB
		err := db.DB.Where("parent_module_version = ? AND path = ?", version.ID, ex.path).
			First(&submodule).Error
		require.NoError(t, err)

		files, err := exampleFileRepo.FindBySubmoduleID(ctx, submodule.ID)
		require.NoError(t, err)
		assert.Len(t, files, ex.fileCount)
	}
}

// TestExampleFile_EmptyContent tests example file with empty content
func TestExampleFile_EmptyContent(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	submoduleRepo := module.NewSubmoduleRepository(db.DB)
	exampleFileRepo := module.NewExampleFileRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-examplefile-empty")
	provider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")
	version := testutils.CreateModuleVersion(t, db, provider.ID, "1.0.0")

	// Create an example submodule
	exampleType := "example"
	exampleName := "Empty Content Example"
	submodule := &sqldb.SubmoduleDB{
		Type: &exampleType,
		Path: "examples/emptycontent",
		Name: &exampleName,
	}
	savedSubmodule, err := submoduleRepo.Save(ctx, version.ID, submodule)
	require.NoError(t, err)

	// Create example file with empty content
	exampleFile := &sqldb.ExampleFileDB{
		SubmoduleID: savedSubmodule.ID,
		Path:        "empty.tf",
		Content:     []byte{},
	}

	// Save the example file
	saved, err := exampleFileRepo.Save(ctx, exampleFile)
	require.NoError(t, err)

	// Verify the file was saved with empty content
	assert.Greater(t, saved.ID, 0)
	assert.Equal(t, savedSubmodule.ID, saved.SubmoduleID)
	assert.Empty(t, saved.Content)
}
