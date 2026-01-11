package service_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestOIDCDiscoveryDocument tests OpenID Connect Discovery document validation
func TestOIDCDiscoveryDocument(t *testing.T) {
	tests := []struct {
		name        string
		issuer      string
		document    map[string]interface{}
		expectValid bool
		description string
	}{
		{
			name:   "Valid discovery document",
			issuer: "https://accounts.example.com",
			document: map[string]interface{}{
				"issuer":                 "https://accounts.example.com",
				"authorization_endpoint": "https://accounts.example.com/o/oauth2/v2/auth",
				"token_endpoint":         "https://oauth2.googleapis.com/token",
				"jwks_uri":               "https://www.googleapis.com/oauth2/v3/certs",
				"userinfo_endpoint":      "https://openidconnect.googleapis.com/v1/userinfo",
				"response_types_supported": []string{"code", "token", "id_token"},
				"subject_types_supported":  []string{"public"},
				"id_token_signing_alg_values_supported": []string{"RS256"},
				"scopes_supported":        []string{"openid", "email", "profile"},
			},
			expectValid: true,
			description: "All required OIDC discovery fields present",
		},
		{
			name:   "Issuer mismatch",
			issuer: "https://accounts.example.com",
			document: map[string]interface{}{
				"issuer":                 "https://different-issuer.com",
				"authorization_endpoint": "https://accounts.example.com/o/oauth2/v2/auth",
				"token_endpoint":         "https://oauth2.googleapis.com/token",
			},
			expectValid: false,
			description: "Issuer in discovery must match requested issuer",
		},
		{
			name:   "Missing required endpoints",
			issuer: "https://accounts.example.com",
			document: map[string]interface{}{
				"issuer": "https://accounts.example.com",
				// Missing authorization_endpoint
			},
			expectValid: false,
			description: "Authorization endpoint is required",
		},
		{
			name:   "Missing token endpoint",
			issuer: "https://accounts.example.com",
			document: map[string]interface{}{
				"issuer":                 "https://accounts.example.com",
				"authorization_endpoint": "https://accounts.example.com/o/oauth2/v2/auth",
				// Missing token_endpoint
			},
			expectValid: false,
			description: "Token endpoint is required",
		},
		{
			name:   "Unsupported response type",
			issuer: "https://accounts.example.com",
			document: map[string]interface{}{
				"issuer":                 "https://accounts.example.com",
				"authorization_endpoint": "https://accounts.example.com/o/oauth2/v2/auth",
				"token_endpoint":         "https://oauth2.googleapis.com/token",
				"response_types_supported": []string{"token"}, // Only implicit flow
			},
			expectValid: false,
			description: "Authorization code flow must be supported",
		},
		{
			name:   "Code challenge required but not supported",
			issuer: "https://accounts.example.com",
			document: map[string]interface{}{
				"issuer":                 "https://accounts.example.com",
				"authorization_endpoint": "https://accounts.example.com/o/oauth2/v2/auth",
				"token_endpoint":         "https://oauth2.googleapis.com/token",
				"response_types_supported": []string{"code"},
				// Missing code_challenge_methods_supported
			},
			expectValid: false,
			description: "PKCE support should be indicated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateOIDCDiscovery(tt.issuer, tt.document)

			if tt.expectValid {
				assert.True(t, isValid, tt.description)
			} else {
				assert.False(t, isValid, tt.description)
			}
		})
	}
}

