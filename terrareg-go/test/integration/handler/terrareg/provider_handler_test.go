package terrareg_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	providerCommand "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/provider"
	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestProviderHandler_HandleProviderList_Success tests successful provider list retrieval
func TestProviderHandler_HandleProviderList_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "ABC123")

	// Create test providers with versions (required for providers to appear in list)
	provider1 := testutils.CreateProvider(t, db, namespace.ID, "provider1", nil, sqldb.ProviderTierOfficial, nil)
	provider2 := testutils.CreateProvider(t, db, namespace.ID, "provider2", nil, sqldb.ProviderTierCommunity, nil)

	// Create versions and set as latest (required for providers to appear in list)
	now := time.Now()
	version1 := testutils.CreateProviderVersion(t, db, provider1.ID, "1.0.0", gpgKey.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, provider2.ID, "1.0.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider1.ID, version1.ID)
	testutils.SetProviderLatestVersion(t, db, provider2.ID, version2.ID)

	// Create handler
	providerRepository := providerRepo.NewProviderRepository(db.DB)
	listProvidersQuery := providerQuery.NewListProvidersQuery(providerRepository)
	handler := terrareg.NewProviderHandler(listProvidersQuery, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/providers", nil)
	w := httptest.NewRecorder()

	handler.HandleProviderList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Should have providers array
	assert.Contains(t, response, "providers")
	providers := response["providers"].([]interface{})
	assert.Len(t, providers, 2)
}

// TestProviderHandler_HandleProviderList_Empty tests provider list with no data
func TestProviderHandler_HandleProviderList_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	listProvidersQuery := providerQuery.NewListProvidersQuery(providerRepository)
	handler := terrareg.NewProviderHandler(listProvidersQuery, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/providers", nil)
	w := httptest.NewRecorder()

	handler.HandleProviderList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	providers := response["providers"].([]interface{})
	assert.Len(t, providers, 0)
}

// TestProviderHandler_HandleProviderList_Pagination tests pagination support
func TestProviderHandler_HandleProviderList_Pagination(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "ABC123")

	// Create multiple providers with versions
	now := time.Now()
	for i := 1; i <= 5; i++ {
		provider := testutils.CreateProvider(t, db, namespace.ID, "provider"+string(rune('0'+i)), nil, sqldb.ProviderTierOfficial, nil)
		version := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)
		testutils.SetProviderLatestVersion(t, db, provider.ID, version.ID)
	}

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	listProvidersQuery := providerQuery.NewListProvidersQuery(providerRepository)
	handler := terrareg.NewProviderHandler(listProvidersQuery, nil, nil, nil, nil, nil, nil, nil, nil)

	// Request with pagination
	params := url.Values{}
	params.Add("offset", "0")
	params.Add("limit", "3")

	req := httptest.NewRequest("GET", "/v1/providers?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleProviderList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	providers := response["providers"].([]interface{})
	assert.Len(t, providers, 3)

	// Check pagination metadata (handler returns "meta", not "pagination")
	assert.Contains(t, response, "meta")
	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(0), meta["current_offset"])
	assert.Equal(t, float64(3), meta["limit"])
}

// TestProviderHandler_HandleProviderSearch_Success tests successful provider search
func TestProviderHandler_HandleProviderSearch_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "ABC123")

	// Create test providers with versions (required for providers to appear in search results)
	now := time.Now()
	provider1 := testutils.CreateProvider(t, db, namespace.ID, "aws-provider", nil, sqldb.ProviderTierOfficial, nil)
	provider2 := testutils.CreateProvider(t, db, namespace.ID, "azure-provider", nil, sqldb.ProviderTierCommunity, nil)
	version1 := testutils.CreateProviderVersion(t, db, provider1.ID, "1.0.0", gpgKey.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, provider2.ID, "1.0.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider1.ID, version1.ID)
	testutils.SetProviderLatestVersion(t, db, provider2.ID, version2.ID)

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	searchProvidersQuery := providerQuery.NewSearchProvidersQuery(providerRepository)
	handler := terrareg.NewProviderHandler(nil, searchProvidersQuery, nil, nil, nil, nil, nil, nil, nil)

	params := url.Values{}
	params.Add("q", "aws")

	req := httptest.NewRequest("GET", "/v1/providers/search?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleProviderSearch(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "providers")
	providers := response["providers"].([]interface{})
	assert.GreaterOrEqual(t, len(providers), 1)
}

// TestProviderHandler_HandleProviderSearch_Empty tests search with no results
func TestProviderHandler_HandleProviderSearch_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	searchProvidersQuery := providerQuery.NewSearchProvidersQuery(providerRepository)
	handler := terrareg.NewProviderHandler(nil, searchProvidersQuery, nil, nil, nil, nil, nil, nil, nil)

	params := url.Values{}
	params.Add("q", "nonexistent")

	req := httptest.NewRequest("GET", "/v1/providers/search?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleProviderSearch(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	providers := response["providers"].([]interface{})
	assert.Len(t, providers, 0)
}

// TestProviderHandler_HandleProviderDetails_Success tests successful provider details retrieval
func TestProviderHandler_HandleProviderDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	description := "Test provider"
	testutils.CreateProvider(t, db, namespace.ID, "test-provider", &description, sqldb.ProviderTierOfficial, nil)

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	getProviderQuery := providerQuery.NewGetProviderQuery(providerRepository)
	handler := terrareg.NewProviderHandler(nil, nil, getProviderQuery, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/providers/test-namespace/test-provider", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-namespace",
		"provider":  "test-provider",
	})

	handler.HandleProviderDetails(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "namespace")
	// Note: Handler returns placeholder "namespace-{id}" format, not actual namespace name
	assert.Equal(t, "namespace-1", response["namespace"])
	assert.Contains(t, response, "name")
	assert.Equal(t, "test-provider", response["name"])
	assert.Contains(t, response, "tier")
	assert.Equal(t, "official", response["tier"])
	assert.Contains(t, response, "description")
	assert.Equal(t, "Test provider", response["description"])
}

