package terrareg_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestConfigHandler_HandleConfig_Success tests getting config successfully
func TestConfigHandler_HandleConfig_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewConfigHandler(cont.GetConfigQuery)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/config", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleConfig(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	testutils.AssertContentType(t, w, "application/json")

	response := testutils.GetJSONBody(t, w)

	// Verify config structure - should have common config fields
	assert.Contains(t, response, "ALLOW_MODULE_HOSTING")
	assert.Contains(t, response, "SECRET_KEY_SET")
	assert.Contains(t, response, "OPENID_CONNECT_ENABLED")
	assert.Contains(t, response, "SAML_ENABLED")
}

// TestConfigHandler_HandleConfig_SensitiveDataFiltered tests that sensitive data is filtered from response
func TestConfigHandler_HandleConfig_SensitiveDataFiltered(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewConfigHandler(cont.GetConfigQuery)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/config", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleConfig(w, req)

	// Assert
	response := testutils.GetJSONBody(t, w)

	// SECRET_KEY should be filtered (shows SECRET_KEY_SET instead)
	assert.NotContains(t, response, "SECRET_KEY")
	assert.Contains(t, response, "SECRET_KEY_SET")

	// Test config has secret key set to true
	assert.Equal(t, true, response["SECRET_KEY_SET"])
}

// TestConfigHandler_HandleConfig_Structure tests the response structure
func TestConfigHandler_HandleConfig_Structure(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewConfigHandler(cont.GetConfigQuery)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/config", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleConfig(w, req)

	// Assert
	response := testutils.GetJSONBody(t, w)

	// Verify boolean fields (some are actual bools, some are strings)
	assert.Contains(t, response, "ALLOW_MODULE_HOSTING")
	assert.Contains(t, response, "TRUSTED_NAMESPACE_LABEL")
	assert.Contains(t, response, "VERIFIED_MODULE_LABEL")

	// SECRET_KEY_SET should be boolean
	if val, ok := response["SECRET_KEY_SET"]; ok {
		assert.IsType(t, false, val, "SECRET_KEY_SET should be boolean")
	}
	// OPENID_CONNECT_ENABLED and SAML_ENABLED should be boolean
	if val, ok := response["OPENID_CONNECT_ENABLED"]; ok {
		assert.IsType(t, false, val, "OPENID_CONNECT_ENABLED should be boolean")
	}
	if val, ok := response["SAML_ENABLED"]; ok {
		assert.IsType(t, false, val, "SAML_ENABLED should be boolean")
	}
}

// TestConfigHandler_HandleConfig_MethodNotAllowed tests that non-GET requests are rejected
func TestConfigHandler_HandleConfig_MethodNotAllowed(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewConfigHandler(cont.GetConfigQuery)

	// Create POST request (should fail)
	req := httptest.NewRequest("POST", "/v1/terrareg/config", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleConfig(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestConfigHandler_HandleConfig_PutMethodNotAllowed tests that PUT requests are rejected
func TestConfigHandler_HandleConfig_PutMethodNotAllowed(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewConfigHandler(cont.GetConfigQuery)

	// Create PUT request (should fail)
	req := httptest.NewRequest("PUT", "/v1/terrareg/config", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleConfig(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestConfigHandler_HandleConfig_DeleteMethodNotAllowed tests that DELETE requests are rejected
func TestConfigHandler_HandleConfig_DeleteMethodNotAllowed(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewConfigHandler(cont.GetConfigQuery)

	// Create DELETE request (should fail)
	req := httptest.NewRequest("DELETE", "/v1/terrareg/config", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleConfig(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestConfigHandler_WithTestConfigValues tests that test config values are properly exposed
func TestConfigHandler_WithTestConfigValues(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler
	cont := testutils.CreateTestContainer(t, db)
	handler := terrareg.NewConfigHandler(cont.GetConfigQuery)

	// Create request
	req := httptest.NewRequest("GET", "/v1/terrareg/config", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleConfig(w, req)

	// Assert
	response := testutils.GetJSONBody(t, w)

	// Verify test config values from CreateTestInfraConfig
	assert.Equal(t, true, response["SECRET_KEY_SET"])
	assert.Equal(t, true, response["OPENID_CONNECT_ENABLED"])
	assert.Equal(t, true, response["SAML_ENABLED"])

	// Test config has ALLOW_MODULE_HOSTING set
	assert.Contains(t, response, "ALLOW_MODULE_HOSTING")
}