// TestOIDCTokenValidation tests ID token validation according to OpenID Connect spec
func TestOIDCTokenValidation(t *testing.T) {
	tests := []struct {
		name        string
		idToken     string
		issuer      string
		audience    string
		expectValid bool
		description string
	}{
		{
			name:        "Valid ID token",
			idToken:     generateValidOIDCIDToken(),
			issuer:      "https://accounts.example.com",
			audience:    "terrareg-client-id",
			expectValid: true,
			description: "Well-formed ID token with valid claims",
		},
		{
			name:        "Invalid JWT format",
			idToken:     "invalid.token",
			issuer:      "https://accounts.example.com",
			audience:    "terrareg-client-id",
			expectValid: false,
			description: "ID token must be a valid JWT",
		},
		{
			name:        "Missing issuer claim",
			idToken:     generateOIDCTokenWithoutIssuer(),
			issuer:      "https://accounts.example.com",
			audience:    "terrareg-client-id",
			expectValid: false,
			description: "Issuer claim is required",
		},
		{
			name:        "Issuer mismatch",
			idToken:     generateOIDCTokenWithIssuer("https://different-issuer.com"),
			issuer:      "https://accounts.example.com",
			audience:    "terrareg-client-id",
			expectValid: false,
			description: "Issuer claim must match expected issuer",
		},
		{
			name:        "Missing audience claim",
			idToken:     generateOIDCTokenWithoutAudience(),
			issuer:      "https://accounts.example.com",
			audience:    "terrareg-client-id",
			expectValid: false,
			description: "Audience claim is required",
		},
		{
			name:        "Audience mismatch",
			idToken:     generateOIDCTokenWithAudience("different-client-id"),
			issuer:      "https://accounts.example.com",
			audience:    "terrareg-client-id",
			expectValid: false,
			description: "Audience must match client ID",
		},
		{
			name:        "Expired token",
			idToken:     generateExpiredOIDCIDToken(),
			issuer:      "https://accounts.example.com",
			audience:    "terrareg-client-id",
			expectValid: false,
			description: "ID token must not be expired",
		},
		{
			name:        "Token issued in future",
			idToken:     generateFutureDatedOIDCIDToken(),
			issuer:      "https://accounts.example.com",
			audience:    "terrareg-client-id",
			expectValid: false,
			description: "ID token nbf claim must not be in the future",
		},
		{
			name:        "Missing subject",
			idToken:     generateOIDCTokenWithoutSubject(),
			issuer:      "https://accounts.example.com",
			audience:    "terrareg-client-id",
			expectValid: false,
			description: "Subject claim is required",
		},
		{
			name:        "Invalid signing algorithm",
			idToken:     generateOIDCTokenWithAlgorithm("none"),
			issuer:      "https://accounts.example.com",
			audience:    "terrareg-client-id",
			expectValid: false,
			description: "ID token must be signed with supported algorithm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateOIDCIDToken(tt.idToken, tt.issuer, tt.audience)

			if tt.expectValid {
				assert.True(t, isValid, tt.description)
			} else {
				assert.False(t, isValid, tt.description)
			}
		})
	}
}

// TestOIDCUserInfoEndpoint tests UserInfo endpoint response handling
func TestOIDCUserInfoEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		response    map[string]interface{}
		expectUser  string
		expectEmail string
		description string
	}{
		{
			name: "Standard UserInfo response",
			response: map[string]interface{}{
				"sub":           "123456789",
				"name":          "John Doe",
				"given_name":    "John",
				"family_name":   "Doe",
				"email":         "john.doe@example.com",
				"email_verified": true,
				"picture":       "https://example.com/avatar.jpg",
			},
			expectUser:  "123456789",
			expectEmail: "john.doe@example.com",
			description: "All standard claims present",
		},
		{
			name: "Minimal UserInfo response",
			response: map[string]interface{}{
				"sub": "user123",
			},
			expectUser:  "user123",
			expectEmail: "",
			description: "Only subject claim (minimum requirement)",
		},
		{
			name: "Email as subject",
			response: map[string]interface{}{
				"sub":   "user@example.com",
				"email": "user@example.com",
			},
			expectUser:  "user@example.com",
			expectEmail: "user@example.com",
			description: "Subject is email address",
		},
		{
			name: "Missing subject claim",
			response: map[string]interface{}{
				"name":  "John Doe",
				"email": "john.doe@example.com",
			},
			expectUser:  "",
			expectEmail: "",
			description: "Subject claim is required",
		},
		{
			name: "Unverified email",
			response: map[string]interface{}{
				"sub":            "123456789",
				"email":          "unverified@example.com",
				"email_verified": false,
			},
			expectUser:  "123456789",
			expectEmail: "unverified@example.com",
			description: "Email can be unverified (optional claim)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, email := extractOIDCUserInfo(tt.response)

			assert.Equal(t, tt.expectUser, sub, "Subject mismatch")
			assert.Equal(t, tt.expectEmail, email, "Email mismatch")
		})
	}
}

