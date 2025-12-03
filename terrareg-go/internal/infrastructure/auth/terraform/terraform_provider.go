package terraform

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"terrareg/internal/domain/identity/model"
	"terrareg/internal/domain/identity/service"
)

var (
	ErrTerraformNotConfigured    = errors.New("Terraform provider not configured")
	ErrTerraformInvalidToken    = errors.New("invalid Terraform token")
	ErrTerraformAuthentication = errors.New("Terraform authentication failed")
)

// TerraformProvider implements Terraform OAuth authentication
type TerraformProvider struct {
	config TerraformConfig
}

// TerraformConfig holds Terraform configuration
type TerraformConfig struct {
	ClientID         string
	ClientSecret     string
	RedirectURL      string
	AuthURL          string
	TokenURL         string
	UserInfoURL      string
	Scopes           []string
	SessionTimeout    time.Duration
	StateStore       map[string]TerraformState
	RequireHTTPS      bool
}

// TerraformState represents OAuth state for Terraform
type TerraformState struct {
	State        string
	ClientID     string
	RedirectURI  string
	Scopes       []string
	ExpiresAt    time.Time
}

// TerraformUser represents Terraform user information
type TerraformUser struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	Organization string `json:"organization"`
	CreatedAt    string `json:"created_at"`
}

// TerraformTokenResponse represents the token response from Terraform
type TerraformTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserID      string `json:"user_id,omitempty"`
}

// NewTerraformProvider creates a new Terraform provider
func NewTerraformProvider(config TerraformConfig) (*TerraformProvider, error) {
	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, ErrTerraformNotConfigured
	}

	// Default Terraform OAuth endpoints (placeholder - would be actual Terraform Cloud endpoints)
	if config.AuthURL == "" {
		config.AuthURL = "https://app.terraform.io/oauth2/authorize"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://app.terraform.io/oauth2/token"
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = "https://app.terraform.io/api/v2/user"
	}

	return &TerraformProvider{
		config: TerraformConfig{
			ClientID:        config.ClientID,
			ClientSecret:    config.ClientSecret,
			RedirectURL:     config.RedirectURL,
			AuthURL:         config.AuthURL,
			TokenURL:        config.TokenURL,
			UserInfoURL:     config.UserInfoURL,
			Scopes:          config.Scopes,
			SessionTimeout:   config.SessionTimeout,
			StateStore:      make(map[string]TerraformState),
			RequireHTTPS:     config.RequireHTTPS,
		},
	}, nil
}

// GetAuthURL returns the Terraform authentication URL
func (p *TerraformProvider) GetAuthURL(ctx context.Context, state string) (string, error) {
	// Generate secure state if not provided
	if state == "" {
		stateBytes := make([]byte, 32)
		rand.Read(stateBytes)
		state = base64.URLEncoding.EncodeToString(stateBytes)
	}

	// Store state for verification
	p.config.StateStore[state] = TerraformState{
		State:       state,
		ClientID:    p.config.ClientID,
		RedirectURI: p.config.RedirectURL,
		Scopes:      p.config.Scopes,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}

	// Build authorization URL
	authURL, err := url.Parse(p.config.AuthURL)
	if err != nil {
		return "", fmt.Errorf("invalid Terraform auth URL: %w", err)
	}

	params := url.Values{}
	params.Set("client_id", p.config.ClientID)
	params.Set("redirect_uri", p.config.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(p.config.Scopes, " "))
	params.Set("state", state)

	authURL.RawQuery = params.Encode()
	return authURL.String(), nil
}

// Authenticate handles Terraform authentication response
func (p *TerraformProvider) Authenticate(ctx context.Context, request *http.Request) (*service.AuthResult, error) {
	// Extract authorization code and state from request
	code := request.FormValue("code")
	state := request.FormValue("state")

	if code == "" {
		return nil, ErrTerraformAuthentication
	}

	// Verify state
	storedState, exists := p.config.StateStore[state]
	if !exists {
		return nil, ErrTerraformAuthentication
	}

	if time.Now().After(storedState.ExpiresAt) {
		delete(p.config.StateStore, state)
		return nil, ErrTerraformAuthentication
	}

	// Clean up state
	delete(p.config.StateStore, state)

	// Exchange authorization code for tokens
	token, err := p.exchangeCodeForToken(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Get user information from Terraform API
	userInfo, err := p.getUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get Terraform user info: %w", err)
	}

	// Create auth result
	return &service.AuthResult{
		UserID:         userInfo.ID,
		Username:       userInfo.Username,
		Email:          userInfo.Email,
		DisplayName:    userInfo.DisplayName,
		ExternalID:     userInfo.ID,
		AuthProviderID: "terraform",
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		ExpiresIn:      int64(token.ExpiresIn),
	}, nil
}

// RefreshToken refreshes an access token using refresh token
func (p *TerraformProvider) RefreshToken(ctx context.Context, refreshToken string) (*service.AuthResult, error) {
	if refreshToken == "" {
		return nil, ErrTerraformInvalidToken
	}

	// Exchange refresh token for new access token
	token, err := p.exchangeRefreshTokenForToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Get user information (Terraform tokens include user ID)
	userInfo := &service.UserInfo{
		ID:          token.UserID,
		Username:    "", // Would need to fetch from API
		Email:       "", // Would need to fetch from API
		DisplayName: "", // Would need to fetch from API
	}

	return &service.AuthResult{
		UserID:         token.UserID,
		Username:       userInfo.Username,
		Email:          userInfo.Email,
		DisplayName:    userInfo.DisplayName,
		ExternalID:     token.UserID,
		AuthProviderID: "terraform",
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		ExpiresIn:      int64(token.ExpiresIn),
	}, nil
}

// GetUserInfo returns user information from Terraform API
func (p *TerraformProvider) GetUserInfo(ctx context.Context, accessToken string) (*service.UserInfo, error) {
	userInfo, err := p.getUserInfo(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	return &service.UserInfo{
		ID:          userInfo.ID,
		Username:    userInfo.Username,
		Email:       userInfo.Email,
		DisplayName: userInfo.DisplayName,
		Groups:      []string{userInfo.Organization}, // Use organization as primary group
	}, nil
}

// exchangeCodeForToken exchanges authorization code for access token
func (p *TerraformProvider) exchangeCodeForToken(ctx context.Context, code string) (*TerraformTokenResponse, error) {
	// Build token request data
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("redirect_uri", p.config.RedirectURL)

	// Make POST request to token endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	var tokenResp TerraformTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// exchangeRefreshTokenForToken exchanges refresh token for new access token
func (p *TerraformProvider) exchangeRefreshTokenForToken(ctx context.Context, refreshToken string) (*TerraformTokenResponse, error) {
	// Build refresh request data
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)

	// Make POST request to token endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make refresh token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh token request failed with status: %d", resp.StatusCode)
	}

	var tokenResp TerraformTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh token response: %w", err)
	}

	return &tokenResp, nil
}

// getUserInfo fetches user information from Terraform API
func (p *TerraformProvider) getUserInfo(ctx context.Context, accessToken string) (*TerraformUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make user info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", resp.StatusCode)
	}

	var user TerraformUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user info response: %w", err)
	}

	return &user, nil
}

// CleanupExpiredStates removes expired OAuth states
func (p *TerraformProvider) CleanupExpiredStates() {
	now := time.Now()
	for state, terraformState := range p.config.StateStore {
		if now.After(terraformState.ExpiresAt) {
			delete(p.config.StateStore, state)
		}
	}
}