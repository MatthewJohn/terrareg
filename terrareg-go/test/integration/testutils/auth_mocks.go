package testutils

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"golang.org/x/oauth2"
)

// MockOIDCServer provides a mock OIDC provider server for testing
type MockOIDCServer struct {
	Server       *httptest.Server
	DiscoveryURL string
	TokenURL     string
	UserInfoURL  string
	ClientID     string
	ClientSecret string

	// Test data that will be returned
	TestUserInfo *service.OIDCUserInfo
	TestIDToken  string

	// Optional: custom handlers
	CustomTokenHandler    http.HandlerFunc
	CustomUserInfoHandler http.HandlerFunc
}

// NewMockOIDCServer creates a new mock OIDC server
func NewMockOIDCServer() *MockOIDCServer {
	server := &MockOIDCServer{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		TestUserInfo: &service.OIDCUserInfo{
			Subject:       "test-user-123",
			Name:          "Test User",
			Username:      "testuser",
			Email:         "test@example.com",
			EmailVerified: true,
			Groups:        []string{"test-group", "admins"},
		},
	}

	mux := http.NewServeMux()
	server.Server = httptest.NewServer(mux)

	// Set URLs
	baseURL := server.Server.URL
	server.DiscoveryURL = fmt.Sprintf("%s/.well-known/openid-configuration", baseURL)
	server.TokenURL = fmt.Sprintf("%s/oauth/token", baseURL)
	server.UserInfoURL = fmt.Sprintf("%s/oauth/userinfo", baseURL)

	// Register handlers
	mux.HandleFunc("/.well-known/openid-configuration", server.handleDiscovery)
	mux.HandleFunc("/oauth/token", server.handleToken)
	mux.HandleFunc("/oauth/userinfo", server.handleUserInfo)
	mux.HandleFunc("/keys", server.handleKeys)

	return server
}

// handleDiscovery returns OIDC discovery document
func (s *MockOIDCServer) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"issuer":                 s.Server.URL,
		"authorization_endpoint": fmt.Sprintf("%s/oauth/auth", s.Server.URL),
		"token_endpoint":         s.TokenURL,
		"userinfo_endpoint":      s.UserInfoURL,
		"jwks_uri":               fmt.Sprintf("%s/keys", s.Server.URL),
		"response_types_supported": []string{"code"},
		"subject_types_supported":  []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":        []string{"openid", "profile", "email"},
	})
}

// handleToken returns mock OAuth tokens
func (s *MockOIDCServer) handleToken(w http.ResponseWriter, r *http.Request) {
	if s.CustomTokenHandler != nil {
		s.CustomTokenHandler(w, r)
		return
	}

	// Verify basic auth
	clientID, clientSecret, _ := r.BasicAuth()
	if clientID != s.ClientID || clientSecret != s.ClientSecret {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": "mock-access-token",
		"token_type":   "Bearer",
		"expires_in":   3600,
		"id_token":     s.generateMockIDToken(),
		"refresh_token": "mock-refresh-token",
	})
}

// handleUserInfo returns mock user info
func (s *MockOIDCServer) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	if s.CustomUserInfoHandler != nil {
		s.CustomUserInfoHandler(w, r)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.TestUserInfo)
}

// handleKeys returns mock JWKS
func (s *MockOIDCServer) handleKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kid": "test-key-id",
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"n":   "mock-n-value",
				"e":   "AQAB",
			},
		},
	})
}