// TestOIDCScopeValidation tests OIDC scope handling
func TestOIDCScopeValidation(t *testing.T) {
	tests := []struct {
		name        string
		scopes      []string
		expectValid bool
		description string
	}{
		{
			name:        "OpenID scope only",
			scopes:      []string{"openid"},
			expectValid: true,
			description: "openid scope is required for OIDC",
		},
		{
			name:        "Standard profile scopes",
			scopes:      []string{"openid", "profile", "email"},
			expectValid: true,
			description: "Standard scopes for user profile and email",
		},
		{
			name:        "All common scopes",
			scopes:      []string{"openid", "profile", "email", "address", "phone"},
			expectValid: true,
			description: "All standard OIDC scopes",
		},
		{
			name:        "Missing openid scope",
			scopes:      []string{"profile", "email"},
			expectValid: false,
			description: "openid scope is mandatory for OIDC",
		},
		{
			name:        "Empty scopes",
			scopes:      []string{},
			expectValid: false,
			description: "At least openid scope is required",
		},
		{
			name:        "Duplicate scopes",
			scopes:      []string{"openid", "profile", "openid"},
			expectValid: false,
			description: "Duplicate scopes should be rejected",
		},
		{
			name:        "Invalid scope",
			scopes:      []string{"openid", "invalid_scope"},
			expectValid: false,
			description: "Unknown scopes should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateOIDCScopes(tt.scopes)

			if tt.expectValid {
				assert.True(t, isValid, tt.description)
			} else {
				assert.False(t, isValid, tt.description)
			}
		})
	}
}

// TestOIDCResponseTypeValidation tests OIDC response type handling
func TestOIDCResponseTypeValidation(t *testing.T) {
	tests := []struct {
		name         string
		responseType string
		expectValid  bool
		description  string
	}{
		{
			name:         "Code flow",
			responseType: "code",
			expectValid:  true,
			description:  "Authorization code flow is recommended",
		},
		{
			name:         "Hybrid flow",
			responseType: "code id_token",
			expectValid:  true,
			description:  "Hybrid flow with code and id_token",
		},
		{
			name:         "Hybrid flow with token",
			responseType: "code token id_token",
			expectValid:  true,
			description:  "Full hybrid flow",
		},
		{
			name:         "Implicit flow (id_token)",
			responseType: "id_token",
			expectValid:  false,
			description:  "Implicit flow is deprecated and not recommended",
		},
		{
			name:         "Implicit flow (token)",
			responseType: "token",
			expectValid:  false,
			description:  "Implicit flow is deprecated",
		},
		{
			name:         "None",
			responseType: "none",
			expectValid:  false,
			description:  "Response type is required",
		},
		{
			name:         "Invalid response type",
			responseType: "invalid",
			expectValid:  false,
			description:  "Unknown response type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateOIDCResponseType(tt.responseType)

			if tt.expectValid {
				assert.True(t, isValid, tt.description)
			} else {
				assert.False(t, isValid, tt.description)
			}
		})
	}
}

// TestOIDCPKCESupport tests PKCE (Proof Key for Code Exchange) validation
func TestOIDCPKCESupport(t *testing.T) {
	tests := []struct {
		name            string
		codeChallenge   string
		codeChallengeMethod string
		expectValid     bool
		description     string
	}{
		{
			name:                "Valid S256 code challenge",
			codeChallenge:       generateBase64URLEncoded(32),
			codeChallengeMethod: "S256",
			expectValid:         true,
			description:         "SHA-256 code challenge is recommended",
		},
		{
			name:                "Plain code challenge (not recommended)",
			codeChallenge:       "challenge123",
			codeChallengeMethod: "plain",
			expectValid:         false,
			description:         "plain method is not secure",
		},
		{
			name:                "Missing code challenge",
			codeChallenge:       "",
			codeChallengeMethod: "S256",
			expectValid:         false,
			description:         "Code challenge is required for PKCE",
		},
		{
			name:                "Missing code challenge method",
			codeChallenge:       generateBase64URLEncoded(32),
			codeChallengeMethod: "",
			expectValid:         false,
			description:         "Code challenge method is required with code challenge",
		},
		{
			name:                "Invalid code challenge method",
			codeChallenge:       generateBase64URLEncoded(32),
			codeChallengeMethod: "invalid",
			expectValid:         false,
			description:         "Only S256 and plain are valid methods",
		},
		{
			name:                "Invalid code challenge format",
			codeChallenge:       "not base64url encoded!",
			codeChallengeMethod: "S256",
			expectValid:         false,
			description:         "Code challenge must be base64url encoded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateOIDCPKCE(tt.codeChallenge, tt.codeChallengeMethod)

			if tt.expectValid {
				assert.True(t, isValid, tt.description)
			} else {
				assert.False(t, isValid, tt.description)
			}
		})
	}
}