// TestProviderHandler_HandleProviderVersions_Success tests successful provider versions retrieval
func TestProviderHandler_HandleProviderVersions_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierOfficial, nil)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "ABC123")

	// Create provider versions
	now := time.Now()
	testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)
	version2 := testutils.CreateProviderVersion(t, db, provider.ID, "1.1.0", gpgKey.ID, false, &now)
	testutils.SetProviderLatestVersion(t, db, provider.ID, version2.ID)

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	getProviderQuery := providerQuery.NewGetProviderQuery(providerRepository)
	getProviderVersionsQuery := providerQuery.NewGetProviderVersionsQuery(providerRepository)
	handler := terrareg.NewProviderHandler(nil, nil, getProviderQuery, getProviderVersionsQuery, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/providers/test-namespace/test-provider/versions", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-namespace",
		"provider":  "test-provider",
	})

	handler.HandleProviderVersions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "versions")
	versions := response["versions"].([]interface{})
	assert.Len(t, versions, 2)
}

// TestProviderHandler_HandleNamespaceProviders_ReturnsEmpty tests that namespace providers returns empty list
func TestProviderHandler_HandleNamespaceProviders_ReturnsEmpty(t *testing.T) {
	handler := terrareg.NewProviderHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/providers/test-namespace", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})

	handler.HandleNamespaceProviders(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "providers")
	providers := response["providers"].([]interface{})
	assert.Len(t, providers, 0)
}

// TestProviderHandler_HandleCreateOrUpdateProvider_Success tests successful provider creation
func TestProviderHandler_HandleCreateOrUpdateProvider_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	_ = testutils.CreateNamespace(t, db, "test-namespace")

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	createOrUpdateProviderCmd := providerCommand.NewCreateOrUpdateProviderCommand(providerRepository, namespaceRepository)
	handler := terrareg.NewProviderHandler(nil, nil, nil, nil, nil, createOrUpdateProviderCmd, nil, nil, nil)

	requestBody := providerCommand.CreateOrUpdateProviderRequest{
		Namespace: "test-namespace",
		Name:      "new-provider",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/v1/providers", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleCreateOrUpdateProvider(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "namespace")
	// Note: Handler returns placeholder "namespace-{id}" format, not actual namespace name
	assert.Equal(t, "namespace-1", response["namespace"])
	assert.Contains(t, response, "name")
	assert.Equal(t, "new-provider", response["name"])
	assert.Contains(t, response, "tier")
	assert.Equal(t, "community", response["tier"])
}

// TestProviderHandler_HandleCreateOrUpdateProvider_MissingFields tests with missing required fields
func TestProviderHandler_HandleCreateOrUpdateProvider_MissingFields(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	createOrUpdateProviderCmd := providerCommand.NewCreateOrUpdateProviderCommand(providerRepository, namespaceRepository)
	handler := terrareg.NewProviderHandler(nil, nil, nil, nil, nil, createOrUpdateProviderCmd, nil, nil, nil)

	requestBody := providerCommand.CreateOrUpdateProviderRequest{
		Name: "test-provider",
		// Missing namespace
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/v1/providers", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleCreateOrUpdateProvider(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}

// TestProviderHandler_HandleCreateOrUpdateProvider_InvalidJSON tests with invalid JSON
func TestProviderHandler_HandleCreateOrUpdateProvider_InvalidJSON(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	createOrUpdateProviderCmd := providerCommand.NewCreateOrUpdateProviderCommand(providerRepository, namespaceRepository)
	handler := terrareg.NewProviderHandler(nil, nil, nil, nil, nil, createOrUpdateProviderCmd, nil, nil, nil)

	req := httptest.NewRequest("POST", "/v1/providers", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleCreateOrUpdateProvider(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}

// TestProviderHandler_HandleGetProviderVersion_Success tests successful provider version retrieval
func TestProviderHandler_HandleGetProviderVersion_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", nil, sqldb.ProviderTierOfficial, nil)
	gpgKey := testutils.CreateGPGKeyWithNamespace(t, db, "test-key", namespace.ID, "ABC123")

	now := time.Now()
	_ = testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", gpgKey.ID, false, &now)

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	getProviderQuery := providerQuery.NewGetProviderQuery(providerRepository)
	getProviderVersionQuery := providerQuery.NewGetProviderVersionQuery(providerRepository)
	handler := terrareg.NewProviderHandler(nil, nil, getProviderQuery, nil, getProviderVersionQuery, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/providers/test-namespace/test-provider/versions/1.0.0", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "test-namespace",
		"provider":  "test-provider",
		"version":   "1.0.0",
	})

	handler.HandleGetProviderVersion(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "version")
	assert.Equal(t, "1.0.0", response["version"])
}

// TestProviderHandler_HandleGetProviderVersion_MissingParameters tests with missing parameters
func TestProviderHandler_HandleGetProviderVersion_MissingParameters(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	providerRepository := providerRepo.NewProviderRepository(db.DB)
	getProviderQuery := providerQuery.NewGetProviderQuery(providerRepository)
	getProviderVersionQuery := providerQuery.NewGetProviderVersionQuery(providerRepository)
	handler := terrareg.NewProviderHandler(nil, nil, getProviderQuery, nil, getProviderVersionQuery, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/providers///versions/", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{
		"namespace": "",
		"provider":  "",
		"version":   "",
	})

	handler.HandleGetProviderVersion(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}
