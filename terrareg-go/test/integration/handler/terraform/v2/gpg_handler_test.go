package v2_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gpgkeyQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/gpgkey"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/service"
	gpgkeyRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/gpgkey"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	v2 "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v2"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_Success tests listing GPG keys with real database
func TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "test-source", namespace.ID, "0829C61C7E3")
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "another-source", namespace.ID, "0829C61C7E3")

	// Create repositories and service
	gpgKeyRepository, err := gpgkeyRepo.NewGPGKeyRepository(db.DB)
	require.NoError(t, err)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	gpgKeyService := service.NewGPGKeyService(gpgKeyRepository, namespaceRepository)

	// Create queries
	getNamespaceGPGKeysQuery := gpgkeyQuery.NewGetNamespaceGPGKeysQuery(gpgKeyService)
	getMultipleNamespaceGPGKeysQuery := gpgkeyQuery.NewGetMultipleNamespaceGPGKeysQuery(gpgKeyService)

	// Create handler
	handler := v2.NewTerraformV2GPGHandler(
		nil, // manageGPGKeyCmd - not needed for this test
		getNamespaceGPGKeysQuery,
		getMultipleNamespaceGPGKeysQuery,
		nil, // getGPGKeyQuery - not needed for this test
	)

	// Create request
	req := httptest.NewRequest("GET", "/v2/gpg-keys?filter[namespace]=test-namespace", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListGPGKeys(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
}

// TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_MultipleNamespaces tests listing GPG keys from multiple namespaces
func TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_MultipleNamespaces(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace1 := testutils.CreateNamespace(t, db, "ns1", nil)
	namespace2 := testutils.CreateNamespace(t, db, "ns2", nil)
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "source1", namespace1.ID, "0829C61C7E3")
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "source2", namespace2.ID, "0829C61C7E3")

	// Create repositories and service
	gpgKeyRepository, err := gpgkeyRepo.NewGPGKeyRepository(db.DB)
	require.NoError(t, err)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	gpgKeyService := service.NewGPGKeyService(gpgKeyRepository, namespaceRepository)

	// Create queries
	getMultipleNamespaceGPGKeysQuery := gpgkeyQuery.NewGetMultipleNamespaceGPGKeysQuery(gpgKeyService)

	// Create handler
	handler := v2.NewTerraformV2GPGHandler(
		nil,
		nil,
		getMultipleNamespaceGPGKeysQuery,
		nil,
	)

	// Create request with multiple namespaces
	req := httptest.NewRequest("GET", "/v2/gpg-keys?filter[namespace]=ns1,ns2", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListGPGKeys(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
}

// TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_MissingFilter tests missing filter parameter
func TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_MissingFilter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler with nil dependencies (won't be called)
	handler := v2.NewTerraformV2GPGHandler(nil, nil, nil, nil)

	// Create request without filter parameter
	req := httptest.NewRequest("GET", "/v2/gpg-keys", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListGPGKeys(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "Missing required parameter")
}

// TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_EmptyFilter tests empty filter parameter
func TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_EmptyFilter(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler with nil dependencies (won't be called)
	handler := v2.NewTerraformV2GPGHandler(nil, nil, nil, nil)

	// Create request with empty filter parameter
	reqURL := "/v2/gpg-keys?filter[namespace]="
	req := httptest.NewRequest("GET", reqURL, nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListGPGKeys(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "Missing required parameter")
}

// TestTerraformV2GPGHandler_Integration_HandleGetGPGKey_Success tests getting a specific GPG key
func TestTerraformV2GPGHandler_Integration_HandleGetGPGKey_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "test-source", namespace.ID, "0829C61C7E3")

	// Create repositories and service
	gpgKeyRepository, err := gpgkeyRepo.NewGPGKeyRepository(db.DB)
	require.NoError(t, err)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	gpgKeyService := service.NewGPGKeyService(gpgKeyRepository, namespaceRepository)

	// Create query
	getGPGKeyQuery := gpgkeyQuery.NewGetGPGKeyQuery(gpgKeyService)

	// Create handler
	handler := v2.NewTerraformV2GPGHandler(
		nil,
		nil,
		nil,
		getGPGKeyQuery,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v2/gpg-keys/test-namespace/0829C61C7E3", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("key_id", "0829C61C7E3")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleGetGPGKey(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure (matches Python's test depth)
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "0829C61C7E3", data["id"])
	assert.Equal(t, "gpg-keys", data["type"])

	attributes := data["attributes"].(map[string]interface{})
	assert.Equal(t, "test-namespace", attributes["namespace"])
	assert.Equal(t, "0829C61C7E3", attributes["key-id"])
	assert.Equal(t, "test-source", attributes["source"])

	// Validate ASCII armor content (Python validates this)
	asciiArmor := attributes["ascii-armor"].(string)
	assert.True(t, strings.Contains(asciiArmor, "BEGIN PGP PUBLIC KEY BLOCK"), "ASCII armor should contain BEGIN header")
	assert.True(t, strings.Contains(asciiArmor, "END PGP PUBLIC KEY BLOCK"), "ASCII armor should contain END header")
	assert.True(t, strings.Contains(asciiArmor, "mI0EZVJWdwEEAN2WER9iSataTlQThf"), "ASCII armor should contain key data")

	// Note: Fingerprint is not exposed in the API response (only used internally)

	// Validate source-url is nil (Python validates null handling)
	sourceURL := attributes["source-url"]
	assert.Nil(t, sourceURL, "source-url should be nil when not set")

	// Validate timestamp format (Python validates timestamp format)
	createdAt := attributes["created-at"].(string)
	assert.NotEmpty(t, createdAt, "created-at should not be empty")
	assert.Contains(t, createdAt, "T", "created-at should be ISO format timestamp")

	updatedAt := attributes["updated-at"].(string)
	assert.NotEmpty(t, updatedAt, "updated-at should not be empty")
	assert.Contains(t, updatedAt, "T", "updated-at should be ISO format timestamp")

	// Validate trust-signature (Python validates this field)
	trustSignature := attributes["trust-signature"].(string)
	assert.Empty(t, trustSignature, "trust-signature should be empty when not set")
}

// TestTerraformV2GPGHandler_Integration_HandleGetGPGKey_NotFound tests GPG key not found
func TestTerraformV2GPGHandler_Integration_HandleGetGPGKey_NotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - namespace exists but no GPG key
	_ = testutils.CreateNamespace(t, db, "test-namespace", nil)

	// Create repositories and service
	gpgKeyRepository, err := gpgkeyRepo.NewGPGKeyRepository(db.DB)
	require.NoError(t, err)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	gpgKeyService := service.NewGPGKeyService(gpgKeyRepository, namespaceRepository)

	// Create query
	getGPGKeyQuery := gpgkeyQuery.NewGetGPGKeyQuery(gpgKeyService)

	// Create handler
	handler := v2.NewTerraformV2GPGHandler(
		nil,
		nil,
		nil,
		getGPGKeyQuery,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v2/gpg-keys/test-namespace/MISSING", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("key_id", "MISSING")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleGetGPGKey(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"], "not found")
}

