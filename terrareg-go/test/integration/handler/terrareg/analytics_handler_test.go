package terrareg_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	analyticsQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/analytics"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	namespaceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	analyticsRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/analytics"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestAnalyticsHandler_HandleGlobalStatsSummary_Success tests successful global stats retrieval
func TestAnalyticsHandler_HandleGlobalStatsSummary_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create analytics data
	timestamp := time.Now()
	testutils.CreateAnalyticsData(t, db, moduleVersion.ID, 5, timestamp)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	globalStatsQuery := analyticsQuery.NewGlobalStatsQuery(namespaceRepository, moduleProviderRepository, analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(globalStatsQuery, nil, nil, nil, nil, nil, nil)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/stats_summary", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleGlobalStatsSummary(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "namespaces")
	assert.Contains(t, response, "modules")
	assert.Contains(t, response, "module_versions")
	assert.Contains(t, response, "downloads")
	assert.Equal(t, float64(1), response["namespaces"])
	assert.Equal(t, float64(1), response["modules"])
	assert.Equal(t, float64(1), response["module_versions"])
	assert.Equal(t, float64(5), response["downloads"])
}

// TestAnalyticsHandler_HandleGlobalStatsSummary_Empty tests global stats with no data
func TestAnalyticsHandler_HandleGlobalStatsSummary_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler with no data
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	globalStatsQuery := analyticsQuery.NewGlobalStatsQuery(namespaceRepository, moduleProviderRepository, analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(globalStatsQuery, nil, nil, nil, nil, nil, nil)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/stats_summary", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleGlobalStatsSummary(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Equal(t, float64(0), response["namespaces"])
	assert.Equal(t, float64(0), response["modules"])
	assert.Equal(t, float64(0), response["module_versions"])
	assert.Equal(t, float64(0), response["downloads"])
}

// TestAnalyticsHandler_HandleGlobalStatsSummary_Structure tests response structure
func TestAnalyticsHandler_HandleGlobalStatsSummary_Structure(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	globalStatsQuery := analyticsQuery.NewGlobalStatsQuery(namespaceRepository, moduleProviderRepository, analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(globalStatsQuery, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/stats_summary", nil)
	w := httptest.NewRecorder()

	handler.HandleGlobalStatsSummary(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Verify all expected fields are present
	assert.Contains(t, response, "namespaces")
	assert.Contains(t, response, "modules")
	assert.Contains(t, response, "module_versions")
	assert.Contains(t, response, "downloads")

	// Verify types
	assert.IsType(t, float64(0), response["namespaces"])
	assert.IsType(t, float64(0), response["modules"])
	assert.IsType(t, float64(0), response["module_versions"])
	assert.IsType(t, float64(0), response["downloads"])
}

// TestAnalyticsHandler_HandleMostRecentlyPublished_Success tests retrieving most recently published module
func TestAnalyticsHandler_HandleMostRecentlyPublished_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create analytics data
	timestamp := time.Now()
	testutils.CreateAnalyticsData(t, db, moduleVersion.ID, 10, timestamp)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getMostRecentlyPublishedQuery := analyticsQuery.NewGetMostRecentlyPublishedQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, nil, nil, getMostRecentlyPublishedQuery, nil, nil)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/most_recently_published_module_version", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleMostRecentlyPublished(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "id")
	assert.Contains(t, response, "namespace")
	assert.Contains(t, response, "name")
	assert.Contains(t, response, "provider")
	assert.Contains(t, response, "version")
	assert.Contains(t, response, "downloads")
	assert.Equal(t, "test-namespace", response["namespace"])
	assert.Equal(t, "test-module", response["name"])
	assert.Equal(t, "aws", response["provider"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, float64(10), response["downloads"])
}

// TestAnalyticsHandler_HandleMostRecentlyPublished_NotFound tests when no modules exist
func TestAnalyticsHandler_HandleMostRecentlyPublished_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler with no data
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getMostRecentlyPublishedQuery := analyticsQuery.NewGetMostRecentlyPublishedQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, nil, nil, getMostRecentlyPublishedQuery, nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/most_recently_published_module_version", nil)
	w := httptest.NewRecorder()

	handler.HandleMostRecentlyPublished(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Equal(t, map[string]interface{}{}, response)
}

// TestAnalyticsHandler_HandleMostRecentlyPublished_Multiple tests with multiple modules
func TestAnalyticsHandler_HandleMostRecentlyPublished_Multiple(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "module1", "aws")
	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "module2", "aws")

	// Create first module
	moduleVersion1 := testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")

	// Create second module with a different timestamp
	time.Sleep(time.Millisecond) // Ensure different timestamp
	moduleVersion2 := testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "2.0.0")

	// Create analytics data for both
	timestamp := time.Now()
	testutils.CreateAnalyticsData(t, db, moduleVersion1.ID, 5, timestamp)
	testutils.CreateAnalyticsData(t, db, moduleVersion2.ID, 3, timestamp)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getMostRecentlyPublishedQuery := analyticsQuery.NewGetMostRecentlyPublishedQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, nil, nil, getMostRecentlyPublishedQuery, nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/most_recently_published_module_version", nil)
	w := httptest.NewRecorder()

	handler.HandleMostRecentlyPublished(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	// Should return the most recently published module (module2)
	assert.Equal(t, "module2", response["name"])
	assert.Equal(t, "2.0.0", response["version"])
}

