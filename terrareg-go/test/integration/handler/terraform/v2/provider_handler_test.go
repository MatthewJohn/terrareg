package v2_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	v2 "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v2"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestTerraformV2ProviderHandler_Integration_HandleProviderDetails_Success tests provider details with real database
func TestTerraformV2ProviderHandler_Integration_HandleProviderDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	testProvider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", stringPtr("Test description"), "community", nil)

	// Create queries and handler
	providerRepository := providerRepo.NewProviderRepository(db.DB)
	getProviderQuery := providerQuery.NewGetProviderQuery(providerRepository)

	handler := v2.NewTerraformV2ProviderHandler(
		getProviderQuery,
		nil,
		nil,
		nil,
		nil,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v2/providers/test-namespace/test-provider", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("provider", "test-provider")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "providers", data["type"])
	assert.Equal(t, float64(testProvider.ID), data["id"])

	attributes := data["attributes"].(map[string]interface{})
	assert.Equal(t, "Test description", attributes["description"])
	assert.Equal(t, "test-namespace/test-provider", attributes["full-name"])
	assert.Equal(t, "test-provider", attributes["name"])
	assert.Equal(t, "test-namespace", attributes["namespace"])
}

// TestTerraformV2ProviderHandler_Integration_HandleProviderDetails_NotFound tests provider not found
func TestTerraformV2ProviderHandler_Integration_HandleProviderDetails_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create queries and handler (no test data)
	providerRepository := providerRepo.NewProviderRepository(db.DB)
	getProviderQuery := providerQuery.NewGetProviderQuery(providerRepository)

	handler := v2.NewTerraformV2ProviderHandler(
		getProviderQuery,
		nil,
		nil,
		nil,
		nil,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v2/providers/nonexistent/provider", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "nonexistent")
	rctx.URLParams.Add("provider", "provider")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// RespondError returns {"message": "..."} not {"error": "..."}
	assert.Contains(t, response["message"], "not found")
}

// TestTerraformV2ProviderHandler_Integration_HandleProviderVersions_Success tests provider versions with real database
func TestTerraformV2ProviderHandler_Integration_HandleProviderVersions_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	testProvider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", stringPtr("Test description"), "community", nil)
	publishedAt := time.Now()
	testProviderVersion1 := testutils.CreateProviderVersion(t, db, testProvider.ID, "1.0.0", 0, false, &publishedAt)
	testProviderVersion2 := testutils.CreateProviderVersion(t, db, testProvider.ID, "2.0.0", 0, false, &publishedAt)

	_ = testProviderVersion1
	_ = testProviderVersion2

	// Create queries and handler
	providerRepository := providerRepo.NewProviderRepository(db.DB)
	getProviderQuery := providerQuery.NewGetProviderQuery(providerRepository)
	getProviderVersionsQuery := providerQuery.NewGetProviderVersionsQuery(providerRepository)

	handler := v2.NewTerraformV2ProviderHandler(
		getProviderQuery,
		getProviderVersionsQuery,
		nil,
		nil,
		nil,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v2/providers/test-namespace/test-provider/versions", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("provider", "test-provider")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleProviderVersions(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, "test-namespace/test-provider", response["id"])
	versions := response["versions"].([]interface{})
	assert.Len(t, versions, 2)
}

