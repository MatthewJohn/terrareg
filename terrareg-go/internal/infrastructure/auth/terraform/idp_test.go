package terraform

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestRSAKey generates a test RSA key pair for testing
func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate test RSA key")
	return privateKey
}

// rsaKeyToPEM converts RSA private key to PEM format for potential use in tests
func rsaKeyToPEM(privateKey *rsa.PrivateKey) string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	return string(privateKeyPEM)
}

func TestNewTerraformIDP(t *testing.T) {
	privateKey := generateTestRSAKey(t)

	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)

	assert.NoError(t, err)
	assert.NotNil(t, idp)
	assert.Equal(t, config.IssuerURL, idp.config.IssuerURL)
	assert.Equal(t, privateKey, idp.privateKey)
	assert.NotEmpty(t, idp.keyID)
}

func TestNewTerraformIDP_InvalidConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    TerraformIDPConfig
		privateKey *rsa.PrivateKey
		expectErr bool
		errMsg    string
	}{
		{
			name:      "nil private key",
			config:    TerraformIDPConfig{IssuerURL: "https://example.com"},
			privateKey: nil,
			expectErr: true,
			errMsg:    "private key is required",
		},
		{
			name:      "empty issuer URL",
			config:    TerraformIDPConfig{IssuerURL: ""},
			privateKey: generateTestRSAKey(t),
			expectErr: true,
			errMsg:    "issuer URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idp, err := NewTerraformIDP(tt.config, tt.privateKey)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, idp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, idp)
			}
		})
	}
}

func TestTerraformIDP_IsEnabled(t *testing.T) {
	tests := []struct {
		name      string
		config    TerraformIDPConfig
		expected  bool
		expectErr bool
	}{
		{
			name:      "enabled with issuer URL",
			config:    TerraformIDPConfig{IssuerURL: "https://example.com"},
			expected:  true,
			expectErr: false,
		},
		{
			name:      "error without issuer URL",
			config:    TerraformIDPConfig{IssuerURL: ""},
			expected:  false,
			expectErr: true, // Should fail during creation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idp, err := NewTerraformIDP(tt.config, generateTestRSAKey(t))

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, idp)
				assert.Contains(t, err.Error(), "issuer URL is required")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, idp)
				assert.Equal(t, tt.expected, idp.IsEnabled())
			}
		})
	}
}

func TestTerraformIDP_GetOpenIDConfiguration(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	openidConfig := idp.GetOpenIDConfiguration()

	assert.Equal(t, config.IssuerURL, openidConfig["issuer"])
	assert.Equal(t, []string{"public"}, openidConfig["subject_types_supported"])
	assert.Equal(t, []string{"code"}, openidConfig["response_types_supported"])
	assert.Equal(t, []string{"authorization_code"}, openidConfig["grant_types_supported"])
	assert.Equal(t, fmt.Sprintf("%s/.well-known/jwks.json", config.IssuerURL), openidConfig["jwks_uri"])
	assert.Equal(t, []string{"RS256"}, openidConfig["id_token_signing_alg_values_supported"])
	assert.Equal(t, fmt.Sprintf("%s/userinfo", config.IssuerURL), openidConfig["userinfo_endpoint"])
	assert.Equal(t, fmt.Sprintf("%s/token", config.IssuerURL), openidConfig["token_endpoint"])
	assert.Equal(t, fmt.Sprintf("%s/authorize", config.IssuerURL), openidConfig["authorization_endpoint"])
}

func TestTerraformIDP_GetJWKS(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	jwks, err := idp.GetJWKS()
	require.NoError(t, err)

	keys, ok := jwks["keys"].([]interface{})
	require.True(t, ok, "JWKS should contain keys array")
	assert.Len(t, keys, 1, "Should have exactly one key")

	key, ok := keys[0].(map[string]interface{})
	require.True(t, ok, "Key should be a map")

	assert.Equal(t, "RSA", key["kty"])
	assert.Equal(t, idp.keyID, key["kid"])
	assert.Equal(t, "RS256", key["alg"])
	assert.Equal(t, "sig", key["use"])
	assert.NotEmpty(t, key["n"])
	assert.NotEmpty(t, key["e"])

	// Verify the modulus is properly base64url encoded
	modulus, ok := key["n"].(string)
	require.True(t, ok)

	// Should be valid base64url without padding
	_, err = base64.RawURLEncoding.DecodeString(modulus)
	assert.NoError(t, err, "Modulus should be valid base64url")
}