// TestTerraformV2GPGHandler_Integration_HandleGetGPGKey_MissingParameters tests missing path parameters
func TestTerraformV2GPGHandler_Integration_HandleGetGPGKey_MissingParameters(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler with nil dependencies (won't be called)
	handler := v2.NewTerraformV2GPGHandler(nil, nil, nil, nil)

	tests := []struct {
		name      string
		namespace string
		keyID     string
	}{
		{
			name:      "missing namespace",
			namespace: "",
			keyID:     "ABCD1234",
		},
		{
			name:      "missing key_id",
			namespace: "test-namespace",
			keyID:     "",
		},
		{
			name:      "missing both",
			namespace: "",
			keyID:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with chi context
			req := httptest.NewRequest("GET", "/v2/gpg-keys/"+tt.namespace+"/"+tt.keyID, nil)
			rctx := chi.NewRouteContext()
			if tt.namespace != "" {
				rctx.URLParams.Add("namespace", tt.namespace)
			}
			if tt.keyID != "" {
				rctx.URLParams.Add("key_id", tt.keyID)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			// Act
			handler.HandleGetGPGKey(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Contains(t, response["message"], "Missing required parameters")
		})
	}
}

// TestTerraformV2GPGHandler_Integration_URLParsing tests URL parsing with special characters
func TestTerraformV2GPGHandler_Integration_URLParsing(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with special characters in namespace name
	namespaceName := "test-namespace-with-dashes"
	namespace := testutils.CreateNamespace(t, db, namespaceName, nil)
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "test-source", namespace.ID, "0829C61C7E3")

	// Create repositories and service
	gpgKeyRepository, err := gpgkeyRepo.NewGPGKeyRepository(db.DB)
	require.NoError(t, err)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	gpgKeyService := service.NewGPGKeyService(gpgKeyRepository, namespaceRepository)

	// Create queries
	getNamespaceGPGKeysQuery := gpgkeyQuery.NewGetNamespaceGPGKeysQuery(gpgKeyService)

	// Create handler
	handler := v2.NewTerraformV2GPGHandler(
		nil,
		getNamespaceGPGKeysQuery,
		nil,
		nil,
	)

	// URL encode the namespace name
	reqURL := "/v2/gpg-keys?filter[namespace]=" + url.QueryEscape(namespaceName)
	req := httptest.NewRequest("GET", reqURL, nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListGPGKeys(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	data := response["data"].([]interface{})
	assert.Len(t, data, 1)
}

// TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_WhitespaceNamespaces tests trimming whitespace from namespace list
func TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_WhitespaceNamespaces(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace1 := testutils.CreateNamespace(t, db, "ns1", nil)
	namespace2 := testutils.CreateNamespace(t, db, "ns2", nil)
	namespace3 := testutils.CreateNamespace(t, db, "ns3", nil)
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "source1", namespace1.ID, "0829C61C7E3")
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "source2", namespace2.ID, "0829C61C7E3")
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "source3", namespace3.ID, "0829C61C7E3")

	// Create repositories and service
	gpgKeyRepository, err := gpgkeyRepo.NewGPGKeyRepository(db.DB)
	require.NoError(t, err)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	gpgKeyService := service.NewGPGKeyService(gpgKeyRepository, namespaceRepository)

	// Create queries
	getMultipleNamespaceGPGKeysQuery := gpgkeyQuery.NewGetMultipleNamespaceGPGKeysQuery(gpgKeyService)

	// Create handler
	handler := v2.NewTerraformV2GPGHandler(
		nil,
		nil,
		getMultipleNamespaceGPGKeysQuery,
		nil,
	)

	// Create request with whitespace around namespaces (should be trimmed)
	// Note: spaces need to be URL-encoded for httptest.NewRequest
	reqURL := "/v2/gpg-keys?filter[namespace]=%20ns1%20%2C%20ns2%20%2C%20ns3%20"
	req := httptest.NewRequest("GET", reqURL, nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListGPGKeys(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure - should get all 3 keys
	data := response["data"].([]interface{})
	assert.Len(t, data, 3)
}

// TestTerraformV2GPGHandler_Integration_HandleGetGPGKey_KeyIDWithSpecialChars tests key IDs with special characters
func TestTerraformV2GPGHandler_Integration_HandleGetGPGKey_KeyIDWithSpecialChars(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with special characters in key ID
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	specialKeyID := "0829/C61C-7E3" // Key IDs can contain special characters
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "test-source", namespace.ID, specialKeyID)

	// Create repositories and service
	gpgKeyRepository, err := gpgkeyRepo.NewGPGKeyRepository(db.DB)
	require.NoError(t, err)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	gpgKeyService := service.NewGPGKeyService(gpgKeyRepository, namespaceRepository)

	// Create query
	getGPGKeyQuery := gpgkeyQuery.NewGetGPGKeyQuery(gpgKeyService)

	// Create handler
	handler := v2.NewTerraformV2GPGHandler(
		nil,
		nil,
		nil,
		getGPGKeyQuery,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v2/gpg-keys/test-namespace/"+specialKeyID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("key_id", specialKeyID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleGetGPGKey(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	data := response["data"].(map[string]interface{})
	assert.Equal(t, specialKeyID, data["id"])
}

// TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_EmptyResult tests listing GPG keys when namespace has no keys
func TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_EmptyResult(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data - namespace exists but has no GPG keys
	_ = testutils.CreateNamespace(t, db, "empty-namespace", nil)

	// Create repositories and service
	gpgKeyRepository, err := gpgkeyRepo.NewGPGKeyRepository(db.DB)
	require.NoError(t, err)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	gpgKeyService := service.NewGPGKeyService(gpgKeyRepository, namespaceRepository)

	// Create queries
	getNamespaceGPGKeysQuery := gpgkeyQuery.NewGetNamespaceGPGKeysQuery(gpgKeyService)

	// Create handler
	handler := v2.NewTerraformV2GPGHandler(
		nil,
		getNamespaceGPGKeysQuery,
		nil,
		nil,
	)

	// Create request
	req := httptest.NewRequest("GET", "/v2/gpg-keys?filter[namespace]=empty-namespace", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListGPGKeys(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure - should be empty array
	data := response["data"].([]interface{})
	assert.Len(t, data, 0)
}

// TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_JSONContentType tests that response has correct content type
func TestTerraformV2GPGHandler_Integration_HandleListGPGKeys_JSONContentType(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "test-source", namespace.ID, "0829C61C7E3")

	// Create repositories and service
	gpgKeyRepository, err := gpgkeyRepo.NewGPGKeyRepository(db.DB)
	require.NoError(t, err)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	gpgKeyService := service.NewGPGKeyService(gpgKeyRepository, namespaceRepository)

	// Create queries
	getNamespaceGPGKeysQuery := gpgkeyQuery.NewGetNamespaceGPGKeysQuery(gpgKeyService)

	// Create handler
	handler := v2.NewTerraformV2GPGHandler(
		nil,
		getNamespaceGPGKeysQuery,
		nil,
		nil,
	)

	// Create request
	req := httptest.NewRequest("GET", "/v2/gpg-keys?filter[namespace]=test-namespace", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListGPGKeys(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Check content type
	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "application/json", contentType)

	// Check response is valid JSON
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "data")
}

// TestTerraformV2GPGHandler_Integration_HandleGetGPGKey_ASCIIArmorInResponse tests that ASCII armor is included in response
func TestTerraformV2GPGHandler_Integration_HandleGetGPGKey_ASCIIArmorInResponse(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data
	namespace := testutils.CreateNamespace(t, db, "test-namespace", nil)
	_ = testutils.CreateGPGKeyWithNamespace(t, db, "test-source", namespace.ID, "0829C61C7E3")

	// Create repositories and service
	gpgKeyRepository, err := gpgkeyRepo.NewGPGKeyRepository(db.DB)
	require.NoError(t, err)
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	gpgKeyService := service.NewGPGKeyService(gpgKeyRepository, namespaceRepository)

	// Create query
	getGPGKeyQuery := gpgkeyQuery.NewGetGPGKeyQuery(gpgKeyService)

	// Create handler
	handler := v2.NewTerraformV2GPGHandler(
		nil,
		nil,
		nil,
		getGPGKeyQuery,
	)

	// Create request with chi context
	req := httptest.NewRequest("GET", "/v2/gpg-keys/test-namespace/0829C61C7E3", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "test-namespace")
	rctx.URLParams.Add("key_id", "0829C61C7E3")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	// Act
	handler.HandleGetGPGKey(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify ASCII armor is in response
	data := response["data"].(map[string]interface{})
	attributes := data["attributes"].(map[string]interface{})
	asciiArmor := attributes["ascii-armor"].(string)
	assert.True(t, strings.Contains(asciiArmor, "BEGIN PGP PUBLIC KEY BLOCK"))
	assert.True(t, strings.Contains(asciiArmor, "END PGP PUBLIC KEY BLOCK"))
	// Note: The real GPG key doesn't contain "Test ASCII armor" - it's a real key
}
