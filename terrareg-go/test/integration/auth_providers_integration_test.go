package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestAuthenticationProvidersIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Create test configuration
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)

	// Update infra config with auth provider settings
	infraConfig.AdminAuthenticationMethod = "saml"
	infraConfig.SAML2SPCertFile = "test-cert.pem"
	infraConfig.SAML2SPKeyFile = "test-key.pem"
	infraConfig.SAML2IDPSSOURL = "https://test-idp.com/sso"
	infraConfig.SAML2IDPSSOURLBinding = "HTTP-POST"
	infraConfig.SAML2IDPSSODescriptorURL = "https://test-idp.com/metadata"
	infraConfig.SAML2IDPAttributeMapping = map[string]string{
		"email":    "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
		"username": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
	}
	infraConfig.SAML2IDPAllowedEmailDomains = []string{"example.com", "test.com"}

	infraConfig.OIDCClientID = "test-client-id"
	infraConfig.OIDCClientSecret = "test-client-secret"
	infraConfig.OIDCIssuerURL = "https://test-oidc.com"
	infraConfig.OIDCAuthorizationURL = "https://test-oidc.com/auth"
	infraConfig.OIDCTokenURL = "https://test-oidc.com/token"
	infraConfig.OIDCUserInfoURL = "https://test-oidc.com/userinfo"
	infraConfig.OIDCAttributeMapping = map[string]string{
		"email":    "email",
		"username": "preferred_username",
		"fullname": "name",
	}
	infraConfig.OIDCAllowedEmailDomains = []string{"example.com", "test.com"}

	infraConfig.GitHubOAuthClientID = "test-github-client-id"
	infraConfig.GitHubOAuthClientSecret = "test-github-client-secret"
	infraConfig.GitHubOAuthAllowedOrganizations = []string{"test-org", "example-org"}

	// Create container with test configuration
	container, err := container.NewContainer(domainConfig, infraConfig, nil, testutils.GetTestLogger(), db)
	require.NoError(t, err)

	// Setup test server
	server := container.Server
	router := server.GetRouter()

	t.Run("OIDC Login Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/openid/login", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should redirect to OIDC provider
		assert.Equal(t, http.StatusFound, w.Code)

		location := w.Header().Get("Location")
		assert.NotEmpty(t, location)

		// Parse redirect URL to verify it contains expected parameters
		redirectURL, err := url.Parse(location)
		require.NoError(t, err)

		assert.Equal(t, infraConfig.OIDCAuthorizationURL, redirectURL.Scheme+"://"+redirectURL.Host+redirectURL.Path)
		assert.Equal(t, infraConfig.OIDCClientID, redirectURL.Query().Get("client_id"))
		assert.Equal(t, "code", redirectURL.Query().Get("response_type"))
		assert.Contains(t, redirectURL.Query().Get("scope"), "openid")
		assert.NotEmpty(t, redirectURL.Query().Get("state"))
		assert.NotEmpty(t, redirectURL.Query().Get("redirect_uri"))
	})

	t.Run("OIDC Callback Endpoint", func(t *testing.T) {
		// Test callback with error
		req := httptest.NewRequest("GET", "/openid/callback?error=access_denied", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "access_denied")

		// Test successful callback (would need actual OIDC provider in real test)
		req = httptest.NewRequest("GET", "/openid/callback?state=test-state&code=test-code", nil)
		w = httptest.NewRecorder()

		// Add test session cookie
		req.AddCookie(&http.Cookie{Name: "terrareg_oauth_state", Value: "test-state"})

		router.ServeHTTP(w, req)

		// Should fail since we don't have a real OIDC provider
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("SAML Login Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/saml/login", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return SAML auth request
		assert.Equal(t, http.StatusOK, w.Code)

		// Response should be a SAMLRequest form
		body := w.Body.String()
		assert.Contains(t, body, "SAMLRequest")
		assert.Contains(t, body, "RelayState")
		assert.Contains(t, body, "https://test-idp.com/sso")
	})

	t.Run("SAML Metadata Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/saml/metadata", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))

		// Response should be XML with EntityDescriptor
		body := w.Body.String()
		assert.Contains(t, body, "EntityDescriptor")
		assert.Contains(t, body, "SPSSODescriptor")
		assert.Contains(t, body, "AssertionConsumerService")
	})

	t.Run("GitHub OAuth Login Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/github/login", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should redirect to GitHub OAuth
		assert.Equal(t, http.StatusFound, w.Code)

		location := w.Header().Get("Location")
		assert.NotEmpty(t, location)

		// Parse redirect URL
		redirectURL, err := url.Parse(location)
		require.NoError(t, err)

		assert.Equal(t, "github.com", redirectURL.Host)
		assert.Equal(t, "/login/oauth/authorize", redirectURL.Path)
		assert.Equal(t, infraConfig.GitHubOAuthClientID, redirectURL.Query().Get("client_id"))
	})

	t.Run("GitHub OAuth Callback Endpoint", func(t *testing.T) {
		// Test callback with error
		req := httptest.NewRequest("GET", "/github/callback?error=redirect_uri_mismatch", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "redirect_uri_mismatch")

		// Test successful callback (would need actual GitHub OAuth in real test)
		req = httptest.NewRequest("GET", "/github/callback?state=test-state&code=test-code", nil)
		w = httptest.NewRecorder()

		// Add test session cookie
		req.AddCookie(&http.Cookie{Name: "terrareg_oauth_state", Value: "test-state"})

		router.ServeHTTP(w, req)

		// Should fail since we don't have a real GitHub OAuth setup
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Admin Login Endpoint", func(t *testing.T) {
		// Test with valid admin token
		loginReq := map[string]interface{}{
			"token": "test-admin-token",
		}
		reqBody, _ := json.Marshal(loginReq)

		req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should fail since we haven't set up the admin token
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Check Session Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/terrareg/auth/admin/is_authenticated", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, float64(false), response["authenticated"])
	})

	t.Run("Multiple Auth Methods Configuration", func(t *testing.T) {
		// Test that all auth methods are properly configured
		assert.NotEmpty(t, infraConfig.SAML2IDPSSOURL)
		assert.NotEmpty(t, infraConfig.OIDCIssuerURL)
		assert.NotEmpty(t, infraConfig.GitHubOAuthClientID)
		assert.NotEmpty(t, infraConfig.OIDCClientSecret)
		assert.NotEmpty(t, infraConfig.GitHubOAuthClientSecret)

		// Verify authentication commands are initialized
		assert.NotNil(t, container.OidcLoginCmd)
		assert.NotNil(t, container.OidcCallbackCmd)
		assert.NotNil(t, container.SamlLoginCmd)
		assert.NotNil(t, container.SamlMetadataCmd)
		assert.NotNil(t, container.GithubOAuthCmd)
	})
}

