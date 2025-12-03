package oidc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"terrareg/internal/domain/identity/model"
	"terrareg/internal/domain/identity/service"
)

var (
	ErrOIDCNotConfigured    = errors.New("OIDC provider not configured")
	ErrOIDCInvalidToken    = errors.New("invalid OIDC token")
	ErrOIDCAuthentication   = errors.New("OIDC authentication failed")
)

// OIDCProvider implements OpenID Connect authentication
type OIDCProvider struct {
	provider   *oidc.Provider
	oauth2Config *oauth2.Config
	verifier   *oidc.IDTokenVerifier
	config     OIDCConfig
}

// OIDCConfig holds OIDC configuration
type OIDCConfig struct {
	IssuerURL        string
	ClientID          string
	ClientSecret      string
	RedirectURL       string
	Scopes            []string
	ClaimMapping      map[string]string
	SessionTimeout    time.Duration
	SkipNonceCheck    bool
	SkipIssuerCheck   bool
}

// OIDCTokenResponse represents the token response from OIDC
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token"`
}

// NewOIDCProvider creates a new OIDC provider
func NewOIDCProvider(config OIDCConfig) (*OIDCProvider, error) {
	if config.IssuerURL == "" || config.ClientID == "" {
		return nil, ErrOIDCNotConfigured
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Discover OIDC configuration
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to discover OIDC configuration: %w", err)
	}

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       config.Scopes,
		RedirectURL:  config.RedirectURL,
	}

	// Create ID token verifier
	verifierConfig := &oidc.Config{
		ClientID:          config.ClientID,
		SkipClientIDCheck:  false,
		SkipExpiryCheck:   false,
		SkipIssuerCheck:    config.SkipIssuerCheck,
	}

	verifier := provider.Verifier(verifierConfig)

	return &OIDCProvider{
		provider:     provider,
		oauth2Config: oauth2Config,
		verifier:     verifier,
		config:       config,
	}, nil
}

// GetAuthURL returns the OIDC authentication URL
func (p *OIDCProvider) GetAuthURL(ctx context.Context, state string) (string, error) {
	// Generate auth code URL with state
	authURL := p.oauth2Config.AuthCodeURL(state)
	return authURL, nil
}

// Authenticate handles OIDC authentication response
func (p *OIDCProvider) Authenticate(ctx context.Context, request *http.Request) (*service.AuthResult, error) {
	// Extract authorization code from request
	code := request.FormValue("code")
	if code == "" {
		return nil, ErrOIDCAuthentication
	}

	// Exchange authorization code for tokens
	token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Extract and verify ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, ErrOIDCInvalidToken
	}

	// Verify ID token
	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract user information from ID token claims
	userInfo, err := p.extractUserInfo(idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user info from ID token: %w", err)
	}

	// Calculate expires in
	var expiresIn int64
	if token.Expiry != nil {
		expiresIn = int64(time.Until(*token.Expiry).Seconds())
	}

	// Create auth result
	return &service.AuthResult{
		UserID:         userInfo.ID,
		Username:       userInfo.Username,
		Email:          userInfo.Email,
		DisplayName:    userInfo.DisplayName,
		ExternalID:     userInfo.ID,
		AuthProviderID: "oidc",
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		ExpiresIn:      expiresIn,
	}, nil
}

// RefreshToken refreshes an access token using refresh token
func (p *OIDCProvider) RefreshToken(ctx context.Context, refreshToken string) (*service.AuthResult, error) {
	// Configure token source with refresh token
	tokenSource := p.oauth2Config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	// Get new token
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Get user info from token endpoint
	userInfo, err := p.getUserInfoFromToken(ctx, newToken.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from token: %w", err)
	}

	// Calculate expires in
	var expiresIn int64
	if newToken.Expiry != nil {
		expiresIn = int64(time.Until(*newToken.Expiry).Seconds())
	}

	return &service.AuthResult{
		UserID:         userInfo.ID,
		Username:       userInfo.Username,
		Email:          userInfo.Email,
		DisplayName:    userInfo.DisplayName,
		ExternalID:     userInfo.ID,
		AuthProviderID: "oidc",
		AccessToken:    newToken.AccessToken,
		RefreshToken:   newToken.RefreshToken,
		ExpiresIn:      expiresIn,
	}, nil
}

// GetUserInfo returns user information from the OIDC userinfo endpoint
func (p *OIDCProvider) GetUserInfo(ctx context.Context, accessToken string) (*service.UserInfo, error) {
	return p.getUserInfoFromToken(ctx, accessToken)
}

// extractUserInfo extracts user information from OIDC ID token
func (p *OIDCProvider) extractUserInfo(idToken *oidc.IDToken) (*service.UserInfo, error) {
	userInfo := &service.UserInfo{}

	// Extract claims using mapping configuration
	claims := map[string]interface{}{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims from ID token: %w", err)
	}

	// Extract user ID (subject)
	if subject, ok := claims["sub"].(string); ok {
		userInfo.ID = subject
	}

	// Extract email
	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}

	// Extract name/username
	if name, ok := claims["name"].(string); ok {
		userInfo.DisplayName = name
	}
	if username, ok := claims["preferred_username"].(string); ok {
		userInfo.Username = username
	} else if userInfo.DisplayName != "" {
		userInfo.Username = userInfo.DisplayName
	} else if userInfo.ID != "" {
		userInfo.Username = userInfo.ID
	}

	// Apply custom claim mappings
	for claimName, targetField := range p.config.ClaimMapping {
		if value, ok := claims[claimName].(string); ok {
			switch targetField {
			case "email":
				userInfo.Email = value
			case "display_name", "name":
				userInfo.DisplayName = value
			case "username", "preferred_username":
				userInfo.Username = value
			}
		}
	}

	// Validate required fields
	if userInfo.ID == "" {
		return nil, ErrOIDCInvalidToken
	}

	return userInfo, nil
}

// getUserInfoFromToken fetches user info from OIDC userinfo endpoint
func (p *OIDCProvider) getUserInfoFromToken(ctx context.Context, accessToken string) (*service.UserInfo, error) {
	// Create HTTP client
	client := p.provider.Client(ctx)

	// Request userinfo from provider
	resp, err := client.Get(fmt.Sprintf("%suserinfo", p.provider.Endpoint().AuthURL))
	if err != nil {
		return nil, fmt.Errorf("failed to request userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	// Parse userinfo response
	var claims map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	userInfo := &service.UserInfo{}

	// Extract user information
	if sub, ok := claims["sub"].(string); ok {
		userInfo.ID = sub
	}
	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}
	if name, ok := claims["name"].(string); ok {
		userInfo.DisplayName = name
	}
	if username, ok := claims["preferred_username"].(string); ok {
		userInfo.Username = username
	}

	return userInfo, nil
}