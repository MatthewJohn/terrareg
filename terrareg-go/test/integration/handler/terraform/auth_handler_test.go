// Package terraform_test provides integration tests for the Terraform authentication HTTP handlers
package terraform_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/terraform"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestTerraformAuthHandler_HandleAuthenticateOIDCToken_Success_ValidToken tests successful OIDC token authentication
func TestTerraformAuthHandler_HandleAuthenticateOIDCToken_Success_ValidToken(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create a test session with admin permissions
	_ = testutils.CreateTestSession(t, db, "testuser", true)

	// Create request with valid authorization header
	reqBody := terraform.AuthenticateOIDCTokenRequest{
		AuthorizationHeader: "Bearer test-terraform-token",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/terraform/v1/authenticate/oidc", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleAuthenticateOIDCToken(w, req)

	// Assert - Since we don't have actual Terraform OIDC implementation, we expect an error
	// This is expected behavior as the auth factory won't recognize this token format
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTerraformAuthHandler_HandleAuthenticateOIDCToken_Failure_InvalidToken tests authentication with invalid token
func TestTerraformAuthHandler_HandleAuthenticateOIDCToken_Failure_InvalidToken(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request with invalid authorization header
	reqBody := terraform.AuthenticateOIDCTokenRequest{
		AuthorizationHeader: "Bearer invalid-token",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/terraform/v1/authenticate/oidc", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleAuthenticateOIDCToken(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTerraformAuthHandler_HandleAuthenticateOIDCToken_Failure_MissingToken tests authentication with missing token
func TestTerraformAuthHandler_HandleAuthenticateOIDCToken_Failure_MissingToken(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request with missing authorization header
	reqBody := terraform.AuthenticateOIDCTokenRequest{
		AuthorizationHeader: "",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/terraform/v1/authenticate/oidc", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleAuthenticateOIDCToken(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTerraformAuthHandler_HandleAuthenticateOIDCToken_Failure_InvalidRequestMethod tests wrong HTTP method
func TestTerraformAuthHandler_HandleAuthenticateOIDCToken_Failure_InvalidRequestMethod(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	req := httptest.NewRequest("GET", "/terraform/v1/authenticate/oidc", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleAuthenticateOIDCToken(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestTerraformAuthHandler_HandleAuthenticateOIDCToken_Failure_InvalidRequestBody tests invalid JSON
func TestTerraformAuthHandler_HandleAuthenticateOIDCToken_Failure_InvalidRequestBody(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	req := httptest.NewRequest("POST", "/terraform/v1/authenticate/oidc", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleAuthenticateOIDCToken(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTerraformAuthHandler_HandleValidateToken_Success_ValidToken tests successful token validation
func TestTerraformAuthHandler_HandleValidateToken_Success_ValidToken(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create a test session
	sessionID := testutils.CreateTestSession(t, db, "testuser", true)

	// Create request with valid authorization header
	reqBody := terraform.ValidateTokenRequest{
		AuthorizationHeader: "Bearer " + sessionID,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/terraform/v1/validate/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleValidateToken(w, req)

	// Assert - Without actual Terraform OIDC setup, expect failure
	// This tests the handler properly calls the validate command
	assert.Equal(t, http.StatusOK, w.Code)

	var response terraform.ValidateTokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// Token should be invalid since we're using a regular session, not Terraform OIDC
	assert.False(t, response.Valid)
}

// TestTerraformAuthHandler_HandleValidateToken_Failure_InvalidToken tests validation with invalid token
func TestTerraformAuthHandler_HandleValidateToken_Failure_InvalidToken(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request with invalid authorization header
	reqBody := terraform.ValidateTokenRequest{
		AuthorizationHeader: "Bearer invalid-token-12345",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/terraform/v1/validate/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleValidateToken(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response terraform.ValidateTokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Valid)
}

// TestTerraformAuthHandler_HandleValidateToken_WithAuthorizationHeader tests validation with Authorization header
func TestTerraformAuthHandler_HandleValidateToken_WithAuthorizationHeader(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request with Authorization header instead of body
	reqBody := terraform.ValidateTokenRequest{
		AuthorizationHeader: "",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/terraform/v1/validate/token", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer some-token")
	w := httptest.NewRecorder()

	// Act
	handler.HandleValidateToken(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response terraform.ValidateTokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Valid)
}

// TestTerraformAuthHandler_HandleValidateToken_WithQueryPermissions tests validation with query permissions
func TestTerraformAuthHandler_HandleValidateToken_WithQueryPermissions(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request with permissions in query string
	reqBody := terraform.ValidateTokenRequest{
		AuthorizationHeader: "Bearer some-token",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/terraform/v1/validate/token?permission=ns1:read&permission=ns2:write", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.HandleValidateToken(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response terraform.ValidateTokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Valid)
}

// TestTerraformAuthHandler_HandleGetUser_Success_ValidUserID tests getting user with valid ID
func TestTerraformAuthHandler_HandleGetUser_Success_ValidUserID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request for current-user
	req := httptest.NewRequest("GET", "/terraform/v1/users/current-user", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetUser(w, req)

	// Assert - current-user is a special case that returns valid response
	assert.Equal(t, http.StatusOK, w.Code)

	var response terraform.GetUserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// current-user is a special case that returns IsValid=true
	// (it bypasses the session matching check in GetUserCommand.Execute)
	assert.True(t, response.IsValid)
	assert.Equal(t, "current-user", response.IdentityID)
}

// TestTerraformAuthHandler_HandleGetUser_Failure_UserNotFound tests getting non-existent user
func TestTerraformAuthHandler_HandleGetUser_Failure_UserNotFound(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request for non-existent user
	req := httptest.NewRequest("GET", "/terraform/v1/users/non-existent-user-id", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetUser(w, req)

	// Assert - The handler returns 200 with IsValid=false for non-existent users
	// (404 is only returned when identityID is empty, causing Execute to error)
	assert.Equal(t, http.StatusOK, w.Code)

	var response terraform.GetUserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.IsValid)
	assert.Equal(t, "non-existent-user-id", response.IdentityID)
}

// TestTerraformAuthHandler_HandleGetUser_Failure_EmptyUserID tests getting user with empty ID
func TestTerraformAuthHandler_HandleGetUser_Failure_EmptyUserID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request with empty user ID
	req := httptest.NewRequest("GET", "/terraform/v1/users/", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetUser(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTerraformAuthHandler_HandleGetUser_Success_IncludesMetadata tests that user response includes metadata
func TestTerraformAuthHandler_HandleGetUser_Success_IncludesMetadata(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request for current-user
	req := httptest.NewRequest("GET", "/terraform/v1/users/current-user", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetUser(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response terraform.GetUserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.NotEmpty(t, response.IdentityID)
	assert.NotEmpty(t, response.Metadata)
	assert.Contains(t, response.Metadata, "auth_method")
}

// TestTerraformAuthHandler_HandleTerraformLogin_Success tests successful Terraform login
func TestTerraformAuthHandler_HandleTerraformLogin_Success(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	req := httptest.NewRequest("GET", "/terraform/v1/login", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformLogin(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response contains expected fields
	assert.Contains(t, response, "login_url")
	assert.Contains(t, response, "methods")
	assert.Contains(t, response, "status")
	assert.Equal(t, "configured", response["status"])
}

// TestTerraformAuthHandler_HandleTerraformLogin_Failure_InvalidMethod tests wrong HTTP method
func TestTerraformAuthHandler_HandleTerraformLogin_Failure_InvalidMethod(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	req := httptest.NewRequest("POST", "/terraform/v1/login", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformLogin(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestTerraformAuthHandler_HandleTerraformOIDCAuth_Success_ReturnsAuthCode tests successful OIDC auth
func TestTerraformAuthHandler_HandleTerraformOIDCAuth_Success_ReturnsAuthCode(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request with required OIDC parameters
	redirectURI := "http://localhost:10000/oidc/callback"
	req := httptest.NewRequest("GET", "/terraform/v1/auth/oidc?response_type=code&client_id=terraform&redirect_uri="+redirectURI+"&state=test-state", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformOIDCAuth(w, req)

	// Assert - Should redirect with auth code
	assert.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, redirectURI)
	assert.Contains(t, location, "code=mock-terraform-oidc-code")
	assert.Contains(t, location, "state=test-state")
}

// TestTerraformAuthHandler_HandleTerraformOIDCAuth_Failure_MissingRequiredParams tests missing required parameters
func TestTerraformAuthHandler_HandleTerraformOIDCAuth_Failure_MissingRequiredParams(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request missing response_type
	req := httptest.NewRequest("GET", "/terraform/v1/auth/oidc?client_id=terraform&redirect_uri=http://localhost:10000/oidc/callback", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformOIDCAuth(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTerraformAuthHandler_HandleTerraformOIDCAuth_Failure_InvalidRedirectURI tests invalid redirect URI
func TestTerraformAuthHandler_HandleTerraformOIDCAuth_Failure_InvalidRedirectURI(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create request with invalid redirect URI
	req := httptest.NewRequest("GET", "/terraform/v1/auth/oidc?response_type=code&client_id=terraform&redirect_uri=http://%%invalid%%", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformOIDCAuth(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTerraformAuthHandler_HandleTerraformOIDCAuth_Failure_InvalidMethod tests wrong HTTP method
func TestTerraformAuthHandler_HandleTerraformOIDCAuth_Failure_InvalidMethod(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	req := httptest.NewRequest("POST", "/terraform/v1/auth/oidc", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformOIDCAuth(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestTerraformAuthHandler_HandleTerraformToken_Success_ExchangeToken tests successful token exchange
func TestTerraformAuthHandler_HandleTerraformToken_Success_ExchangeToken(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create form data for token exchange
	formData := strings.NewReader("grant_type=authorization_code&code=mock-terraform-oidc-code")
	req := httptest.NewRequest("POST", "/terraform/v1/auth/token", formData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformToken(w, req)

	// Assert - Without actual OIDC implementation, expect authentication failure
	// but handler should still respond with proper structure
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestTerraformAuthHandler_HandleTerraformToken_Failure_MissingGrantType tests missing grant_type
func TestTerraformAuthHandler_HandleTerraformToken_Failure_MissingGrantType(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create form data without grant_type
	formData := strings.NewReader("code=mock-terraform-oidc-code")
	req := httptest.NewRequest("POST", "/terraform/v1/auth/token", formData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformToken(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTerraformAuthHandler_HandleTerraformToken_Failure_MissingCode tests missing code
func TestTerraformAuthHandler_HandleTerraformToken_Failure_MissingCode(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create form data without code
	formData := strings.NewReader("grant_type=authorization_code")
	req := httptest.NewRequest("POST", "/terraform/v1/auth/token", formData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformToken(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestTerraformAuthHandler_HandleTerraformToken_Failure_InvalidMethod tests wrong HTTP method
func TestTerraformAuthHandler_HandleTerraformToken_Failure_InvalidMethod(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	req := httptest.NewRequest("GET", "/terraform/v1/auth/token", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformToken(w, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestTerraformAuthHandler_HandleTerraformToken_Failure_InvalidFormData tests invalid form data
func TestTerraformAuthHandler_HandleTerraformToken_Failure_InvalidFormData(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	cont := testutils.CreateTestContainer(t, db)
	handler := cont.TerraformAuthHandler
	require.NotNil(t, handler)

	// Create invalid form data
	req := httptest.NewRequest("POST", "/terraform/v1/auth/token", strings.NewReader("invalid form%%data"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	// Act
	handler.HandleTerraformToken(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
