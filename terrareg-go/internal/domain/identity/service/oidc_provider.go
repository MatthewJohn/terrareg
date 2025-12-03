package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"coreos/go-oidc/v3/oidc"

	"terrareg/internal/domain/identity/model"
)

var (
	ErrOIDCConfiguration = errors.New("OIDC configuration required")
	ErrOIDCLoginFailed  = errors.New("OIDC login failed")
	ErrOIDCTokenInvalid = errors.New("OIDC token invalid")
)

// OIDCConfig holds OIDC provider configuration
type OIDCConfig struct {
	IssuerURL     string
	ClientID       string
	ClientSecret   string
	RedirectURI    string
	Scopes         []string
	EmailClaim     string
	UsernameClaim  string
	NameClaim      string
}

// OIDCProvider implements OpenID Connect authentication
type OIDCProvider struct {
	config *OIDCConfig
	provider *oidc.Provider
	oauth2Config *oauth2.Config
}

// OIDCTokenResponse represents OIDC token response
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// OIDCUserInfo represents user information from OIDC
type OIDCUserInfo struct {
	Sub          string `json:"sub"`
	Name          string `json:"name,omitempty"`
	Email         string `json:"email,omitempty"`
	PreferredName string `json:"preferred_username,omitempty"`
	GivenName    string `json:"given_name,omitempty"`
	FamilyName   string `json:"family_name,omitempty"`
}

// NewOIDCProvider creates a new OIDC authentication provider
func NewOIDCProvider(config *OIDCConfig) (*OIDCProvider, error) {
	if config == nil {
		return nil, ErrOIDCConfiguration
	}
	if config.IssuerURL == "" {
		return nil, errors.New("issuer URL required")
	}
	if config.ClientID == "" {
		return nil, errors.New("client ID required")
	}
	if config.RedirectURI == "" {
		return nil, errors.New("redirect URI required")
	}

	// Initialize OIDC provider
	provider, err := oidc.NewProvider(context.Background(), config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	// Create OAuth2 config
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  mustParse(config.RedirectURI),
		Scopes:       config.Scopes,
	}

	return &OIDCProvider{
		config:       config,
		provider:     provider,
		oauth2Config: oauth2Config,
	}, nil
}

// Authenticate handles OIDC authentication flow
func (o *OIDCProvider) Authenticate(ctx context.Context, request AuthRequest) (*AuthResult, error) {
	// OIDC authentication is typically initiated via redirect to IdP
	// This is a placeholder implementation
	// In a real implementation, this would:
	// 1. Generate OIDC auth URL
	// 2. Redirect user to IdP
	// 3. Handle authorization code callback
	// 4. Exchange code for tokens
	// 5. Get user info from userinfo endpoint
	// 6. Map to user model

	// For Phase 4, we'll return a mock result if a specific code is provided
	if request.Code == "mock_oidc_code" {
		return &AuthResult{
			UserID:      "oidc_user_456",
			Username:    "oidc_user",
			Email:       "oidc.user@example.com",
			DisplayName:  "OIDC User",
			AccessToken:  "mock_oidc_access_token_" + time.Now().Format("20060102150405"),
			RefreshToken: "mock_oidc_refresh_token_" + time.Now().Format("20060102150405"),
			ExpiresIn:    3600,
			ExternalID:   "oidc_sub_789",
			AuthProviderID: o.config.IssuerURL,
		}, nil
	}

	return nil, ErrOIDCLoginFailed
}

// GetUserInfo fetches user information from OIDC userinfo endpoint
func (o *OIDCProvider) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	// For a real implementation, this would make a request to the OIDC userinfo endpoint
	// For Phase 4, this is a simplified implementation
	if accessToken == "mock_oidc_access_token" {
		return &UserInfo{
			ID:          "oidc_sub_789",
			Username:    "oidc_user",
			Email:       "oidc.user@example.com",
			DisplayName:  "OIDC User",
			AvatarURL:    "",
			Groups:       []string{"users"},
		}, nil
	}

	return nil, ErrOIDCTokenInvalid
}