// TestAnalyticsHandler_HandleMostDownloadedThisWeek_Success tests retrieving most downloaded module this week
func TestAnalyticsHandler_HandleMostDownloadedThisWeek_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create analytics data within this week
	timestamp := time.Now()
	testutils.CreateAnalyticsData(t, db, moduleVersion.ID, 15, timestamp)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getMostDownloadedThisWeekQuery := analyticsQuery.NewGetMostDownloadedThisWeekQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, nil, nil, nil, getMostDownloadedThisWeekQuery, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/most_downloaded_module_provider_this_week", nil)
	w := httptest.NewRecorder()

	handler.HandleMostDownloadedThisWeek(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "namespace")
	assert.Contains(t, response, "module")
	assert.Contains(t, response, "provider")
	assert.Contains(t, response, "downloads")
	assert.Equal(t, "test-namespace", response["namespace"])
	assert.Equal(t, "test-module", response["module"])
	assert.Equal(t, "aws", response["provider"])
	assert.Equal(t, float64(15), response["downloads"])
}

// TestAnalyticsHandler_HandleMostDownloadedThisWeek_NotFound tests when no downloads this week
func TestAnalyticsHandler_HandleMostDownloadedThisWeek_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getMostDownloadedThisWeekQuery := analyticsQuery.NewGetMostDownloadedThisWeekQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, nil, nil, nil, getMostDownloadedThisWeekQuery, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/most_downloaded_module_provider_this_week", nil)
	w := httptest.NewRecorder()

	handler.HandleMostDownloadedThisWeek(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Equal(t, map[string]interface{}{}, response)
}

// TestAnalyticsHandler_HandleMostDownloadedThisWeek_Multiple tests with multiple modules
func TestAnalyticsHandler_HandleMostDownloadedThisWeek_Multiple(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "module1", "aws")
	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "module2", "aws")

	moduleVersion1 := testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")
	moduleVersion2 := testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "1.0.0")

	timestamp := time.Now()
	// Create 5 downloads for module1
	for i := 0; i < 5; i++ {
		testutils.CreateAnalyticsDataWithDetails(t, db, moduleVersion1.ID, timestamp, "1.5.0", "token", "auth", "prod", "test-namespace", "module1", "aws")
	}
	// Create 10 downloads for module2
	for i := 0; i < 10; i++ {
		testutils.CreateAnalyticsDataWithDetails(t, db, moduleVersion2.ID, timestamp, "1.5.0", "token", "auth", "prod", "test-namespace", "module2", "aws")
	}

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getMostDownloadedThisWeekQuery := analyticsQuery.NewGetMostDownloadedThisWeekQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, nil, nil, nil, getMostDownloadedThisWeekQuery, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/most_downloaded_module_provider_this_week", nil)
	w := httptest.NewRecorder()

	handler.HandleMostDownloadedThisWeek(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	// Should return module2 with most downloads
	assert.Equal(t, "module2", response["module"])
	assert.Equal(t, float64(10), response["downloads"])
}