func TestAuthenticationProvidersErrorHandling(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Create test configuration with invalid settings
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)

	// Set invalid configuration
	infraConfig.SAML2IDPSSOURL = "invalid-url"
	infraConfig.OIDCIssuerURL = "https://invalid-url-that-does-not-exist.com"

	container, err := container.NewContainer(domainConfig, infraConfig, nil, testutils.GetTestLogger(), db)
	require.NoError(t, err)

	server := container.Server
	router := server.GetRouter()

	t.Run("Handle invalid SAML configuration", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/saml/login", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return error due to invalid configuration
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "SAML")
	})

	t.Run("Handle invalid OIDC configuration", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/openid/login", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return error due to invalid OIDC configuration
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "OIDC")
	})
}

func TestAuthenticationProvidersSecurity(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)

	container, err := container.NewContainer(domainConfig, infraConfig, nil, testutils.GetTestLogger(), db)
	require.NoError(t, err)

	server := container.Server
	router := server.GetRouter()

	t.Run("CSRF Protection", func(t *testing.T) {
		// Test that state parameter is generated for OAuth flows
		req := httptest.NewRequest("GET", "/openid/login", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code == http.StatusFound {
			location := w.Header().Get("Location")
			redirectURL, err := url.Parse(location)
			require.NoError(t, err)

			// State should be present and not empty
			state := redirectURL.Query().Get("state")
			assert.NotEmpty(t, state)
			assert.Len(t, state, 32) // Should be a secure random string
		}
	})

	t.Run("Invalid State Parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/openid/callback?state=invalid-state&code=test-code", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid state parameter")
	})

	t.Run("Session Cookie Security", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/terrareg/auth/admin/is_authenticated", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Check that session cookies have security attributes
		cookies := w.Result().Cookies()
		for _, cookie := range cookies {
			if cookie.Name == "terrareg_session" || cookie.Name == "terrareg_oauth_state" {
				assert.True(t, cookie.HttpOnly)
				assert.True(t, cookie.Secure)
				assert.Equal(t, "Strict", cookie.SameSite)
			}
		}
	})
}