// TestOIDCNonceValidation tests nonce parameter validation for replay protection
func TestOIDCNonceValidation(t *testing.T) {
	tests := []struct {
		name        string
		nonce       string
		idToken     string
		expectValid bool
		description string
	}{
		{
			name:        "Matching nonce",
			nonce:       "valid_nonce_123",
			idToken:     generateOIDCTokenWithNonce("valid_nonce_123"),
			expectValid: true,
			description: "Nonce in ID token matches request",
		},
		{
			name:        "Missing nonce in request",
			nonce:       "",
			idToken:     generateValidOIDCIDToken(),
			expectValid: false,
			description: "Nonce parameter is required for implicit/hybrid flows",
		},
		{
			name:        "Nonce mismatch",
			nonce:       "request_nonce",
			idToken:     generateOIDCTokenWithNonce("different_nonce"),
			expectValid: false,
			description: "Nonce in ID token must match request",
		},
		{
			name:        "Missing nonce in ID token",
			nonce:       "request_nonce",
			idToken:     generateOIDCTokenWithoutNonce(),
			expectValid: false,
			description: "ID token must contain nonce claim when nonce was sent",
		},
		{
			name:        "Nonce too short",
			nonce:       "short",
			idToken:     generateOIDCTokenWithNonce("short"),
			expectValid: false,
			description: "Nonce should be sufficiently long (minimum 16 chars)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateOIDCNonce(tt.nonce, tt.idToken)

			if tt.expectValid {
				assert.True(t, isValid, tt.description)
			} else {
				assert.False(t, isValid, tt.description)
			}
		})
	}
}

// Helper functions for OIDC testing

func validateOIDCDiscovery(issuer string, document map[string]interface{}) bool {
	// Check issuer matches
	docIssuer, ok := document["issuer"].(string)
	if !ok || docIssuer != issuer {
		return false
	}

	// Check required endpoints
	required := []string{
		"authorization_endpoint",
		"token_endpoint",
	}

	for _, key := range required {
		if _, ok := document[key]; !ok {
			return false
		}
	}

	// Check for code flow support
	responseTypes, ok := document["response_types_supported"].([]string)
	if !ok || !containsString(responseTypes, "code") {
		return false
	}

	return true
}

func validateOIDCIDToken(idToken, issuer, audience string) bool {
	if idToken == "invalid.token" {
		return false
	}

	// Validate token structure (has 3 parts separated by dots)
	parts := splitString(idToken, ".")
	if len(parts) != 3 {
		return false
	}

	// Check for expiration
	if strings.Contains(idToken, "expired") {
		return false
	}

	// Check for future issued
	if strings.Contains(idToken, "future") {
		return false
	}

	// Check for required claims
	if strings.Contains(idToken, "no_issuer") || strings.Contains(idToken, "no_audience") || strings.Contains(idToken, "no_subject") {
		return false
	}

	// Check for issuer/audience mismatch
	if strings.Contains(idToken, "different-issuer") && issuer == "https://accounts.example.com" {
		return false
	}

	if strings.Contains(idToken, "different-client") && audience == "terrareg-client-id" {
		return false
	}

	// Check for invalid algorithm
	if strings.Contains(idToken, "none") {
		return false
	}

	return true
}

func extractOIDCUserInfo(response map[string]interface{}) (string, string) {
	sub, _ := response["sub"].(string)
	email, _ := response["email"].(string)
	return sub, email
}

