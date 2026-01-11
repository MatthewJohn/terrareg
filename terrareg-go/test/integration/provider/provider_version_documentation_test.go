package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	providerpkg "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	providerrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestProviderVersionDocumentation_GenerateSlugFromName tests slug generation from various name formats
// Python reference: test_provider_version_documentation.py::TestProviderVersionDocumentation::test_generate_slug_from_name
func TestProviderVersionDocumentation_GenerateSlugFromName(t *testing.T) {
	testCases := []struct {
		name         string
		expectedSlug string
	}{
		{"test_name", "test_name"},
		{"test_name.md", "test_name"},
		{"test_name.html", "test_name"},
		{"test_name.markdown", "test_name"},
		{"test_name.html.md", "test_name"},
		{"test_name.html.markdown", "test_name"},
		{"Test Name_with+special$chars-and--double_-_dashes--.md", "test_name_with_special_chars-and_double_-_dashes"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slug := providerpkg.GenerateSlugFromName(tc.name)
			assert.Equal(t, tc.expectedSlug, slug)
		})
	}
}

// TestProviderVersionDocumentation_Create tests creating documentation with various parameter combinations
// Python reference: test_provider_version_documentation.py::TestProviderVersionDocumentation::test_create
func TestProviderVersionDocumentation_Create(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-doc-create")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)

	// Test all documentation types
	docTypes := []sqldb.ProviderDocumentationType{
		sqldb.ProviderDocTypeDataSource,
		sqldb.ProviderDocTypeGuide,
		sqldb.ProviderDocTypeOverview,
		sqldb.ProviderDocTypeResource,
		sqldb.ProviderDocTypeFunction,
	}

	// Test with various optional parameters
	subcategories := []*string{nil, strPtr("Some Subcategory")}
	descriptions := [][]byte{nil, []byte("Some description")}
	titles := []*string{nil, strPtr("Unittest Title")}

	for _, docType := range docTypes {
		for _, subcategory := range subcategories {
			for _, description := range descriptions {
				for _, title := range titles {
					t.Run(string(docType)+"_"+boolStr(subcategory != nil)+"_"+boolStr(description != nil)+"_"+boolStr(title != nil), func(t *testing.T) {
						slug := providerpkg.GenerateSlugFromName("test-provider-documentation.md")

						doc := providerpkg.NewProviderVersionDocumentation(
							versionDB.ID,
							"test-provider-documentation.md",
							slug,
							title,
							description,
							"hcl",
							subcategory,
							"docs/resources/test-provider-documentation.md",
							string(docType),
							[]byte("Some test documentation\nContent!!!"),
						)

						err := providerRepo.SaveDocumentation(ctx, doc)
						require.NoError(t, err)

						// Verify documentation was saved
						savedDoc, err := providerRepo.FindDocumentationByID(ctx, doc.ID())
						require.NoError(t, err)
						require.NotNil(t, savedDoc)

						assert.Equal(t, "test-provider-documentation.md", savedDoc.Name())
						assert.Equal(t, slug, savedDoc.Slug())
						assert.Equal(t, string(docType), savedDoc.DocumentationType())
						assert.Equal(t, "hcl", savedDoc.Language())

						if title != nil {
							assert.Equal(t, title, savedDoc.Title())
						} else {
							assert.Nil(t, savedDoc.Title())
						}

						if subcategory != nil {
							assert.Equal(t, subcategory, savedDoc.Subcategory())
						} else {
							assert.Nil(t, savedDoc.Subcategory())
						}

						if description != nil {
							assert.Equal(t, description, savedDoc.Description())
						} else {
							assert.Nil(t, savedDoc.Description())
						}
					})
				}
			}
		}
	}
}

// TestProviderVersionDocumentation_FindByID tests finding documentation by ID
// Python reference: test_provider_version_documentation.py::TestProviderVersionDocumentation::test_get_by_pk
func TestProviderVersionDocumentation_FindByID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-doc-get")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)

	t.Run("Find existing documentation", func(t *testing.T) {
		docDB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "unittest-docs", "some-unittest-slug", sqldb.ProviderDocTypeGuide)

		doc, err := providerRepo.FindDocumentationByID(ctx, docDB.ID)
		require.NoError(t, err)
		require.NotNil(t, doc)
		assert.Equal(t, docDB.ID, doc.ID())
		assert.Equal(t, "unittest-docs", doc.Name())
	})

	t.Run("Find non-existent documentation returns nil", func(t *testing.T) {
		doc, err := providerRepo.FindDocumentationByID(ctx, 83158315)
		require.NoError(t, err)
		assert.Nil(t, doc)
	})
}