// TestAnalyticsHandler_HandleModuleDownloadsSummary_Success tests successful download summary retrieval
func TestAnalyticsHandler_HandleModuleDownloadsSummary_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	timestamp := time.Now()
	testutils.CreateAnalyticsData(t, db, moduleVersion.ID, 25, timestamp)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getDownloadSummaryQuery := analyticsQuery.NewGetDownloadSummaryQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, getDownloadSummaryQuery, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/modules/test-namespace/test-module/aws/downloads/summary", nil)
	w := httptest.NewRecorder()

	// Act - using testutils to add Chi context
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-namespace",
		"name":      "test-module",
		"provider":  "aws",
	})

	handler.HandleModuleDownloadsSummary(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "type")
	assert.Contains(t, data, "id")
	assert.Contains(t, data, "attributes")
	assert.Equal(t, "module-downloads", data["type"])
	assert.Equal(t, "test-namespace/test-module/aws", data["id"])
	attributes := data["attributes"].(map[string]interface{})
	assert.Equal(t, float64(25), attributes["total"])
}

// TestAnalyticsHandler_HandleModuleDownloadsSummary_NotFound tests with non-existent module
func TestAnalyticsHandler_HandleModuleDownloadsSummary_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getDownloadSummaryQuery := analyticsQuery.NewGetDownloadSummaryQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, getDownloadSummaryQuery, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/modules/nonexistent/module/aws/downloads/summary", nil)
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "nonexistent",
		"name":      "module",
		"provider":  "aws",
	})
	w := httptest.NewRecorder()

	handler.HandleModuleDownloadsSummary(w, req)

	// Should still return 200 with 0 downloads (query creates stats with 0)
	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	data := response["data"].(map[string]interface{})
	attributes := data["attributes"].(map[string]interface{})
	assert.Equal(t, float64(0), attributes["total"])
}

// TestAnalyticsHandler_HandleModuleDownloadsSummary_PathParams tests path parameter handling
func TestAnalyticsHandler_HandleModuleDownloadsSummary_PathParams(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "my-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "my-module", "gcp")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "2.5.0")

	timestamp := time.Now()
	testutils.CreateAnalyticsDataWithDetails(t, db, moduleVersion.ID, timestamp, "1.4.0", "token-123", "auth-456", "dev", "my-namespace", "my-module", "gcp")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getDownloadSummaryQuery := analyticsQuery.NewGetDownloadSummaryQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, getDownloadSummaryQuery, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/modules/my-namespace/my-module/gcp/downloads/summary", nil)
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "my-namespace",
		"name":      "my-module",
		"provider":  "gcp",
	})
	w := httptest.NewRecorder()

	handler.HandleModuleDownloadsSummary(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "my-namespace/my-module/gcp", data["id"])
}

// TestAnalyticsHandler_HandleModuleDownloadsSummary_Format tests JSON:API format
func TestAnalyticsHandler_HandleModuleDownloadsSummary_Format(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "format-ns")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "format-module", "azurerm")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	timestamp := time.Now()
	// Create 100 downloads for format-module
	for i := 0; i < 100; i++ {
		testutils.CreateAnalyticsDataWithDetails(t, db, moduleVersion.ID, timestamp, "1.5.0", "token", "auth", "prod", "format-ns", "format-module", "azurerm")
	}

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getDownloadSummaryQuery := analyticsQuery.NewGetDownloadSummaryQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, getDownloadSummaryQuery, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/modules/format-ns/format-module/azurerm/downloads/summary", nil)
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "format-ns",
		"name":      "format-module",
		"provider":  "azurerm",
	})
	w := httptest.NewRecorder()

	handler.HandleModuleDownloadsSummary(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Verify JSON:API format
	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "type")
	assert.Contains(t, data, "id")
	assert.Contains(t, data, "attributes")
	// Should NOT have "links" or "relationships" for simple response
	attributes := data["attributes"].(map[string]interface{})
	assert.Contains(t, attributes, "total")
	assert.Equal(t, float64(100), attributes["total"])
}

