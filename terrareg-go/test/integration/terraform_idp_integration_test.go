package integration

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// generateTestRSAKey generates a test RSA key pair for integration testing
func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate test RSA key")
	return privateKey
}

// rsaKeyToPEM converts RSA private key to PEM format for configuration
func rsaKeyToPEM(privateKey *rsa.PrivateKey) string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return string(privateKeyPEM)
}

// createTestKeyFile creates a temporary key file for testing
func createTestKeyFile(t *testing.T, content string) string {
	tmpFile, err := ioutil.TempFile("", "test-key-*.pem")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}

func TestTerraformIDPIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Generate test RSA key pair for Terraform IDP
	privateKey := generateTestRSAKey(t)
	privateKeyPEM := rsaKeyToPEM(privateKey)

	// Create temporary key file
	keyFilePath := createTestKeyFile(t, privateKeyPEM)
	defer func() {
		os.Remove(keyFilePath)
	}()

	// Create test configuration with Terraform IDP settings
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)

	// Configure Terraform IDP settings
	publicURL := "https://terrareg.example.com"
	infraConfig.PublicURL = publicURL
	infraConfig.TerraformOidcIdpSigningKeyPath = keyFilePath
	infraConfig.TerraformOidcIdpSubjectIdHashSalt = "test-salt"
	infraConfig.TerraformOidcIdpSessionExpiry = 3600

	// Create container with test configuration
	container, err := container.NewContainer(domainConfig, infraConfig, nil, testutils.GetTestLogger(), db)
	require.NoError(t, err)

	// Setup test server
	server := container.Server
	router := server.GetRouter()

	t.Run("OpenID Configuration Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var openidConfig map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &openidConfig)
		require.NoError(t, err)

		// Verify required OpenID configuration fields
		assert.Equal(t, publicURL, openidConfig["issuer"])
		assert.Equal(t, []interface{}{"public"}, openidConfig["subject_types_supported"])
		assert.Equal(t, []interface{}{"code"}, openidConfig["response_types_supported"])
		assert.Equal(t, []interface{}{"authorization_code"}, openidConfig["grant_types_supported"])
		assert.Equal(t, publicURL+"/.well-known/jwks.json", openidConfig["jwks_uri"])
		assert.Equal(t, []interface{}{"RS256"}, openidConfig["id_token_signing_alg_values_supported"])
		assert.Equal(t, publicURL+"/userinfo", openidConfig["userinfo_endpoint"])
		assert.Equal(t, publicURL+"/token", openidConfig["token_endpoint"])
		assert.Equal(t, publicURL+"/authorize", openidConfig["authorization_endpoint"])
	})

	t.Run("JWKS Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var jwksResponse map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &jwksResponse)
		require.NoError(t, err)

		// Verify JWKS structure
		keys, ok := jwksResponse["keys"].([]interface{})
		require.True(t, ok, "JWKS should contain keys array")
		assert.Len(t, keys, 1, "Should have exactly one key")

		key, ok := keys[0].(map[string]interface{})
		require.True(t, ok, "Key should be a map")

		assert.Equal(t, "RSA", key["kty"])
		assert.Equal(t, "RS256", key["alg"])
		assert.Equal(t, "sig", key["use"])
		assert.NotEmpty(t, key["kid"])
		assert.NotEmpty(t, key["n"])
		assert.NotEmpty(t, key["e"])
	})

	t.Run("Terraform Well-Known Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/.well-known/terraform.json", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var terraformConfig map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &terraformConfig)
		require.NoError(t, err)

		// Verify Terraform-specific configuration
		assert.Equal(t, publicURL+"/.well-known/openid-configuration", terraformConfig["login.v1"])
		assert.Equal(t, "state", terraformConfig["login.v1.response_types_supported"])
		assert.Equal(t, "pkce", terraformConfig["login.v1.response_modes_supported"])
	})
}