// TestProviderVersionDocumentation_FindByTypeSlugAndLanguage tests finding documentation by type, slug, and language
// Python reference: test_provider_version_documentation.py::TestProviderVersionDocumentation::test_get
func TestProviderVersionDocumentation_FindByTypeSlugAndLanguage(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "initial-providers")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "multiple-versions", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.5.0", gpgKey.ID, false, nil)

	// Create test documentation
	testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "some_resource", "some_resource", sqldb.ProviderDocTypeResource)

	t.Run("Find existing documentation with correct parameters", func(t *testing.T) {
		doc, err := providerRepo.FindDocumentationByTypeSlugAndLanguage(ctx, versionDB.ID, "resource", "some_resource", "hcl")
		require.NoError(t, err)
		require.NotNil(t, doc)
		assert.Equal(t, "some_resource", doc.Slug())
		assert.Equal(t, "resource", doc.DocumentationType())
	})

	t.Run("Wrong language returns nil", func(t *testing.T) {
		doc, err := providerRepo.FindDocumentationByTypeSlugAndLanguage(ctx, versionDB.ID, "resource", "some_resource", "other")
		require.NoError(t, err)
		assert.Nil(t, doc)
	})

	t.Run("Wrong documentation type returns nil", func(t *testing.T) {
		doc, err := providerRepo.FindDocumentationByTypeSlugAndLanguage(ctx, versionDB.ID, "overview", "some_resource", "hcl")
		require.NoError(t, err)
		assert.Nil(t, doc)
	})

	t.Run("Non-existent slug returns nil", func(t *testing.T) {
		doc, err := providerRepo.FindDocumentationByTypeSlugAndLanguage(ctx, versionDB.ID, "resource", "does_not_exist", "hcl")
		require.NoError(t, err)
		assert.Nil(t, doc)
	})
}

// TestProviderVersionDocumentation_FindByVersion tests finding all documentation for a version
// Python reference: test_provider_version_documentation.py::TestProviderVersionDocumentation::test_get_by_provider_version
func TestProviderVersionDocumentation_FindByVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "initial-providers")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "multiple-versions", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.5.0", gpgKey.ID, false, nil)

	// Create multiple documentation items
	doc1DB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "overview", "overview", sqldb.ProviderDocTypeOverview)
	_ = testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "some_resource", "some_resource", sqldb.ProviderDocTypeResource)
	_ = testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "another_resource", "another_resource", sqldb.ProviderDocTypeResource)
	_ = testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "some_guide", "some_guide", sqldb.ProviderDocTypeGuide)

	// Get all documentation for the version
	docs, err := providerRepo.FindDocumentationByVersion(ctx, versionDB.ID)
	require.NoError(t, err)
	assert.Len(t, docs, 4)

	// Verify IDs are ordered
	ids := make([]int, len(docs))
	for i, doc := range docs {
		ids[i] = doc.ID()
	}
	assert.Equal(t, []int{doc1DB.ID, doc1DB.ID + 1, doc1DB.ID + 2, doc1DB.ID + 3}, ids)
}

// TestProviderVersionDocumentation_Search tests searching documentation
// Python reference: test_provider_version_documentation.py::TestProviderVersionDocumentation::test_search
func TestProviderVersionDocumentation_Search(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "initial-providers")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "multiple-versions", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.5.0", gpgKey.ID, false, nil)

	overviewDB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "overview", "overview", sqldb.ProviderDocTypeOverview)
	_ = testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "some_resource", "some_resource", sqldb.ProviderDocTypeResource)

	t.Run("Search by correct category and slug", func(t *testing.T) {
		docs, err := providerRepo.SearchDocumentation(ctx, versionDB.ID, "overview", "overview", "hcl")
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, overviewDB.ID, docs[0].ID())
	})

	t.Run("Wrong type returns empty", func(t *testing.T) {
		docs, err := providerRepo.SearchDocumentation(ctx, versionDB.ID, "data-source", "overview", "hcl")
		require.NoError(t, err)
		assert.Empty(t, docs)
	})

	t.Run("Non-existent title returns empty", func(t *testing.T) {
		docs, err := providerRepo.SearchDocumentation(ctx, versionDB.ID, "resource", "some-non-existent", "hcl")
		require.NoError(t, err)
		assert.Empty(t, docs)
	})

	t.Run("Wrong language returns empty", func(t *testing.T) {
		docs, err := providerRepo.SearchDocumentation(ctx, versionDB.ID, "resource", "overview", "python")
		require.NoError(t, err)
		assert.Empty(t, docs)
	})
}