func TestTerraformIDP_HandleTokenRequest(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	tokenRequest := map[string]interface{}{
		"subject": "terraform-user",
		"client_id": "test-client",
	}

	response, err := idp.HandleTokenRequest(context.Background(), tokenRequest)
	require.NoError(t, err)

	// Verify response structure
	assert.NotEmpty(t, response["access_token"])
	assert.Equal(t, "Bearer", response["token_type"])
	assert.Equal(t, int64(3600), response["expires_in"])
	assert.Equal(t, "openid", response["scope"])

	// Verify JWT token
	tokenStr, ok := response["access_token"].(string)
	require.True(t, ok, "Access token should be a string")

	// Parse and verify the JWT
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return privateKey.Public(), nil
	})
	require.NoError(t, err, "Should be able to parse JWT")
	require.True(t, token.Valid, "JWT should be valid")

	// Verify claims
	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok, "Should have map claims")

	assert.Equal(t, config.IssuerURL, claims["iss"])
	assert.Equal(t, "terraform-user", claims["sub"])
	assert.Equal(t, "terraform", claims["aud"])
	assert.Equal(t, idp.keyID, token.Header["kid"])
}

func TestTerraformIDP_HandleTokenRequest_DefaultSubject(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	// Request without subject
	tokenRequest := map[string]interface{}{
		"client_id": "test-client",
	}

	response, err := idp.HandleTokenRequest(context.Background(), tokenRequest)
	require.NoError(t, err)

	tokenStr, ok := response["access_token"].(string)
	require.True(t, ok)

	// Parse and verify the JWT
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return privateKey.Public(), nil
	})
	require.NoError(t, err)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Should use default subject
	assert.Equal(t, "terraform-user", claims["sub"])
}

func TestTerraformIDP_HandleTokenRequest_CustomExpiration(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      2 * time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	tokenRequest := map[string]interface{}{
		"subject": "terraform-user",
	}

	response, err := idp.HandleTokenRequest(context.Background(), tokenRequest)
	require.NoError(t, err)

	assert.Equal(t, int64(7200), response["expires_in"]) // 2 hours in seconds
}

func TestTerraformIDP_HandleUserInfoRequest(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	// Create a valid token
	tokenRequest := map[string]interface{}{
		"subject": "terraform-user",
		"name":    "Terraform User",
		"email":   "terraform@example.com",
	}

	tokenResponse, err := idp.HandleTokenRequest(context.Background(), tokenRequest)
	require.NoError(t, err)

	tokenStr := tokenResponse["access_token"].(string)

	// Test user info request
	userInfo, err := idp.HandleUserInfoRequest(context.Background(), tokenStr)
	require.NoError(t, err)

	assert.Equal(t, "terraform-user", userInfo["sub"])
	assert.Equal(t, "Terraform User", userInfo["name"])
	assert.Equal(t, "terraform@example.com", userInfo["email"])
}

func TestTerraformIDP_HandleUserInfoRequest_InvalidToken(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	// Test with invalid token
	invalidToken := "invalid.jwt.token"
	_, err = idp.HandleUserInfoRequest(context.Background(), invalidToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JWT token")
}

func TestTerraformIDP_HandleUserInfoRequest_TokenWithDifferentKey(t *testing.T) {
	privateKey1 := generateTestRSAKey(t)
	privateKey2 := generateTestRSAKey(t)

	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey1)
	require.NoError(t, err)

	// Create a valid token with a different private key
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": config.IssuerURL,
		"sub": "terraform-user",
		"aud": "terraform",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	})

	tokenStr, err := token.SignedString(privateKey2)
	require.NoError(t, err)

	// Should fail to validate token signed with different key
	_, err = idp.HandleUserInfoRequest(context.Background(), tokenStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JWT token")
}

