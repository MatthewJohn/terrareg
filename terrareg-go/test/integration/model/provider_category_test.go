package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestProviderCategory_Create tests creating provider categories
func TestProviderCategory_Create(t *testing.T) {
	t.Run("with name", func(t *testing.T) {
		db := testutils.SetupTestDatabase(t)
		defer testutils.CleanupTestDatabase(t, db)

		// Clean provider_category table
		db.DB.Exec("DELETE FROM provider_category")

		name := "Test Provider Category"
		slug := "test-provider-category"
		userSelectable := true

		category := sqldb.ProviderCategoryDB{
			Name:           &name,
			Slug:           slug,
			UserSelectable: userSelectable,
		}

		err := db.DB.Create(&category).Error
		require.NoError(t, err)

		assert.NotZero(t, category.ID)
		assert.Equal(t, name, *category.Name)
		assert.Equal(t, slug, category.Slug)
		assert.Equal(t, userSelectable, category.UserSelectable)
	})

	t.Run("without name", func(t *testing.T) {
		db := testutils.SetupTestDatabase(t)
		defer testutils.CleanupTestDatabase(t, db)

		// Clean provider_category table
		db.DB.Exec("DELETE FROM provider_category")

		slug := "test-provider-category-no-name"

		category := sqldb.ProviderCategoryDB{
			Name:           nil,
			Slug:           slug,
			UserSelectable: true, // Use default value from DB
		}

		err := db.DB.Create(&category).Error
		require.NoError(t, err)

		assert.NotZero(t, category.ID)
		assert.Nil(t, category.Name)
		assert.Equal(t, slug, category.Slug)
		assert.True(t, category.UserSelectable) // DB default is true
	})
}

// TestProviderCategory_GetByPK tests retrieving by ID
func TestProviderCategory_GetByPK(t *testing.T) {
	testCases := []struct {
		name           string
		exists         bool
		userSelectable bool
	}{
		{"exists - user selectable", true, true},
		{"exists - not user selectable", true, false},
		{"not exists", false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			// Clean provider_category table
			db.DB.Exec("DELETE FROM provider_category")

			testID := 564341

			if tc.exists {
				name := "Test Category"
				// For the not user selectable case, use raw SQL to ensure false is set
				if !tc.userSelectable {
					err := db.DB.Exec("INSERT INTO provider_category (id, name, slug, user_selectable) VALUES (?, ?, ?, ?)",
						testID, name, "test-category", false).Error
					require.NoError(t, err)
				} else {
					category := sqldb.ProviderCategoryDB{
						Name:           &name,
						Slug:           "test-category",
						UserSelectable: tc.userSelectable,
					}
					// Use specific ID
					category.ID = testID
					err := db.DB.Create(&category).Error
					require.NoError(t, err)
				}
			}

			// Query by ID
			var category sqldb.ProviderCategoryDB
			err := db.DB.First(&category, testID).Error

			if tc.exists {
				require.NoError(t, err)
				assert.Equal(t, testID, category.ID)
				assert.Equal(t, "Test Category", *category.Name)
				assert.Equal(t, tc.userSelectable, category.UserSelectable)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestProviderCategory_UniqueSlug tests that slugs must be unique
func TestProviderCategory_UniqueSlug(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean provider_category table
	db.DB.Exec("DELETE FROM provider_category")

	// Create first category
	name1 := "Category One"
	category1 := sqldb.ProviderCategoryDB{
		Name:           &name1,
		Slug:           "duplicate-slug",
		UserSelectable: true,
	}

	err := db.DB.Create(&category1).Error
	require.NoError(t, err)

	// Try to create duplicate - should fail due to unique index
	name2 := "Category Two"
	category2 := sqldb.ProviderCategoryDB{
		Name:           &name2,
		Slug:           "duplicate-slug",
		UserSelectable: false,
	}

	err = db.DB.Create(&category2).Error
	assert.Error(t, err, "Expected error for duplicate slug")
}

// TestProviderCategory_Update tests updating provider categories
func TestProviderCategory_Update(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean provider_category table
	db.DB.Exec("DELETE FROM provider_category")

	// Create category
	name := "Original Name"
	category := sqldb.ProviderCategoryDB{
		Name:           &name,
		Slug:           "test-category",
		UserSelectable: true,
	}

	err := db.DB.Create(&category).Error
	require.NoError(t, err)

	// Update category
	newName := "Updated Name"
	category.Name = &newName
	category.UserSelectable = false
	err = db.DB.Save(&category).Error
	require.NoError(t, err)

	// Verify updated
	var updated sqldb.ProviderCategoryDB
	err = db.DB.First(&updated, category.ID).Error
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", *updated.Name)
	assert.Equal(t, false, updated.UserSelectable)
}

// TestProviderCategory_Delete tests deleting provider categories
func TestProviderCategory_Delete(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Clean provider_category table
	db.DB.Exec("DELETE FROM provider_category")

	// Create category
	name := "To Delete"
	category := sqldb.ProviderCategoryDB{
		Name:           &name,
		Slug:           "to-delete",
		UserSelectable: true,
	}

	err := db.DB.Create(&category).Error
	require.NoError(t, err)
	assert.NotZero(t, category.ID)

	// Delete category
	err = db.DB.Delete(&category).Error
	require.NoError(t, err)

	// Verify deleted
	var count int64
	err = db.DB.Model(&sqldb.ProviderCategoryDB{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestProviderCategory_DomainModel tests the domain model
func TestProviderCategory_DomainModel(t *testing.T) {
	t.Run("create with valid slug", func(t *testing.T) {
		category, err := model.NewProviderCategory("test-category", true)
		assert.NoError(t, err)
		assert.NotNil(t, category)
		assert.Equal(t, "test-category", category.Slug())
		assert.True(t, category.UserSelectable())
	})

	t.Run("create with invalid slug", func(t *testing.T) {
		category, err := model.NewProviderCategory("", true)
		assert.Error(t, err)
		assert.Nil(t, category)
	})

	t.Run("reconstruct from persistence", func(t *testing.T) {
		name := "Test Category"
		category := model.ReconstructProviderCategory(123, &name, "test-category", true, testutils.Now(), testutils.Now())
		assert.NotNil(t, category)
		assert.Equal(t, 123, category.ID())
		assert.Equal(t, "Test Category", *category.Name())
		assert.Equal(t, "test-category", category.Slug())
		assert.True(t, category.UserSelectable())
	})

	t.Run("update name", func(t *testing.T) {
		category, _ := model.NewProviderCategory("test-category", true)
		newName := "Updated Name"
		category.UpdateName(&newName)
		assert.Equal(t, "Updated Name", *category.Name())
	})

	t.Run("update user selectable", func(t *testing.T) {
		category, _ := model.NewProviderCategory("test-category", true)
		category.UpdateUserSelectable(false)
		assert.False(t, category.UserSelectable())
	})

	t.Run("get display name with name", func(t *testing.T) {
		name := "Display Name"
		category := model.ReconstructProviderCategory(1, &name, "test-category", true, testutils.Now(), testutils.Now())
		assert.Equal(t, "Display Name", category.GetDisplayName())
	})

	t.Run("get display name without name", func(t *testing.T) {
		category := model.ReconstructProviderCategory(1, nil, "test-category", true, testutils.Now(), testutils.Now())
		assert.Equal(t, "test-category", category.GetDisplayName())
	})
}