// generateMockIDToken generates a mock JWT ID token
func (s *MockOIDCServer) generateMockIDToken() string {
	// In a real implementation, this would generate a proper signed JWT
	// For testing purposes, we return a mock token that the tests can verify
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","kid":"test-key-id"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{
		"iss": "%s",
		"sub": "%s",
		"aud": "%s",
		"exp": %d,
		"iat": %d,
		"name": "%s",
		"email": "%s",
		"email_verified": true
	}`, s.Server.URL, s.TestUserInfo.Subject, s.ClientID,
		time.Now().Add(time.Hour).Unix(), time.Now().Unix(),
		s.TestUserInfo.Name, s.TestUserInfo.Email)))
	signature := "mock-signature"
	return fmt.Sprintf("%s.%s.%s", header, payload, signature)
}

// Close closes the mock server
func (s *MockOIDCServer) Close() {
	s.Server.Close()
}

// MockSAMLServer provides a mock SAML IDP server for testing
type MockSAMLServer struct {
	Server        *httptest.Server
	MetadataURL   string
	SSOURL        string
	EntityID      string
	SPEntityID    string

	// Test configuration
	TestUsername  string
	TestEmail     string
	TestName      string
	TestGroups    []string

	// SAML keys
	privateKey *rsa.PrivateKey
	certificate *x509.Certificate

	// Optional: custom handlers
	CustomMetadataHandler http.HandlerFunc
	CustomSSOHandler      http.HandlerFunc
}

// NewMockSAMLServer creates a new mock SAML IDP server
func NewMockSAMLServer() (*MockSAMLServer, error) {
	server := &MockSAMLServer{
		EntityID:     "https://idp.example.com/saml",
		SPEntityID:   "https://sp.example.com/saml",
		TestUsername: "testuser",
		TestEmail:    "test@example.com",
		TestName:     "Test User",
		TestGroups:   []string{"test-group", "admins"},
	}

	// Generate test keys
	if err := server.generateKeys(); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	server.Server = httptest.NewServer(mux)

	// Set URLs
	baseURL := server.Server.URL
	server.MetadataURL = fmt.Sprintf("%s/saml/metadata", baseURL)
	server.SSOURL = fmt.Sprintf("%s/saml/sso", baseURL)

	// Register handlers
	mux.HandleFunc("/saml/metadata", server.handleMetadata)
	mux.HandleFunc("/saml/sso", server.handleSSO)

	return server, nil
}

// generateKeys generates RSA keys for SAML
func (s *MockSAMLServer) generateKeys() error {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	s.privateKey = privateKey

	// Generate self-signed certificate
	serialNumber := big.NewInt(1)
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: "Test SAML IDP"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	s.certificate, err = x509.ParseCertificate(certBytes)
	if err != nil {
		return err
	}

	return nil
}

// handleMetadata returns SAML metadata
func (s *MockSAMLServer) handleMetadata(w http.ResponseWriter, r *http.Request) {
	if s.CustomMetadataHandler != nil {
		s.CustomMetadataHandler(w, r)
		return
	}

	metadata := saml.EntityDescriptor{
		EntityID: s.EntityID,
		IDPSSODescriptors: []saml.IDPSSODescriptor{
			{
				SSODescriptor: saml.SSODescriptor{
					RoleDescriptor: saml.RoleDescriptor{
						ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
					},
				},
				SingleSignOnServices: []saml.Endpoint{
					{
						Binding:  saml.HTTPRedirectBinding,
						Location: s.SSOURL,
					},
					{
						Binding:  saml.HTTPPostBinding,
						Location: s.SSOURL,
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	xml.NewEncoder(w).Encode(metadata)
}

// handleSSO handles SAML SSO requests
func (s *MockSAMLServer) handleSSO(w http.ResponseWriter, r *http.Request) {
	if s.CustomSSOHandler != nil {
		s.CustomSSOHandler(w, r)
		return
	}

	// Parse SAML request
	samlRequest := r.URL.Query().Get("SAMLRequest")
	if samlRequest == "" {
		http.Error(w, "missing SAMLRequest", http.StatusBadRequest)
		return
	}

	// Decode and parse request
	decodedRequest, err := base64.StdEncoding.DecodeString(samlRequest)
	if err != nil {
		http.Error(w, "invalid SAMLRequest encoding", http.StatusBadRequest)
		return
	}

	var authReq saml.AuthnRequest
	if err := xml.Unmarshal(decodedRequest, &authReq); err != nil {
		http.Error(w, "invalid SAMLRequest XML", http.StatusBadRequest)
		return
	}

	// Create SAML response
	response := saml.Response{
		Destination:  authReq.Issuer.Value,
		ID:           fmt.Sprintf("_%d", time.Now().UnixNano()),
		InResponseTo: authReq.ID,
		IssueInstant: saml.TimeNow(),
		Version:      "2.0",
		Issuer: &saml.Issuer{
			Value: s.EntityID,
		},
		Status: saml.Status{
			StatusCode: saml.StatusCode{
				Value: saml.StatusSuccess,
			},
		},
		Assertion: samlAssertion(s.SPEntityID, s.TestUsername, s.TestEmail, s.TestName, s.TestGroups),
	}

	// Sign and encode response
	responseXML, err := xml.Marshal(response)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}

	encodedResponse := base64.StdEncoding.EncodeToString(responseXML)

	// Redirect back to ACS with SAML response
	relayState := r.URL.Query().Get("RelayState")
	redirectURL := fmt.Sprintf("%s?SAMLResponse=%s", authReq.AssertionConsumerServiceURL, url.QueryEscape(encodedResponse))
	if relayState != "" {
		redirectURL += "&RelayState=" + url.QueryEscape(relayState)
	}

	w.Header().Set("Location", redirectURL)
	w.WriteHeader(http.StatusFound)
}

// samlAssertion creates a test SAML assertion
func samlAssertion(spEntityID, username, email, name string, groups []string) *saml.Assertion {
	now := saml.TimeNow()
	notOnOrAfter := now.Add(8 * time.Hour)
	assertion := &saml.Assertion{
		ID:           fmt.Sprintf("_%d", now.UnixNano()),
		IssueInstant: now,
		Version:      "2.0",
		Issuer: saml.Issuer{
			Format: "urn:oasis:names:tc:SAML:2.0:nameid-format:entity",
			Value:  "https://idp.example.com/saml",
		},
		Subject: &saml.Subject{
			NameID: &saml.NameID{
				Format:          "urn:oasis:names:tc:SAML:2.0:nameid-format:transient",
				NameQualifier:   "https://idp.example.com/saml",
				SPNameQualifier: spEntityID,
				Value:           username,
			},
			SubjectConfirmations: []saml.SubjectConfirmation{
				{
					Method: "urn:oasis:names:tc:SAML:2.0:cm:bearer",
					SubjectConfirmationData: &saml.SubjectConfirmationData{
						NotOnOrAfter: now.Add(5 * time.Minute),
						Recipient:    spEntityID + "/saml/acs",
					},
				},
			},
		},
		Conditions: &saml.Conditions{
			NotBefore:    now,
			NotOnOrAfter: now.Add(5 * time.Minute),
			AudienceRestrictions: []saml.AudienceRestriction{
				{
					Audience: saml.Audience{
						Value: spEntityID,
					},
				},
			},
		},
		AuthnStatements: []saml.AuthnStatement{
			{
				AuthnInstant:        now,
				SessionIndex:        fmt.Sprintf("_%d", now.UnixNano()),
				SessionNotOnOrAfter: &notOnOrAfter,
				AuthnContext: saml.AuthnContext{
					AuthnContextClassRef: &saml.AuthnContextClassRef{
						Value: "urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport",
					},
				},
			},
		},
		AttributeStatements: []saml.AttributeStatement{
			{
				Attributes: []saml.Attribute{
					{
						Name:       "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
						NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
						Values: []saml.AttributeValue{
							{Value: email},
						},
					},
					{
						Name:       "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
						NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
						Values: []saml.AttributeValue{
							{Value: name},
						},
					},
				},
			},
		},
	}

	// Add groups if present
	if len(groups) > 0 {
		groupAttr := saml.Attribute{
			Name:       "http://schemas.xmlsoap.org/claims/Group",
			NameFormat: "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
		}
		for _, group := range groups {
			groupAttr.Values = append(groupAttr.Values, saml.AttributeValue{Value: group})
		}
		assertion.AttributeStatements[0].Attributes = append(assertion.AttributeStatements[0].Attributes, groupAttr)
	}

	return assertion
}

// Close closes the mock server
func (s *MockSAMLServer) Close() {
	s.Server.Close()
}

// MockGitHubOAuthServer provides a mock GitHub OAuth server for testing
type MockGitHubOAuthServer struct {
	Server        *httptest.Server
	AuthorizeURL  string
	TokenURL      string
	UserURL       string
	ClientID      string
	ClientSecret  string

	// Test data that will be returned
	TestUserInfo *GitHubUserInfo

	// Optional: custom handlers
	CustomTokenHandler    http.HandlerFunc
	CustomUserHandler     http.HandlerFunc
}

// GitHubUserInfo represents GitHub user information
type GitHubUserInfo struct {
	Login     string   `json:"login"`
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	AvatarURL string   `json:"avatar_url"`
	Groups    []string `json:"groups"` // Mocked for testing
}

// NewMockGitHubOAuthServer creates a new mock GitHub OAuth server
func NewMockGitHubOAuthServer() *MockGitHubOAuthServer {
	server := &MockGitHubOAuthServer{
		ClientID:     "test-github-client-id",
		ClientSecret: "test-github-client-secret",
		TestUserInfo: &GitHubUserInfo{
			Login:     "testuser",
			ID:        12345,
			Name:      "Test User",
			Email:     "test@example.com",
			AvatarURL: "https://example.com/avatar.png",
			Groups:    []string{"test-group", "admins"},
		},
	}

	mux := http.NewServeMux()
	server.Server = httptest.NewServer(mux)

	// Set URLs
	baseURL := server.Server.URL
	server.AuthorizeURL = fmt.Sprintf("%s/login/oauth/authorize", baseURL)
	server.TokenURL = fmt.Sprintf("%s/login/oauth/access_token", baseURL)
	server.UserURL = fmt.Sprintf("%s/user", baseURL)

	// Register handlers
	mux.HandleFunc("/login/oauth/authorize", server.handleAuthorize)
	mux.HandleFunc("/login/oauth/access_token", server.handleToken)
	mux.HandleFunc("/user", server.handleUser)

	return server
}

// handleAuthorize handles GitHub OAuth authorize request
func (s *MockGitHubOAuthServer) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	// Get parameters
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")

	if clientID != s.ClientID {
		http.Error(w, "invalid client_id", http.StatusUnauthorized)
		return
	}

	// Redirect back with mock authorization code
	authCode := fmt.Sprintf("mock-auth-code-%d", time.Now().UnixNano())
	redirectURL, _ := url.Parse(redirectURI)
	query := redirectURL.Query()
	query.Set("code", authCode)
	query.Set("state", state)
	redirectURL.RawQuery = query.Encode()

	w.Header().Set("Location", redirectURL.String())
	w.WriteHeader(http.StatusFound)
}

// handleToken handles GitHub OAuth token request
func (s *MockGitHubOAuthServer) handleToken(w http.ResponseWriter, r *http.Request) {
	if s.CustomTokenHandler != nil {
		s.CustomTokenHandler(w, r)
		return
	}

	// Verify client credentials
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	if clientID != s.ClientID || clientSecret != s.ClientSecret {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unauthorized",
		})
		return
	}

	// Return mock token response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": "mock-github-access-token",
		"token_type":   "bearer",
		"scope":        "read:user,user:email",
	})
}

// handleUser handles GitHub user info request
func (s *MockGitHubOAuthServer) handleUser(w http.ResponseWriter, r *http.Request) {
	if s.CustomUserHandler != nil {
		s.CustomUserHandler(w, r)
		return
	}

	// Verify authorization
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Bad credentials",
		})
		return
	}

	// Return mock user info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.TestUserInfo)
}

// Close closes the mock server
func (s *MockGitHubOAuthServer) Close() {
	s.Server.Close()
}

// Helper function to create configured test context with mock OIDC
func CreateTestContextWithMockOIDC(ctx context.Context, mockServer *MockOIDCServer) context.Context {
	// Store mock server URL in context for test use
	return context.WithValue(ctx, "mock_oidc_server", mockServer.Server.URL)
}

// Helper function to create configured test context with mock SAML
func CreateTestContextWithMockSAML(ctx context.Context, mockServer *MockSAMLServer) context.Context {
	return context.WithValue(ctx, "mock_saml_server", mockServer.Server.URL)
}

// Helper function to create configured test context with mock GitHub OAuth
func CreateTestContextWithMockGitHub(ctx context.Context, mockServer *MockGitHubOAuthServer) context.Context {
	return context.WithValue(ctx, "mock_github_server", mockServer.Server.URL)
}

// MockAuthConfig returns a mock auth configuration for testing
func MockAuthConfigWithOIDC(oidcServer *MockOIDCServer) map[string]interface{} {
	return map[string]interface{}{
		"OPENID_CONNECT_ISSUER":        oidcServer.Server.URL,
		"OPENID_CONNECT_CLIENT_ID":     oidcServer.ClientID,
		"OPENID_CONNECT_CLIENT_SECRET": oidcServer.ClientSecret,
		"PUBLIC_URL":                   "https://test.example.com",
		"SECRET_KEY":                   "test-secret-key",
	}
}

// MockAuthConfigWithSAML returns a mock auth configuration for SAML testing
func MockAuthConfigWithSAML(samlServer *MockSAMLServer) map[string]interface{} {
	return map[string]interface{}{
		"SAML2_ENTITY_ID":         "https://sp.example.com/saml",
		"SAML2_IDP_METADATA_URL":  samlServer.MetadataURL,
		"SAML2_PRIVATE_KEY":       "", // Mock server handles signing
		"SAML2_PUBLIC_KEY":        "", // Mock server handles verification
		"PUBLIC_URL":              "https://test.example.com",
		"SECRET_KEY":              "test-secret-key",
	}
}

// MockAuthConfigWithGitHub returns a mock auth configuration for GitHub OAuth testing
func MockAuthConfigWithGitHub(githubServer *MockGitHubOAuthServer) map[string]interface{} {
	return map[string]interface{}{
		"GITHUB_OAUTH_CLIENT_ID":     githubServer.ClientID,
		"GITHUB_OAUTH_CLIENT_SECRET": githubServer.ClientSecret,
		"PUBLIC_URL":                 "https://test.example.com",
		"SECRET_KEY":                 "test-secret-key",
	}
}

// Helper to verify OAuth state parameter matches expected format
func ValidateOAuthState(state string) bool {
	// State should be a base64-encoded non-empty string
	if state == "" {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(state)
	return err == nil
}

// Helper to create a test OAuth2 config with mock server
func CreateTestOAuth2Config(mockServer *MockOIDCServer) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     mockServer.ClientID,
		ClientSecret: mockServer.ClientSecret,
		RedirectURL:  "http://localhost:5000/openid/callback",
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/auth", mockServer.Server.URL),
			TokenURL: mockServer.TokenURL,
		},
	}
}
