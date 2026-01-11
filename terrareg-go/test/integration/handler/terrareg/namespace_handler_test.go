package terrareg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	namespaceQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/namespace"
	namespaceCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/namespace"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	namespaceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestNamespaceHandler_HandleNamespaceList_Success tests successful namespace list retrieval
func TestNamespaceHandler_HandleNamespaceList_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespaces
	testutils.CreateNamespace(t, db, "namespace1")
	testutils.CreateNamespace(t, db, "namespace2")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	listNamespacesQuery := moduleQuery.NewListNamespacesQuery(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(listNamespacesQuery, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces", nil)
	w := httptest.NewRecorder()

	handler.HandleNamespaceList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// For array responses, unmarshal directly
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON array")
	assert.Len(t, response, 2)
}

// TestNamespaceHandler_HandleNamespaceList_Empty tests namespace list with no data
func TestNamespaceHandler_HandleNamespaceList_Empty(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler with no data
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	listNamespacesQuery := moduleQuery.NewListNamespacesQuery(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(listNamespacesQuery, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces", nil)
	w := httptest.NewRecorder()

	handler.HandleNamespaceList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// For array responses, unmarshal directly
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON array")
	assert.Len(t, response, 0)
}

// TestNamespaceHandler_HandleNamespaceList_WithPagination tests pagination support
func TestNamespaceHandler_HandleNamespaceList_WithPagination(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespaces
	testutils.CreateNamespace(t, db, "namespace1")
	testutils.CreateNamespace(t, db, "namespace2")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	listNamespacesQuery := moduleQuery.NewListNamespacesQuery(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(listNamespacesQuery, nil, nil, nil, nil)

	// Request with pagination
	params := url.Values{}
	params.Add("limit", "10")
	params.Add("offset", "0")

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces?"+params.Encode(), nil)
	w := httptest.NewRecorder()

	handler.HandleNamespaceList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// With pagination, should return wrapped object with "namespaces" key
	assert.Contains(t, response, "namespaces")

	namespaces := response["namespaces"].([]interface{})
	assert.Len(t, namespaces, 2)
}

// TestNamespaceHandler_HandleNamespaceList_MultipleNamespaces tests with multiple namespaces
func TestNamespaceHandler_HandleNamespaceList_MultipleNamespaces(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create multiple namespaces
	for i := 1; i <= 5; i++ {
		testutils.CreateNamespace(t, db, "namespace"+string(rune('0'+i)))
	}

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	listNamespacesQuery := moduleQuery.NewListNamespacesQuery(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(listNamespacesQuery, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces", nil)
	w := httptest.NewRecorder()

	handler.HandleNamespaceList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// For array responses, unmarshal directly
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON array")
	assert.Len(t, response, 5)
}

// TestNamespaceHandler_HandleNamespaceDetails_Success tests successful namespace details retrieval
func TestNamespaceHandler_HandleNamespaceDetails_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test namespace
	testutils.CreateNamespace(t, db, "test-namespace")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	namespaceDetailsQuery := namespaceQuery.NewNamespaceDetailsQuery(namespaceRepository, namespaceSvc)
	handler := terrareg.NewNamespaceHandler(nil, nil, nil, nil, namespaceDetailsQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces/test-namespace", nil)
	w := httptest.NewRecorder()

	// Add Chi context for path parameter
	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test-namespace"})

	handler.HandleNamespaceDetails(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "name")
	assert.Equal(t, "test-namespace", response["name"])
}

// TestNamespaceHandler_HandleNamespaceDetails_NotFound tests with non-existent namespace
func TestNamespaceHandler_HandleNamespaceDetails_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	namespaceDetailsQuery := namespaceQuery.NewNamespaceDetailsQuery(namespaceRepository, namespaceSvc)
	handler := terrareg.NewNamespaceHandler(nil, nil, nil, nil, namespaceDetailsQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces/nonexistent", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "nonexistent"})

	handler.HandleNamespaceDetails(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Equal(t, map[string]interface{}{}, response)
}

// TestNamespaceHandler_HandleNamespaceDetails_MissingParameter tests with missing namespace parameter
func TestNamespaceHandler_HandleNamespaceDetails_MissingParameter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	namespaceSvc := namespaceService.NewNamespaceService(testutils.CreateTestDomainConfig(t))
	namespaceDetailsQuery := namespaceQuery.NewNamespaceDetailsQuery(namespaceRepository, namespaceSvc)
	handler := terrareg.NewNamespaceHandler(nil, nil, nil, nil, namespaceDetailsQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/namespaces/", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": ""})

	handler.HandleNamespaceDetails(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
	assert.Contains(t, response["error"].(string), "namespace is required")
}

// TestNamespaceHandler_HandleNamespaceCreate_Success tests successful namespace creation
func TestNamespaceHandler_HandleNamespaceCreate_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	createNamespaceCmd := namespaceCmd.NewCreateNamespaceCommand(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(nil, createNamespaceCmd, nil, nil, nil)

	// Create request body
	displayName := "New Namespace"
	requestBody := dto.NamespaceCreateRequest{
		Name:        "new-namespace",
		DisplayName: &displayName,
		Type:        "NONE",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/v1/terrareg/namespaces", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleNamespaceCreate(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "name")
	assert.Equal(t, "new-namespace", response["name"])
	assert.Contains(t, response, "display_name")
}

// TestNamespaceHandler_HandleNamespaceCreate_InvalidJSON tests with invalid JSON
func TestNamespaceHandler_HandleNamespaceCreate_InvalidJSON(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	createNamespaceCmd := namespaceCmd.NewCreateNamespaceCommand(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(nil, createNamespaceCmd, nil, nil, nil)

	req := httptest.NewRequest("POST", "/v1/terrareg/namespaces", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandleNamespaceCreate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}

// TestNamespaceHandler_HandleNamespaceDelete_Success tests successful namespace deletion
func TestNamespaceHandler_HandleNamespaceDelete_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace to delete
	_ = testutils.CreateNamespace(t, db, "delete-me")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	deleteNamespaceCmd := namespaceCmd.NewDeleteNamespaceCommand(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(nil, nil, nil, deleteNamespaceCmd, nil)

	req := httptest.NewRequest("DELETE", "/v1/terrareg/namespaces/delete-me", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "delete-me"})

	handler.HandleNamespaceDelete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Equal(t, map[string]interface{}{}, response)

	// Verify namespace was deleted
	namespaces, err := namespaceRepository.List(requireContext(t))
	require.NoError(t, err)
	assert.Empty(t, namespaces)
}

// TestNamespaceHandler_HandleNamespaceDelete_NotFound tests deleting non-existent namespace
func TestNamespaceHandler_HandleNamespaceDelete_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	deleteNamespaceCmd := namespaceCmd.NewDeleteNamespaceCommand(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(nil, nil, nil, deleteNamespaceCmd, nil)

	req := httptest.NewRequest("DELETE", "/v1/terrareg/namespaces/nonexistent", nil)
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "nonexistent"})

	handler.HandleNamespaceDelete(w, req)

	// Should return error for non-existent namespace
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}

// TestNamespaceHandler_HandleNamespaceUpdate_Success tests successful namespace update
func TestNamespaceHandler_HandleNamespaceUpdate_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace
	testutils.CreateNamespace(t, db, "update-namespace")

	// Create handler
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	updateNamespaceCmd := namespaceCmd.NewUpdateNamespaceCommand(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(nil, nil, updateNamespaceCmd, nil, nil)

	// Create request body
	displayName := "Updated Display Name"
	requestBody := dto.NamespaceUpdateRequest{
		DisplayName: &displayName,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/v1/terrareg/namespaces/update-namespace", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "update-namespace"})

	handler.HandleNamespaceUpdate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Contains(t, response, "name")
	assert.Equal(t, "update-namespace", response["name"])
	assert.Contains(t, response, "display_name")
}

// TestNamespaceHandler_HandleNamespaceUpdate_MissingParameter tests with missing namespace parameter
func TestNamespaceHandler_HandleNamespaceUpdate_MissingParameter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	updateNamespaceCmd := namespaceCmd.NewUpdateNamespaceCommand(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(nil, nil, updateNamespaceCmd, nil, nil)

	displayName := "Test"
	requestBody := dto.NamespaceUpdateRequest{
		DisplayName: &displayName,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/v1/terrareg/namespaces/", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": ""})

	handler.HandleNamespaceUpdate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}

// TestNamespaceHandler_HandleNamespaceUpdate_InvalidJSON tests with invalid JSON
func TestNamespaceHandler_HandleNamespaceUpdate_InvalidJSON(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	updateNamespaceCmd := namespaceCmd.NewUpdateNamespaceCommand(namespaceRepository)
	handler := terrareg.NewNamespaceHandler(nil, nil, updateNamespaceCmd, nil, nil)

	req := httptest.NewRequest("POST", "/v1/terrareg/namespaces/test", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	req = testutils.AddChiContext(t, req, map[string]string{"namespace": "test"})

	handler.HandleNamespaceUpdate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "error")
}

func requireContext(t *testing.T) context.Context {
	ctx := context.Background()
	return ctx
}
