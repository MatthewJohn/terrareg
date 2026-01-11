package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"golang.org/x/oauth2"
)

// OIDCService handles OpenID Connect authentication
type OIDCService struct {
	config       *config.InfrastructureConfig
	provider     *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

// OIDCUserInfo represents user information extracted from OIDC tokens
type OIDCUserInfo struct {
	Subject             string                 `json:"sub"`
	Name                string                 `json:"name,omitempty"`
	Username            string                 `json:"preferred_username,omitempty"`
	Email               string                 `json:"email"`
	EmailVerified       bool                   `json:"email_verified"`
	Groups              []string               `json:"groups,omitempty"`
	Picture             string                 `json:"picture,omitempty"`
	GivenName           string                 `json:"given_name,omitempty"`
	FamilyName          string                 `json:"family_name,omitempty"`
	MiddleName          string                 `json:"middle_name,omitempty"`
	Nickname            string                 `json:"nickname,omitempty"`
	Profile             string                 `json:"profile,omitempty"`
	Website             string                 `json:"website,omitempty"`
	ZoneInfo            string                 `json:"zoneinfo,omitempty"`
	Locale              string                 `json:"locale,omitempty"`
	UpdatedAt           int64                  `json:"updated_at,omitempty"`
	Birthdate           string                 `json:"birthdate,omitempty"`
	Gender              string                 `json:"gender,omitempty"`
	PhoneNumber         string                 `json:"phone_number,omitempty"`
	PhoneNumberVerified bool                   `json:"phone_number_verified"`
	Address             map[string]string      `json:"address,omitempty"`
	RawClaims           map[string]interface{} `json:"-"`
}

// OIDCSession represents an OIDC authentication session
type OIDCSession struct {
	State        string    `json:"state"`
	Nonce        string    `json:"nonce"`
	CodeVerifier string    `json:"code_verifier"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// NewOIDCService creates a new OIDC service
func NewOIDCService(ctx context.Context, config *config.InfrastructureConfig) (*OIDCService, error) {
	if !isOIDCConfigured(config) {
		return nil, fmt.Errorf("OIDC is not configured")
	}

	// Create OIDC provider
	provider, err := oidc.NewProvider(ctx, config.OpenIDConnectIssuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	// Configure OAuth2 client
	oauth2Config := &oauth2.Config{
		ClientID:     config.OpenIDConnectClientID,
		ClientSecret: config.OpenIDConnectClientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       getOIDCScopes(config.OpenIDConnectScopes),
		RedirectURL:  getOIDCRedirectURL(config),
	}

	// Create ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.OpenIDConnectClientID,
	})

	return &OIDCService{
		config:       config,
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
	}, nil
}

// GetAuthURL generates the OIDC authorization URL with proper security parameters
func (s *OIDCService) GetAuthURL(ctx context.Context, state, redirectURL string) (string, *OIDCSession, error) {
	if state == "" {
		return "", nil, fmt.Errorf("state parameter cannot be empty")
	}

	// Generate PKCE parameters for enhanced security
	codeVerifier, codeChallenge, err := generatePKCE()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate PKCE parameters: %w", err)
	}

	// Generate nonce for replay protection
	nonce, err := generateRandomString(32)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create OIDC session
	session := &OIDCSession{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(10 * time.Minute),
	}

	// Build OAuth2 auth code URL options
	var opts []oauth2.AuthCodeOption
	opts = append(opts, oidc.Nonce(nonce))
	opts = append(opts, oauth2.AccessTypeOffline)

	// Add PKCE challenge if supported by the provider
	if codeChallenge != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge", codeChallenge))
		opts = append(opts, oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	}

	// Generate authorization URL
	authURL := s.oauth2Config.AuthCodeURL(state, opts...)

	return authURL, session, nil
}

// ExchangeCode exchanges the authorization code for tokens and extracts user info
func (s *OIDCService) ExchangeCode(ctx context.Context, session *OIDCSession, code, state string) (*OIDCUserInfo, error) {
	// Validate state
	if session.State != state {
		return nil, fmt.Errorf("state parameter mismatch - possible CSRF attack")
	}

	// Validate session expiry
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("OIDC session has expired")
	}

	// Build OAuth2 token exchange options
	var opts []oauth2.AuthCodeOption
	opts = append(opts, oidc.Nonce(session.Nonce))

	// Add PKCE verifier if present
	if session.CodeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", session.CodeVerifier))
	}

	// Exchange authorization code for tokens
	oauth2Token, err := s.oauth2Config.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Extract and verify ID token
	idToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no ID token found in OAuth2 response")
	}

	// Verify the ID token
	verifiedIDToken, err := s.verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract claims from ID token
	var claims map[string]interface{}
	if err := verifiedIDToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims from ID token: %w", err)
	}

	// Convert claims to user info
	userInfo := s.claimsToUserInfo(claims)

	// If access token is available, try to get additional user info from userinfo endpoint
	if oauth2Token.AccessToken != "" {
		if additionalInfo, err := s.getUserInfoFromEndpoint(ctx, oauth2Token.AccessToken); err == nil {
			userInfo = s.mergeUserInfo(userInfo, additionalInfo)
		}
	}

	return userInfo, nil
}

// getUserInfoFromEndpoint fetches additional user information from the OIDC userinfo endpoint
func (s *OIDCService) getUserInfoFromEndpoint(ctx context.Context, accessToken string) (*OIDCUserInfo, error) {
	// Get userinfo endpoint from provider
	userInfoEndpoint := s.provider.UserInfoEndpoint()

	// Create HTTP client and request
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned status %d", resp.StatusCode)
	}

	// Parse response
	var userInfo OIDCUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return &userInfo, nil
}

// claimsToUserInfo converts OIDC claims to user info
func (s *OIDCService) claimsToUserInfo(claims map[string]interface{}) *OIDCUserInfo {
	userInfo := &OIDCUserInfo{
		RawClaims: make(map[string]interface{}),
	}

	// Copy raw claims
	for k, v := range claims {
		userInfo.RawClaims[k] = v
	}

	// Extract standard claims
	if sub, ok := claims["sub"].(string); ok {
		userInfo.Subject = sub
	}

	if name, ok := claims["name"].(string); ok {
		userInfo.Name = name
	}

	if username, ok := claims["preferred_username"].(string); ok {
		userInfo.Username = username
	} else if name, ok := claims["name"].(string); ok {
		// Fallback to name if username not available
		userInfo.Username = name
	}

	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}

	if emailVerified, ok := claims["email_verified"].(bool); ok {
		userInfo.EmailVerified = emailVerified
	}

	// Extract groups if available
	if groups, ok := claims["groups"].([]interface{}); ok {
		userInfo.Groups = make([]string, len(groups))
		for i, group := range groups {
			if str, ok := group.(string); ok {
				userInfo.Groups[i] = str
			}
		}
	}

	// Extract other optional fields
	if picture, ok := claims["picture"].(string); ok {
		userInfo.Picture = picture
	}

	if givenName, ok := claims["given_name"].(string); ok {
		userInfo.GivenName = givenName
	}

	if familyName, ok := claims["family_name"].(string); ok {
		userInfo.FamilyName = familyName
	}

	return userInfo
}

// mergeUserInfo merges additional userinfo into the base user info
func (s *OIDCService) mergeUserInfo(base, additional *OIDCUserInfo) *OIDCUserInfo {
	if base == nil {
		return additional
	}
	if additional == nil {
		return base
	}

	// Create merged result
	merged := *base

	// Override with additional info if base values are empty
	if merged.Name == "" && additional.Name != "" {
		merged.Name = additional.Name
	}

	if merged.Username == "" && additional.Username != "" {
		merged.Username = additional.Username
	}

	if merged.Email == "" && additional.Email != "" {
		merged.Email = additional.Email
		merged.EmailVerified = additional.EmailVerified
	}

	if merged.Picture == "" && additional.Picture != "" {
		merged.Picture = additional.Picture
	}

	if merged.GivenName == "" && additional.GivenName != "" {
		merged.GivenName = additional.GivenName
	}

	if merged.FamilyName == "" && additional.FamilyName != "" {
		merged.FamilyName = additional.FamilyName
	}

	// Merge groups (union)
	if len(additional.Groups) > 0 {
		groupSet := make(map[string]bool)
		for _, group := range merged.Groups {
			groupSet[group] = true
		}
		for _, group := range additional.Groups {
			groupSet[group] = true
		}

		merged.Groups = make([]string, 0, len(groupSet))
		for group := range groupSet {
			merged.Groups = append(merged.Groups, group)
		}
	}

	return &merged
}

// isOIDCConfigured checks if OIDC is properly configured
func isOIDCConfigured(config *config.InfrastructureConfig) bool {
	return config != nil &&
		config.OpenIDConnectIssuer != "" &&
		config.OpenIDConnectClientID != "" &&
		config.OpenIDConnectClientSecret != ""
}

// getOIDCScopes returns the OIDC scopes to use
func getOIDCScopes(configuredScopes []string) []string {
	if len(configuredScopes) > 0 {
		return configuredScopes
	}

	// Default scopes
	return []string{"openid", "profile", "email"}
}

// getOIDCRedirectURL constructs the OIDC redirect URL
func getOIDCRedirectURL(config *config.InfrastructureConfig) string {
	if config.PublicURL != "" {
		redirectURL, _ := url.Parse(config.PublicURL)
		redirectURL.Path = "/openid/callback"
		return redirectURL.String()
	}

	// Fallback URL
	return "http://localhost:5000/openid/callback"
}

// generatePKCE generates PKCE code verifier and challenge for enhanced security
func generatePKCE() (string, string, error) {
	// Generate code verifier (random string)
	verifier, err := generateRandomString(128)
	if err != nil {
		return "", "", err
	}

	// Create SHA256 hash of verifier
	hash := sha256.Sum256([]byte(verifier))

	// Base64 URL encode the hash (without padding)
	challenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	return verifier, challenge, nil
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// VerifyIDToken verifies an ID token signature and returns user info
func (s *OIDCService) VerifyIDToken(ctx context.Context, idToken string) (*auth.UserInfo, error) {
	// Verify the ID token signature and claims
	verifiedToken, err := s.verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("ID token verification failed: %w", err)
	}

	// Extract claims from the verified token
	var claims map[string]interface{}
	if err := verifiedToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims from ID token: %w", err)
	}

	// Create UserInfo from claims
	userInfo := &auth.UserInfo{
		RawClaims: claims,
	}

	// Extract standard claims
	if sub, ok := claims["sub"].(string); ok {
		userInfo.Sub = sub
	}
	if name, ok := claims["name"].(string); ok {
		userInfo.Name = name
	}
	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}
	if emailVerified, ok := claims["email_verified"].(bool); ok {
		userInfo.EmailVerified = emailVerified
	}

	// Extract groups if available
	if groups, ok := claims["groups"].([]interface{}); ok {
		userInfo.Groups = make([]string, len(groups))
		for i, group := range groups {
			if str, ok := group.(string); ok {
				userInfo.Groups[i] = str
			}
		}
	}

	return userInfo, nil
}

// IsConfigured checks if OIDC is properly configured
func (s *OIDCService) IsConfigured() bool {
	return s != nil &&
		s.config != nil &&
		s.config.OpenIDConnectIssuer != "" &&
		s.config.OpenIDConnectClientID != "" &&
		s.config.OpenIDConnectClientSecret != ""
}