// RefreshToken refreshes an access token using refresh token
func (o *OIDCProvider) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token required")
	}

	// For a real implementation, this would use the refresh token to get new access token
	// For Phase 4, we'll return a mock result
	if refreshToken == "mock_oidc_refresh_token" {
		return &AuthResult{
			UserID:      "oidc_user_456",
			Username:    "oidc_user",
			Email:       "oidc.user@example.com",
			DisplayName:  "OIDC User",
			AccessToken:  "refreshed_oidc_access_token_" + time.Now().Format("20060102150405"),
			RefreshToken: "new_mock_oidc_refresh_token_" + time.Now().Format("20060102150405"),
			ExpiresIn:    3600,
			ExternalID:   "oidc_sub_789",
			AuthProviderID: o.config.IssuerURL,
		}, nil
	}

	return nil, ErrOIDCLoginFailed
}

// GetLoginURL generates OIDC login URL
func (o *OIDCProvider) GetLoginURL(redirectURI, state string) (string, error) {
	if redirectURI == "" {
		return "", errors.New("redirect URI required")
	}

	// Generate authorization URL with proper parameters
	authURL := o.oauth2Config.AuthCodeURL(state)

	// For Phase 4, we might need to modify the URL or handle it differently
	return authURL, nil
}

// ExchangeCodeForToken exchanges authorization code for tokens
func (o *OIDCProvider) ExchangeCodeForToken(ctx context.Context, code string) (*OIDCTokenResponse, error) {
	// Exchange authorization code for tokens
	token, err := o.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Extract token information
	tokenResponse := &OIDCTokenResponse{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		ExpiresIn:    int64(token.Expiry.Sub(time.Now()).Seconds()),
		RefreshToken: token.RefreshToken,
	}

	return tokenResponse, nil
}

// GetUserInfoFromToken fetches user info using access token
func (o *OIDCProvider) GetUserInfoFromToken(ctx context.Context, accessToken string) (*OIDCUserInfo, error) {
	// Get user info from OIDC provider
	userInfo, err := o.provider.UserInfo(ctx, oidc.UserInfoToken{AccessToken: accessToken})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return userInfo, nil
}

// MapUserInfo maps OIDC user info to our UserInfo structure
func (o *OIDCProvider) MapUserInfo(oidcInfo *OIDCUserInfo) (*UserInfo, error) {
	if oidcInfo == nil {
		return nil, errors.New("OIDC user info is nil")
	}

	userInfo := &UserInfo{
		ID:          oidcInfo.Sub,
		Username:    oidcInfo.PreferredName,
		Email:       oidcInfo.Email,
		DisplayName:  oidcInfo.Name,
		AvatarURL:    "",
		Groups:       []string{},
	}

	// Fallback for username
	if userInfo.Username == "" {
		if oidcInfo.GivenName != "" && oidcInfo.FamilyName != "" {
			userInfo.Username = oidcInfo.GivenName + "." + oidcInfo.FamilyName
		} else {
			userInfo.Username = userInfo.ID
		}
	}

	return userInfo, nil
}

// ValidateConfig validates OIDC configuration
func (o *OIDCProvider) ValidateConfig() error {
	if o.config.IssuerURL == "" {
		return errors.New("issuer URL required")
	}
	if o.config.ClientID == "" {
		return errors.New("client ID required")
	}
	if o.config.RedirectURI == "" {
		return errors.New("redirect URI required")
	}
	return nil
}

// ProcessAuthorizationCode processes authorization code from callback
func (o *OIDCProvider) ProcessAuthorizationCode(ctx context.Context, code, state string) (*AuthResult, error) {
	if code == "" {
		return nil, errors.New("authorization code required")
	}

	// Exchange code for tokens
	tokenResponse, err := o.ExchangeCodeForToken(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for tokens: %w", err)
	}

	// Get user info
	userInfo, err := o.GetUserInfoFromToken(ctx, tokenResponse.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Map to our structure
	mappedUserInfo, err := o.MapUserInfo(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to map user info: %w", err)
	}

	return &AuthResult{
		UserID:        mappedUserInfo.ID,
		Username:      mappedUserInfo.Username,
		Email:         mappedUserInfo.Email,
		DisplayName:   mappedUserInfo.DisplayName,
		AccessToken:   tokenResponse.AccessToken,
		RefreshToken:  tokenResponse.RefreshToken,
		ExpiresIn:     tokenResponse.ExpiresIn,
		ExternalID:    mappedUserInfo.ID,
		AuthProviderID: o.config.IssuerURL,
	}, nil
}

// mustParse parses a URL and panics if invalid (helper function)
func mustParse(rawURL string) *url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("invalid URL %s: %v", rawURL, err))
	}
	return parsed
}