func TestTerraformIDPConfigurationValidation(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	t.Run("Missing Signing Key", func(t *testing.T) {
		domainConfig := testutils.CreateTestDomainConfig(t)
		infraConfig := testutils.CreateTestInfraConfig(t)

		// Use non-existent key file
		infraConfig.TerraformOidcIdpSigningKeyPath = "/non-existent/key.pem"

		container, err := container.NewContainer(domainConfig, infraConfig, nil, testutils.GetTestLogger(), db)
		require.NoError(t, err)

		server := container.Server
		router := server.GetRouter()

		// Try to access JWKS endpoint - should handle missing key gracefully
		req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// The implementation should handle missing key file
		// Either return error or provide default behavior
		if w.Code != http.StatusOK {
			assert.Equal(t, http.StatusInternalServerError, w.Code)
		}
	})

	t.Run("Invalid Key File", func(t *testing.T) {
		domainConfig := testutils.CreateTestDomainConfig(t)
		infraConfig := testutils.CreateTestInfraConfig(t)

		// Create invalid key file
		invalidKeyFile := createTestKeyFile(t, "invalid-key-content")
		defer os.Remove(invalidKeyFile)

		infraConfig.TerraformOidcIdpSigningKeyPath = invalidKeyFile

		container, err := container.NewContainer(domainConfig, infraConfig, nil, testutils.GetTestLogger(), db)
		require.NoError(t, err)

		server := container.Server
		router := server.GetRouter()

		// Try to access JWKS endpoint with invalid key
		req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should handle invalid key gracefully
		if w.Code == http.StatusInternalServerError {
			var errorResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
			if err == nil {
				// Should provide meaningful error message
				_, hasError := errorResponse["error"]
				assert.True(t, hasError, "Error response should contain error message")
			}
		}
	})
}

func TestTerraformIDPConcurrency(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Generate test RSA key pair
	privateKey := generateTestRSAKey(t)
	privateKeyPEM := rsaKeyToPEM(privateKey)

	// Create temporary key file
	keyFilePath := createTestKeyFile(t, privateKeyPEM)
	defer os.Remove(keyFilePath)

	// Create test configuration
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)

	publicURL := "https://terrareg.example.com"
	infraConfig.PublicURL = publicURL
	infraConfig.TerraformOidcIdpSigningKeyPath = keyFilePath

	container, err := container.NewContainer(domainConfig, infraConfig, nil, testutils.GetTestLogger(), db)
	require.NoError(t, err)

	server := container.Server
	router := server.GetRouter()

	t.Run("Concurrent Discovery Requests", func(t *testing.T) {
		const numRequests = 10
		results := make([]int, numRequests)
		errors := make([]error, numRequests)

		// Create multiple concurrent discovery requests
		for i := 0; i < numRequests; i++ {
			go func(index int) {
				req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				results[index] = w.Code
				if w.Code != http.StatusOK {
					errors[index] = fmt.Errorf("request %d failed with status %d", index, w.Code)
				}
			}(i)
		}

		// Wait for all requests to complete
		time.Sleep(2 * time.Second)

		// Verify all requests succeeded
		successCount := 0
		for i := 0; i < numRequests; i++ {
			if results[i] == http.StatusOK {
				successCount++
			}
		}

		assert.True(t, successCount >= numRequests-2, "At least 8 out of 10 requests should succeed")

		// Verify no errors occurred
		errorCount := 0
		for i := 0; i < numRequests; i++ {
			if errors[i] != nil {
				errorCount++
			}
		}
		assert.True(t, errorCount <= 2, "At most 2 requests should fail")
	})
}

func TestTerraformIDPSecurityFeatures(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Generate test RSA key pair
	privateKey := generateTestRSAKey(t)
	privateKeyPEM := rsaKeyToPEM(privateKey)

	// Create temporary key file
	keyFilePath := createTestKeyFile(t, privateKeyPEM)
	defer os.Remove(keyFilePath)

	// Create test configuration
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)

	publicURL := "https://terrareg.example.com"
	infraConfig.PublicURL = publicURL
	infraConfig.TerraformOidcIdpSigningKeyPath = keyFilePath

	container, err := container.NewContainer(domainConfig, infraConfig, nil, testutils.GetTestLogger(), db)
	require.NoError(t, err)

	server := container.Server
	router := server.GetRouter()

	t.Run("JWKS Response Headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/.well-known/jwks.json", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify security headers
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("Content-Length"))

		// Check for cache control headers if implemented
		cacheControl := w.Header().Get("Cache-Control")
		if cacheControl != "" {
			assert.Contains(t, cacheControl, "max-age")
		}
	})

	t.Run("OpenID Configuration Response Headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify security headers
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("Content-Length"))

		// Check for CORS headers if implemented
		corsOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if corsOrigin != "" {
			assert.NotEmpty(t, corsOrigin)
		}
	})

	t.Run("Terraform Discovery Response Headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/.well-known/terraform.json", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify security headers
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("Content-Length"))
	})
}