// TestTerraformV2ProviderHandler_Integration_HandleProviderDownloadsSummary tests downloads summary endpoint
func TestTerraformV2ProviderHandler_Integration_HandleProviderDownloadsSummary(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	handler := v2.NewTerraformV2ProviderHandler(
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v2/providers/42/downloads/summary", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider_id", "42")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleProviderDownloadsSummary(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure (empty summary as per implementation)
	assert.Equal(t, "42", response["id"])
	downloads := response["downloads"].(map[string]interface{})
	assert.Equal(t, float64(0), downloads["total"])
}

// Helper function to create a string pointer
func stringPtr(s string) *string {
	return &s
}

// TestTerraformV2ProviderHandler_Integration_HandleProviderDetails_WithProviderVersionsInclude tests the include parameter for provider-versions
// Python reference: /app/terrareg/server/api/terraform/v2/provider.py lines 47-53
func TestTerraformV2ProviderHandler_Integration_HandleProviderDetails_WithProviderVersionsInclude(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with provider and versions
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	testProvider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", stringPtr("Test description"), "community", nil)
	publishedAt := time.Now()
	testProviderVersion1 := testutils.CreateProviderVersion(t, db, testProvider.ID, "1.0.0", 0, false, &publishedAt)
	testProviderVersion2 := testutils.CreateProviderVersion(t, db, testProvider.ID, "2.0.0", 0, false, &publishedAt)

	_ = testProviderVersion1
	_ = testProviderVersion2

	// Create queries and handler
	providerRepository := providerRepo.NewProviderRepository(db.DB)
	getProviderQuery := providerQuery.NewGetProviderQuery(providerRepository)
	getProviderVersionsQuery := providerQuery.NewGetProviderVersionsQuery(providerRepository)

	handler := v2.NewTerraformV2ProviderHandler(
		getProviderQuery,
		getProviderVersionsQuery,
		nil,
		nil,
		nil,
	)

	// Create request with include parameter
	req := httptest.NewRequest("GET", "/v2/providers/test-namespace/test-provider?include=provider-versions", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("provider", "test-provider")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify main data structure
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "providers", data["type"])
	assert.Equal(t, float64(testProvider.ID), data["id"])

	// Verify included array exists (JSON arrays unmarshal to []interface{})
	includedRaw, ok := response["included"].([]interface{})
	require.True(t, ok, "included array should be present when include=provider-versions")
	assert.Len(t, includedRaw, 2, "should include both provider versions")

	// Verify first included version structure
	firstVersion := includedRaw[0].(map[string]interface{})
	assert.Equal(t, "provider-versions", firstVersion["type"])
	assert.Equal(t, "2", firstVersion["id"]) // ID as string (strconv.Itoa)
	attributes := firstVersion["attributes"].(map[string]interface{})
	assert.Equal(t, "2.0.0", attributes["version"])
	// Tag will be empty string since we didn't set it - that's okay for this test
	assert.Equal(t, "", attributes["tag"])
	assert.Equal(t, float64(0), attributes["downloads"]) // TODO: analytics integration

	// Verify links structure exists
	links, ok := firstVersion["links"].(map[string]interface{})
	assert.True(t, ok, "links should exist")
	assert.Contains(t, links["self"], "/v2/provider-versions/")
}

// TestTerraformV2ProviderHandler_Integration_HandleProviderDetails_NoInclude tests provider details without include parameter
func TestTerraformV2ProviderHandler_Integration_HandleProviderDetails_NoInclude(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	_ = testutils.CreateProvider(t, db, namespace.ID, "test-provider", stringPtr("Test description"), "community", nil)

	// Create queries and handler
	providerRepository := providerRepo.NewProviderRepository(db.DB)
	getProviderQuery := providerQuery.NewGetProviderQuery(providerRepository)
	getProviderVersionsQuery := providerQuery.NewGetProviderVersionsQuery(providerRepository)

	handler := v2.NewTerraformV2ProviderHandler(
		getProviderQuery,
		getProviderVersionsQuery,
		nil,
		nil,
		nil,
	)

	// Create request without include parameter
	req := httptest.NewRequest("GET", "/v2/providers/test-namespace/test-provider", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("provider", "test-provider")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify main data structure exists
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "providers", data["type"])

	// Verify included array does NOT exist when not requested
	_, ok := response["included"]
	assert.False(t, ok, "included array should not be present when include parameter is not specified")
}

// TestTerraformV2ProviderHandler_Integration_HandleProviderDetails_WithWhitespaceInclude tests include parameter with whitespace
func TestTerraformV2ProviderHandler_Integration_HandleProviderDetails_WithWhitespaceInclude(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	testProvider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", stringPtr("Test description"), "community", nil)
	publishedAt := time.Now()
	testProviderVersion := testutils.CreateProviderVersion(t, db, testProvider.ID, "1.0.0", 0, false, &publishedAt)

	_ = testProviderVersion

	// Create queries and handler
	providerRepository := providerRepo.NewProviderRepository(db.DB)
	getProviderQuery := providerQuery.NewGetProviderQuery(providerRepository)
	getProviderVersionsQuery := providerQuery.NewGetProviderVersionsQuery(providerRepository)

	handler := v2.NewTerraformV2ProviderHandler(
		getProviderQuery,
		getProviderVersionsQuery,
		nil,
		nil,
		nil,
	)

	// Create request with include parameter with whitespace
	req := httptest.NewRequest("GET", "/v2/providers/test-namespace/test-provider?include=provider-versions%2C%20categories", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("provider", "test-provider")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleProviderDetails(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify included array exists with provider versions (JSON arrays unmarshal to []interface{})
	includedRaw, ok := response["included"].([]interface{})
	require.True(t, ok, "included array should be present")
	assert.Len(t, includedRaw, 1, "should include provider version (categories not implemented yet)")
}