// TestAnalyticsHandler_HandleTokenVersions_Success tests token versions retrieval
func TestAnalyticsHandler_HandleTokenVersions_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	timestamp := time.Now()
	testutils.CreateAnalyticsDataWithDetails(t, db, moduleVersion.ID, timestamp, "1.5.0", "my-token", "auth-token", "production", "test-namespace", "test-module", "aws")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getTokenVersionsQuery := analyticsQuery.NewGetTokenVersionsQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, nil, nil, nil, nil, getTokenVersionsQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/test-namespace/test-module/aws/token_versions", nil)
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-namespace",
		"name":      "test-module",
		"provider":  "aws",
	})
	w := httptest.NewRecorder()

	handler.HandleTokenVersions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.IsType(t, map[string]interface{}{}, response)
	// Should contain the token
	assert.Contains(t, response, "my-token")
	tokenInfo := response["my-token"].(map[string]interface{})
	assert.Contains(t, tokenInfo, "terraform_version")
	assert.Contains(t, tokenInfo, "module_version")
	assert.Equal(t, "1.5.0", tokenInfo["terraform_version"])
	assert.Equal(t, "1.0.0", tokenInfo["module_version"])
}

// TestAnalyticsHandler_HandleTokenVersions_NotFound tests with non-existent module
func TestAnalyticsHandler_HandleTokenVersions_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getTokenVersionsQuery := analyticsQuery.NewGetTokenVersionsQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, nil, nil, nil, nil, getTokenVersionsQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/nonexistent/module/aws/token_versions", nil)
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "nonexistent",
		"name":      "module",
		"provider":  "aws",
	})
	w := httptest.NewRecorder()

	handler.HandleTokenVersions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Equal(t, map[string]interface{}{}, response)
}

// TestAnalyticsHandler_HandleTokenVersions_MissingParams tests with missing path parameters
func TestAnalyticsHandler_HandleTokenVersions_MissingParams(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getTokenVersionsQuery := analyticsQuery.NewGetTokenVersionsQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, nil, nil, nil, nil, getTokenVersionsQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics///token_versions", nil)
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "",
		"name":      "",
		"provider":  "",
	})
	w := httptest.NewRecorder()

	handler.HandleTokenVersions(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "message")
	assert.Equal(t, "Missing required path parameters", response["message"])
}

// TestAnalyticsHandler_HandleTokenVersions_Format tests response format
func TestAnalyticsHandler_HandleTokenVersions_Format(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "format-ns")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "format-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "3.0.0")

	timestamp := time.Now()
	// Create analytics with multiple tokens
	testutils.CreateAnalyticsDataWithDetails(t, db, moduleVersion.ID, timestamp, "1.6.0", "token-1", "auth-1", "prod", "format-ns", "format-module", "aws")
	time.Sleep(time.Millisecond) // Ensure different timestamps
	testutils.CreateAnalyticsDataWithDetails(t, db, moduleVersion.ID, timestamp.Add(time.Minute), "1.7.0", "token-2", "auth-2", "dev", "format-ns", "format-module", "aws")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	getTokenVersionsQuery := analyticsQuery.NewGetTokenVersionsQuery(analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, nil, nil, nil, nil, nil, getTokenVersionsQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/format-ns/format-module/aws/token_versions", nil)
	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "format-ns",
		"name":      "format-module",
		"provider":  "aws",
	})
	w := httptest.NewRecorder()

	handler.HandleTokenVersions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Should have both tokens
	assert.Contains(t, response, "token-1")
	assert.Contains(t, response, "token-2")

	// Verify token structure
	token1Info := response["token-1"].(map[string]interface{})
	assert.Contains(t, token1Info, "terraform_version")
	assert.Contains(t, token1Info, "module_version")
	assert.Contains(t, token1Info, "environment")

	token2Info := response["token-2"].(map[string]interface{})
	assert.Equal(t, "1.7.0", token2Info["terraform_version"])
	assert.Equal(t, "dev", token2Info["environment"])
}