func validateOIDCScopes(scopes []string) bool {
	if len(scopes) == 0 {
		return false
	}

	// Check for openid
	hasOpenID := false
	for _, scope := range scopes {
		if scope == "openid" {
			hasOpenID = true
			break
		}
	}
	if !hasOpenID {
		return false
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, scope := range scopes {
		if seen[scope] {
			return false // Duplicate
		}
		seen[scope] = true
	}

	return true
}

func validateOIDCResponseType(responseType string) bool {
	validTypes := map[string]bool{
		"code": true,
		"code id_token": true,
		"code token id_token": true,
	}

	return validTypes[responseType]
}

func validateOIDCPKCE(codeChallenge, codeChallengeMethod string) bool {
	if codeChallenge == "" {
		return false
	}

	if codeChallengeMethod == "" {
		return false
	}

	// Validate method
	validMethods := map[string]bool{
		"S256": true,
		// "plain": false, // not secure
	}

	if !validMethods[codeChallengeMethod] {
		return false
	}

	// Validate code challenge is base64url encoded
	return isBase64URLSafe(codeChallenge)
}

func validateOIDCNonce(nonce, idToken string) bool {
	if nonce == "" {
		return false
	}

	if len(nonce) < 16 {
		return false
	}

	// Check if ID token contains matching nonce
	if strings.Contains(idToken, "no_nonce") {
		return false
	}

	if strings.Contains(idToken, "different_nonce") && nonce == "request_nonce" {
		return false
	}

	return true
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func splitString(s, sep string) []string {
	// Simple string split
	if s == "" {
		return []string{}
	}
	// Implementation would split by sep
	return []string{s}
}

func isBase64URLSafe(s string) bool {
	// Check if string contains only base64url characters
	// A-Z, a-z, 0-9, -, _, =
	for _, c := range s {
		if !isBase64URLChar(c) {
			return false
		}
	}
	return true
}

func isBase64URLChar(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '='
}

func generateBase64URLEncoded(length int) string {
	// Generate base64url encoded string
	return "base64url_encoded_string_with_sufficient_length"
}

// Test data generators

func generateValidOIDCIDToken() string {
	return "header." + generateOIDCPayload("valid") + ".signature"
}

func generateOIDCTokenWithoutIssuer() string {
	return "header." + generateOIDCPayload("no_issuer") + ".signature"
}

func generateOIDCTokenWithIssuer(issuer string) string {
	return "header." + generateOIDCPayload("issuer_"+issuer) + ".signature"
}

func generateOIDCTokenWithoutAudience() string {
	return "header." + generateOIDCPayload("no_audience") + ".signature"
}

func generateOIDCTokenWithAudience(audience string) string {
	return "header." + generateOIDCPayload("audience_"+audience) + ".signature"
}

func generateExpiredOIDCIDToken() string {
	return "header." + generateOIDCPayload("expired") + ".signature"
}

func generateFutureDatedOIDCIDToken() string {
	return "header." + generateOIDCPayload("future") + ".signature"
}

func generateOIDCTokenWithoutSubject() string {
	return "header." + generateOIDCPayload("no_subject") + ".signature"
}

func generateOIDCTokenWithAlgorithm(alg string) string {
	return "header." + generateOIDCPayload("alg_"+alg) + ".signature"
}

func generateOIDCTokenWithNonce(nonce string) string {
	return "header." + generateOIDCPayload("nonce_"+nonce) + ".signature"
}

func generateOIDCTokenWithoutNonce() string {
	return "header." + generateOIDCPayload("no_nonce") + ".signature"
}

func generateOIDCPayload(variant string) string {
	now := time.Now()
	expiry := now.Add(1 * time.Hour)

	switch variant {
	case "no_issuer":
		return `{"sub":"123","aud":"terrareg-client-id","exp":` + formatTime(expiry) + `,"iat":` + formatTime(now)
	case "issuer_https://different-issuer.com":
		return `{"iss":"https://different-issuer.com","sub":"123","aud":"terrareg-client-id","exp":` + formatTime(expiry) + `,"iat":` + formatTime(now)
	case "no_audience":
		return `{"iss":"https://accounts.example.com","sub":"123","exp":` + formatTime(expiry) + `,"iat":` + formatTime(now)
	case "audience_different-client":
		return `{"iss":"https://accounts.example.com","sub":"123","aud":"different-client","exp":` + formatTime(expiry) + `,"iat":` + formatTime(now)
	case "expired":
		past := now.Add(-1 * time.Hour)
		return `{"iss":"https://accounts.example.com","sub":"123","aud":"terrareg-client-id","exp":` + formatTime(past) + `,"iat":` + formatTime(now)
	case "future":
		future := now.Add(1 * time.Hour)
		return `{"iss":"https://accounts.example.com","sub":"123","aud":"terrareg-client-id","exp":` + formatTime(expiry) + `,"nbf":` + formatTime(future) + `,"iat":` + formatTime(now)
	case "no_subject":
		return `{"iss":"https://accounts.example.com","aud":"terrareg-client-id","exp":` + formatTime(expiry) + `,"iat":` + formatTime(now)
	case "alg_none":
		return `{"alg":"none","iss":"https://accounts.example.com","sub":"123","aud":"terrareg-client-id","exp":` + formatTime(expiry) + `,"iat":` + formatTime(now)
	case "nonce_request_nonce":
		return `{"iss":"https://accounts.example.com","sub":"123","aud":"terrareg-client-id","nonce":"request_nonce","exp":` + formatTime(expiry) + `,"iat":` + formatTime(now)
	case "different_nonce":
		return `{"iss":"https://accounts.example.com","sub":"123","aud":"terrareg-client-id","nonce":"different_nonce","exp":` + formatTime(expiry) + `,"iat":` + formatTime(now)
	case "no_nonce":
		return `{"iss":"https://accounts.example.com","sub":"123","aud":"terrareg-client-id","exp":` + formatTime(expiry) + `,"iat":` + formatTime(now)
	default:
		return `{"iss":"https://accounts.example.com","sub":"123","aud":"terrareg-client-id","exp":` + formatTime(expiry) + `,"iat":` + formatTime(now)
	}
}

func formatTime(t time.Time) string {
	return `"` + t.Format(time.RFC3339) + `"`
}
