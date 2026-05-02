package v2_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestTerraformV2CategoryHandler_NewHandler tests creating a new handler
func TestTerraformV2CategoryHandler_NewHandler(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)

	// Handler should be properly wired with repository
	assert.NotNil(t, cont.TerraformV2CategoryHandler, "Handler should be created")
}

// TestTerraformV2CategoryHandler_HandleListCategories_Empty tests listing categories with no categories
func TestTerraformV2CategoryHandler_HandleListCategories_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Test listing categories when none exist
	req := httptest.NewRequest("GET", "/v2/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 OK with empty data array
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "data")
	data := response["data"].([]interface{})
	assert.Empty(t, data, "Should return empty array when no categories exist")
}

// TestTerraformV2CategoryHandler_HandleListCategories_Success tests listing categories with categories present
func TestTerraformV2CategoryHandler_HandleListCategories_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test categories
	_ = testutils.CreateProviderCategory(t, db, "Database", "database", true)
	_ = testutils.CreateProviderCategory(t, db, "Networking", "networking", true)
	_ = testutils.CreateProviderCategory(t, db, "Hidden Category", "hidden", false) // Not user-selectable

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Test listing categories
	req := httptest.NewRequest("GET", "/v2/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 OK with categories data
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "data")
	data := response["data"].([]interface{})

	// Should only return user-selectable categories (2 out of 3)
	assert.Len(t, data, 2, "Should return only user-selectable categories")

	// Verify first category structure
	cat1 := data[0].(map[string]interface{})
	assert.Equal(t, "categories", cat1["type"])

	attributes := cat1["attributes"].(map[string]interface{})
	assert.NotEmpty(t, attributes["name"])
	assert.NotEmpty(t, attributes["slug"])
	assert.True(t, attributes["user-selectable"].(bool))

	links := cat1["links"].(map[string]interface{})
	assert.Contains(t, links["self"], "/v2/categories/")
}

// TestTerraformV2CategoryHandler_HandleListCategories_OnlyUserSelectable tests that only user-selectable categories are returned
func TestTerraformV2CategoryHandler_HandleListCategories_OnlyUserSelectable(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test categories with mixed user-selectable flag
	_ = testutils.CreateProviderCategory(t, db, "Public Category", "public", true)
	_ = testutils.CreateProviderCategory(t, db, "Private Category", "private", false)
	_ = testutils.CreateProviderCategory(t, db, "Another Public", "public2", true)
	_ = testutils.CreateProviderCategory(t, db, "Another Private", "private2", false)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Test listing categories
	req := httptest.NewRequest("GET", "/v2/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 200 OK
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].([]interface{})

	// Should only return user-selectable categories
	assert.Len(t, data, 2, "Should return only user-selectable categories")

	// Verify all returned categories are user-selectable
	for _, cat := range data {
		attributes := cat.(map[string]interface{})["attributes"].(map[string]interface{})
		assert.True(t, attributes["user-selectable"].(bool), "All returned categories should be user-selectable")
	}
}

// TestTerraformV2CategoryHandler_JSONAPIFormat tests that the response matches JSON:API format
func TestTerraformV2CategoryHandler_JSONAPIFormat(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create a test category
	testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)

	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	req := httptest.NewRequest("GET", "/v2/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify JSON:API structure
	assert.Contains(t, response, "data")
	data := response["data"].([]interface{})
	assert.NotEmpty(t, data)

	category := data[0].(map[string]interface{})

	// Required JSON:API fields
	assert.Contains(t, category, "type")
	assert.Contains(t, category, "id")
	assert.Contains(t, category, "attributes")
	assert.Contains(t, category, "links")

	// Verify type
	assert.Equal(t, "categories", category["type"])

	// Verify attributes contain expected fields
	attributes := category["attributes"].(map[string]interface{})
	assert.Contains(t, attributes, "name")
	assert.Contains(t, attributes, "slug")
	assert.Contains(t, attributes, "user-selectable")

	// Verify links contain self
	links := category["links"].(map[string]interface{})
	assert.Contains(t, links, "self")
	assert.Contains(t, links["self"], "/v2/categories/")
}