// TestAnalyticsHandler_HandleGlobalUsageStats_Success tests successful global usage stats retrieval
func TestAnalyticsHandler_HandleGlobalUsageStats_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	moduleVersion := testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	timestamp := time.Now()
	testutils.CreateAnalyticsData(t, db, moduleVersion.ID, 7, timestamp)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	globalUsageStatsQuery := analyticsQuery.NewGlobalUsageStatsQuery(moduleProviderRepository, analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, globalUsageStatsQuery, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/usage_stats", nil)
	w := httptest.NewRecorder()

	handler.HandleGlobalUsageStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "module_provider_count")
	assert.Contains(t, response, "module_provider_usage_breakdown_with_auth_token")
	assert.Contains(t, response, "module_provider_usage_count_with_auth_token")
	assert.Contains(t, response, "module_provider_usage_including_empty_auth_token")
	assert.Contains(t, response, "module_provider_usage_count_including_empty_auth_token")
	assert.Equal(t, float64(1), response["module_provider_count"])
	assert.Equal(t, float64(7), response["module_provider_usage_count_with_auth_token"])
}

// TestAnalyticsHandler_HandleGlobalUsageStats_Empty tests usage stats with no data
func TestAnalyticsHandler_HandleGlobalUsageStats_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	globalUsageStatsQuery := analyticsQuery.NewGlobalUsageStatsQuery(moduleProviderRepository, analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, globalUsageStatsQuery, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/usage_stats", nil)
	w := httptest.NewRecorder()

	handler.HandleGlobalUsageStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Equal(t, float64(0), response["module_provider_count"])
	assert.Equal(t, float64(0), response["module_provider_usage_count_with_auth_token"])
}

// TestAnalyticsHandler_HandleGlobalUsageStats_Structure tests response structure
func TestAnalyticsHandler_HandleGlobalUsageStats_Structure(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "struct-ns")
	moduleProvider1 := testutils.CreateModuleProvider(t, db, namespace.ID, "struct-module1", "aws")
	moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace.ID, "struct-module2", "gcp")

	moduleVersion1 := testutils.CreatePublishedModuleVersion(t, db, moduleProvider1.ID, "1.0.0")
	moduleVersion2 := testutils.CreatePublishedModuleVersion(t, db, moduleProvider2.ID, "1.0.0")

	timestamp := time.Now()
	// Create 5 downloads for struct-module1
	for i := 0; i < 5; i++ {
		testutils.CreateAnalyticsDataWithDetails(t, db, moduleVersion1.ID, timestamp, "1.5.0", "token", "auth", "prod", "struct-ns", "struct-module1", "aws")
	}
	// Create 10 downloads for struct-module2
	for i := 0; i < 10; i++ {
		testutils.CreateAnalyticsDataWithDetails(t, db, moduleVersion2.ID, timestamp, "1.5.0", "token", "auth", "prod", "struct-ns", "struct-module2", "gcp")
	}

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	analyticsRepository := analyticsRepo.NewAnalyticsRepository(db.DB, namespaceRepository, namespaceSvc)
	globalUsageStatsQuery := analyticsQuery.NewGlobalUsageStatsQuery(moduleProviderRepository, analyticsRepository)
	handler := terrareg.NewAnalyticsHandler(nil, globalUsageStatsQuery, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/analytics/global/usage_stats", nil)
	w := httptest.NewRecorder()

	handler.HandleGlobalUsageStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Verify structure
	assert.Equal(t, float64(2), response["module_provider_count"])
	assert.Equal(t, float64(15), response["module_provider_usage_count_with_auth_token"])

	// Verify breakdown is a map
	breakdown := response["module_provider_usage_breakdown_with_auth_token"].(map[string]interface{})
	assert.IsType(t, map[string]interface{}{}, breakdown)
	assert.Len(t, breakdown, 2)
}