func TestTerraformIDP_GenerateIDToken(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	subject := "terraform-user"
	additionalClaims := map[string]interface{}{
		"name":  "Terraform User",
		"email": "terraform@example.com",
		"groups": []string{"developers", "admins"},
	}

	idToken, err := idp.GenerateIDToken(context.Background(), subject, additionalClaims)
	require.NoError(t, err)
	assert.NotEmpty(t, idToken)

	// Parse and verify the ID token
	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return privateKey.Public(), nil
	})
	require.NoError(t, err)
	require.True(t, token.Valid)

	// Verify claims
	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, config.IssuerURL, claims["iss"])
	assert.Equal(t, subject, claims["sub"])
	assert.Equal(t, "terraform", claims["aud"])
	assert.Equal(t, "Terraform User", claims["name"])
	assert.Equal(t, "terraform@example.com", claims["email"])
	assert.Equal(t, idp.keyID, token.Header["kid"])
}

func TestTerraformIDP_ValidateToken(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	// Create a valid token
	subject := "terraform-user"
	additionalClaims := map[string]interface{}{
		"name":  "Terraform User",
		"email": "terraform@example.com",
	}

	idToken, err := idp.GenerateIDToken(context.Background(), subject, additionalClaims)
	require.NoError(t, err)

	// Validate the token
	claims, err := idp.ValidateToken(context.Background(), idToken)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, config.IssuerURL, claims["iss"])
	assert.Equal(t, subject, claims["sub"])
	assert.Equal(t, "terraform", claims["aud"])
	assert.Equal(t, "Terraform User", claims["name"])
	assert.Equal(t, "terraform@example.com", claims["email"])
}

func TestTerraformIDP_ValidateToken_InvalidToken(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	// Test with invalid token
	invalidToken := "invalid.jwt.token"
	_, err = idp.ValidateToken(context.Background(), invalidToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JWT token")
}

func TestTerraformIDP_GetKeyID(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	keyID := idp.GetKeyID()
	assert.NotEmpty(t, keyID)
	assert.Equal(t, idp.keyID, keyID)

	// Verify key ID is a UUID
	uuidPattern := `^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`
	assert.Regexp(t, uuidPattern, keyID)
}

func TestTerraformIDP_JWKSConsistency(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	// Get JWKS
	jwks, err := idp.GetJWKS()
	require.NoError(t, err)

	keys, ok := jwks["keys"].([]interface{})
	require.True(t, ok)
	require.Len(t, keys, 1)

	key, ok := keys[0].(map[string]interface{})
	require.True(t, ok)

	keyID := key["kid"].(string)
	modulusStr := key["n"].(string)

	// Decode modulus from base64url
	modulus, err := base64.RawURLEncoding.DecodeString(modulusStr)
	require.NoError(t, err)

	// Verify modulus matches the public key
	publicKey := &privateKey.PublicKey
	assert.Equal(t, publicKey.N.Bytes(), modulus)

	// Create a token and verify it uses the same key ID
	tokenRequest := map[string]interface{}{
		"subject": "test-user",
	}

	response, err := idp.HandleTokenRequest(context.Background(), tokenRequest)
	require.NoError(t, err)

	tokenStr := response["access_token"].(string)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return privateKey.Public(), nil
	})
	require.NoError(t, err)

	assert.Equal(t, keyID, token.Header["kid"])
}

// Integration-style tests

func TestTerraformIDP_FullOAuthFlow(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      30 * time.Minute,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	// Step 1: Get OpenID Configuration
	openidConfig := idp.GetOpenIDConfiguration()
	assert.NotNil(t, openidConfig)

	// Step 2: Get JWKS
	_, err = idp.GetJWKS()
	require.NoError(t, err)

	// Step 3: Handle token request (simulating authorization code exchange)
	tokenRequest := map[string]interface{}{
		"subject": "integration-test-user",
		"name":    "Integration Test User",
		"email":   "integration@example.com",
		"groups":  []string{"testers", "developers"},
	}

	tokenResponse, err := idp.HandleTokenRequest(context.Background(), tokenRequest)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenResponse["access_token"])

	// Step 4: Validate the token was properly signed
	accessToken := tokenResponse["access_token"].(string)
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return privateKey.Public(), nil
	})
	require.NoError(t, err)
	require.True(t, token.Valid)

	// Step 5: Get user info from token
	userInfo, err := idp.HandleUserInfoRequest(context.Background(), accessToken)
	require.NoError(t, err)

	assert.Equal(t, "integration-test-user", userInfo["sub"])
	assert.Equal(t, "Integration Test User", userInfo["name"])
	assert.Equal(t, "integration@example.com", userInfo["email"])

	// Step 6: Verify token validation works
	claims, err := idp.ValidateToken(context.Background(), accessToken)
	require.NoError(t, err)
	assert.Equal(t, "integration-test-user", claims["sub"])
	assert.Equal(t, config.IssuerURL, claims["iss"])

	// Step 7: Test token expiration
	expiredClaims := jwt.MapClaims{
		"iss": config.IssuerURL,
		"sub": "expired-user",
		"aud": "terraform",
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
		"nbf": time.Now().Add(-2 * time.Hour).Unix(),
	}

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodRS256, expiredClaims)
	expiredToken.Header["kid"] = idp.GetKeyID()
	expiredTokenStr, err := expiredToken.SignedString(privateKey)
	require.NoError(t, err)

	// Should fail to validate expired token
	_, err = idp.ValidateToken(context.Background(), expiredTokenStr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Token is expired")
}