// TestProviderVersionDocumentation_Properties tests various property accessors
// Python reference: test_provider_version_documentation.py (various property tests)
func TestProviderVersionDocumentation_Properties(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-doc-props")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)

	t.Run("Title property with title set", func(t *testing.T) {
		title := "unittest-title"
		docDB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "unittest-name", "unittest-name", sqldb.ProviderDocTypeGuide)
		db.DB.Model(&docDB).Update("title", title)

		doc, err := providerRepo.FindDocumentationByID(ctx, docDB.ID)
		require.NoError(t, err)
		assert.Equal(t, title, doc.GetDisplayTitle())
	})

	t.Run("Title property derived from name when title is nil", func(t *testing.T) {
		docDB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "unittest-name.md", "unittest-name_md", sqldb.ProviderDocTypeGuide)
		// Title is nil in CreateProviderVersionDocumentation

		doc, err := providerRepo.FindDocumentationByID(ctx, docDB.ID)
		require.NoError(t, err)
		assert.Equal(t, "unittest-name", doc.GetDisplayTitle())
	})

	t.Run("Category/DocumentationType property", func(t *testing.T) {
		docDB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "some_resource", "some_resource", sqldb.ProviderDocTypeResource)

		doc, err := providerRepo.FindDocumentationByID(ctx, docDB.ID)
		require.NoError(t, err)
		assert.Equal(t, "resource", doc.DocumentationType())
	})

	t.Run("Language property", func(t *testing.T) {
		docDB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "some_resource", "some_resource", sqldb.ProviderDocTypeResource)

		doc, err := providerRepo.FindDocumentationByID(ctx, docDB.ID)
		require.NoError(t, err)
		assert.Equal(t, "hcl", doc.Language())
	})

	t.Run("Filename property", func(t *testing.T) {
		docDB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "some_resource", "some_resource", sqldb.ProviderDocTypeResource)

		doc, err := providerRepo.FindDocumentationByID(ctx, docDB.ID)
		require.NoError(t, err)
		assert.Equal(t, "docs/some_resource", doc.Filename())
	})

	t.Run("Slug property", func(t *testing.T) {
		docDB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "some_resource", "some_resource", sqldb.ProviderDocTypeResource)

		doc, err := providerRepo.FindDocumentationByID(ctx, docDB.ID)
		require.NoError(t, err)
		assert.Equal(t, "some_resource", doc.Slug())
	})
}

// TestProviderVersionDocumentation_Content tests content handling
// Python reference: test_provider_version_documentation.py::test_get_content
func TestProviderVersionDocumentation_Content(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()
	providerRepo := providerrepo.NewProviderRepository(db.DB)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-doc-content")
	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "test-key-id")

	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierCommunity, &category.ID)
	versionDB := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, nil)

	t.Run("Content with markdown", func(t *testing.T) {
		content := []byte("# Hi!")
		docDB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "test-doc", "test-doc", sqldb.ProviderDocTypeResource)
		db.DB.Model(&docDB).Update("content", content)

		doc, err := providerRepo.FindDocumentationByID(ctx, docDB.ID)
		require.NoError(t, err)
		assert.Equal(t, content, doc.Content())
	})

	t.Run("Empty content returns empty bytes", func(t *testing.T) {
		content := []byte{}
		docDB := testutils.CreateProviderVersionDocumentation(t, db, versionDB.ID, "test-doc2", "test-doc2", sqldb.ProviderDocTypeResource)
		db.DB.Model(&docDB).Update("content", content)

		doc, err := providerRepo.FindDocumentationByID(ctx, docDB.ID)
		require.NoError(t, err)
		assert.Equal(t, content, doc.Content())
	})
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