func TestTerraformIDP_MultipleTokensDifferentUsers(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	users := []map[string]interface{}{
		{
			"subject": "user1",
			"name":    "User One",
			"email":   "user1@example.com",
		},
		{
			"subject": "user2",
			"name":    "User Two",
			"email":   "user2@example.com",
		},
		{
			"subject": "user3",
			"name":    "User Three",
			"email":   "user3@example.com",
		},
	}

	tokens := make([]string, len(users))

	// Create tokens for different users
	for i, user := range users {
		response, err := idp.HandleTokenRequest(context.Background(), user)
		require.NoError(t, err)
		tokens[i] = response["access_token"].(string)
	}

	// Verify each token contains correct user information
	for i, token := range tokens {
		userInfo, err := idp.HandleUserInfoRequest(context.Background(), token)
		require.NoError(t, err)

		expectedUser := users[i]
		assert.Equal(t, expectedUser["subject"], userInfo["sub"])
		assert.Equal(t, expectedUser["name"], userInfo["name"])
		assert.Equal(t, expectedUser["email"], userInfo["email"])

		// Verify claims are also correct
		claims, err := idp.ValidateToken(context.Background(), token)
		require.NoError(t, err)
		assert.Equal(t, expectedUser["subject"], claims["sub"])
	}
}

func TestTerraformIDP_TokenSecurity(t *testing.T) {
	privateKey := generateTestRSAKey(t)
	config := TerraformIDPConfig{
		IssuerURL:            "https://terrareg.example.com",
		ClientID:             "test-client",
		RedirectURIs:         []string{"http://localhost:3000/callback"},
		TokenExpiration:      time.Hour,
		AllowUnsafeRedirects: false,
	}

	idp, err := NewTerraformIDP(config, privateKey)
	require.NoError(t, err)

	tokenRequest := map[string]interface{}{
		"subject": "security-test-user",
	}

	response, err := idp.HandleTokenRequest(context.Background(), tokenRequest)
	require.NoError(t, err)

	token := response["access_token"].(string)

	// Test 1: Token should not be verifiable with a different public key
	differentPrivateKey := generateTestRSAKey(t)
	_, err = jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return differentPrivateKey.Public(), nil
	})
	assert.Error(t, err, "Token should not be verifiable with different key")

	// Test 2: Token tampering should be detected
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3, "JWT should have 3 parts")

	// Tamper with the payload
	tamperedToken := parts[0] + "." + "tampered.payload" + "." + parts[2]
	_, err = jwt.Parse(tamperedToken, func(token *jwt.Token) (interface{}, error) {
		return privateKey.Public(), nil
	})
	assert.Error(t, err, "Tampered token should be invalid")

	// Test 3: Token should have proper structure
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return privateKey.Public(), nil
	})
	require.NoError(t, err)

	// Verify header
	assert.Equal(t, "RS256", parsedToken.Header["alg"])
	assert.Equal(t, idp.keyID, parsedToken.Header["kid"])

	// Verify essential claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Contains(t, claims, "iss")
	assert.Contains(t, claims, "sub")
	assert.Contains(t, claims, "aud")
	assert.Contains(t, claims, "exp")
	assert.Contains(t, claims, "iat")
	assert.Contains(t, claims, "nbf")
}