package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	oidcAuth "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/auth"
)

// TerraformIdpService manages OAuth flows for Terraform OIDC Identity Provider
type TerraformIdpService struct {
	authCodeRepo          repository.TerraformIdpAuthorizationCodeRepository
	accessTokenRepo       repository.TerraformIdpAccessTokenRepository
	subjectIdentifierRepo repository.TerraformIdpSubjectIdentifierRepository
	keyManager            *oidcAuth.OIDCKeyManager
	tokenSigner           *oidcAuth.TokenSigner
	issuer                string
}

// AuthorizationCodeRequest represents an OAuth authorization code request
type AuthorizationCodeRequest struct {
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
	ResponseType string `json:"response_type"`
}

// AuthorizationCodeResponse represents an OAuth authorization code response
type AuthorizationCodeResponse struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// AccessTokenRequest represents an OAuth access token request
type AccessTokenRequest struct {
	GrantType   string `json:"grant_type"`
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
	ClientID    string `json:"client_id"`
}

// AccessTokenResponse represents an OAuth access token response
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// UserInfo represents the user information returned by the userinfo endpoint
type UserInfo struct {
	Sub      string `json:"sub"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Issuer   string `json:"iss"`
	Audience string `json:"aud"`
}

// NewTerraformIdpService creates a new Terraform IDP service
func NewTerraformIdpService(
	authCodeRepo repository.TerraformIdpAuthorizationCodeRepository,
	accessTokenRepo repository.TerraformIdpAccessTokenRepository,
	subjectIdentifierRepo repository.TerraformIdpSubjectIdentifierRepository,
) *TerraformIdpService {
	// Create key manager for OIDC token signing
	keyManager, err := oidcAuth.NewOIDCKeyManager()
	if err != nil {
		panic(fmt.Sprintf("Failed to create OIDC key manager: %v", err))
	}

	// Create token signer
	tokenSigner := oidcAuth.NewTokenSigner(keyManager, "http://localhost:3000")

	return &TerraformIdpService{
		authCodeRepo:          authCodeRepo,
		accessTokenRepo:       accessTokenRepo,
		subjectIdentifierRepo: subjectIdentifierRepo,
		keyManager:            keyManager,
		tokenSigner:           tokenSigner,
		issuer:                "http://localhost:3000",
	}
}

// CreateAuthorizationCode generates and stores an authorization code for OAuth flow
func (s *TerraformIdpService) CreateAuthorizationCode(ctx context.Context, req AuthorizationCodeRequest) (*AuthorizationCodeResponse, error) {
	if req.ResponseType != "code" {
		return nil, fmt.Errorf("unsupported response_type: %s", req.ResponseType)
	}

	// Generate secure authorization code
	code, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate authorization code: %w", err)
	}

	// Store authorization request data
	authData := map[string]interface{}{
		"client_id":     req.ClientID,
		"redirect_uri":  req.RedirectURI,
		"scope":         req.Scope,
		"state":         req.State,
		"response_type": req.ResponseType,
		"created_at":    time.Now().Unix(),
	}

	dataBytes, err := json.Marshal(authData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authorization data: %w", err)
	}

	// Store with expiration (typically 10 minutes)
	expiry := time.Now().Add(10 * time.Minute)
	err = s.authCodeRepo.Create(ctx, code, dataBytes, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store authorization code: %w", err)
	}

	return &AuthorizationCodeResponse{
		Code:  code,
		State: req.State,
	}, nil
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func (s *TerraformIdpService) ExchangeCodeForToken(ctx context.Context, req AccessTokenRequest) (*AccessTokenResponse, error) {
	if req.GrantType != "authorization_code" {
		return nil, fmt.Errorf("unsupported grant_type: %s", req.GrantType)
	}

	// Validate authorization code
	authCode, err := s.authCodeRepo.FindByKey(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code: %w", err)
	}

	// Parse stored authorization data
	var authData map[string]interface{}
	err = json.Unmarshal(authCode.Data, &authData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse authorization data: %w", err)
	}

	// Validate request matches stored authorization
	storedClientID, ok := authData["client_id"].(string)
	if !ok || storedClientID != req.ClientID {
		return nil, fmt.Errorf("client_id mismatch")
	}

	storedRedirectURI, ok := authData["redirect_uri"].(string)
	if !ok || storedRedirectURI != req.RedirectURI {
		return nil, fmt.Errorf("redirect_uri mismatch")
	}

	// Generate access token
	accessToken, err := generateSecureToken(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Create user info for the token
	userInfo := &UserInfo{
		Sub:      fmt.Sprintf("terraform-user-%s", req.ClientID),
		Name:     "Terraform CLI User",
		Email:    fmt.Sprintf("terraform-%s@example.com", req.ClientID),
		Issuer:   "terraform-idp",
		Audience: req.ClientID,
	}

	// Get scope from auth data, with fallback to default scope
	var scope string
	if s, ok := authData["scope"].(string); ok {
		scope = s
	} else {
		scope = "openid profile"
	}

	tokenData := map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        scope,
		"user_info":    userInfo,
		"created_at":   time.Now().Unix(),
	}

	dataBytes, err := json.Marshal(tokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Store access token with expiration (1 hour)
	expiry := time.Now().Add(time.Hour)
	err = s.accessTokenRepo.Create(ctx, accessToken, dataBytes, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store access token: %w", err)
	}

	// Delete the used authorization code
	err = s.authCodeRepo.DeleteByKey(ctx, req.Code)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to delete used authorization code: %v\n", err)
	}

	return &AccessTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		Scope:       scope,
	}, nil
}

// ValidateToken validates an access token and returns the associated user info
func (s *TerraformIdpService) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	tokenRecord, err := s.accessTokenRepo.FindByKey(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	// Parse token data
	var tokenData map[string]interface{}
	err = json.Unmarshal(tokenRecord.Data, &tokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token data: %w", err)
	}

	// Extract user info
	userInfoData, ok := tokenData["user_info"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("user info not found in token")
	}

	// Convert to UserInfo struct
	userInfoBytes, err := json.Marshal(userInfoData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user info: %w", err)
	}

	var userInfo UserInfo
	err = json.Unmarshal(userInfoBytes, &userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return &userInfo, nil
}

// RevokeToken revokes an access token
func (s *TerraformIdpService) RevokeToken(ctx context.Context, token string) error {
	err := s.accessTokenRepo.DeleteByKey(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

// StoreSubjectIdentifier stores a subject identifier for user tracking
func (s *TerraformIdpService) StoreSubjectIdentifier(ctx context.Context, subject string, clientID string, data map[string]interface{}) error {
	key := fmt.Sprintf("%s:%s", subject, clientID)

	// Add metadata
	data["subject"] = subject
	data["client_id"] = clientID
	data["created_at"] = time.Now().Unix()

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal subject identifier data: %w", err)
	}

	// Store with long expiration (1 year)
	expiry := time.Now().Add(365 * 24 * time.Hour)
	return s.subjectIdentifierRepo.Create(ctx, key, dataBytes, expiry)
}

// GetSubjectIdentifier retrieves stored subject identifier data
func (s *TerraformIdpService) GetSubjectIdentifier(ctx context.Context, subject, clientID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("%s:%s", subject, clientID)

	record, err := s.subjectIdentifierRepo.FindByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("subject identifier not found: %w", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(record.Data, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subject identifier data: %w", err)
	}

	return data, nil
}

// CleanupExpired removes all expired authorization codes, access tokens, and subject identifiers
func (s *TerraformIdpService) CleanupExpired(ctx context.Context) error {
	// Clean up expired authorization codes
	authCodesDeleted, err := s.authCodeRepo.DeleteExpired(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired authorization codes: %w", err)
	}

	// Clean up expired access tokens
	tokensDeleted, err := s.accessTokenRepo.DeleteExpired(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired access tokens: %w", err)
	}

	// Clean up expired subject identifiers
	subjectsDeleted, err := s.subjectIdentifierRepo.DeleteExpired(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired subject identifiers: %w", err)
	}

	fmt.Printf("Cleanup completed: %d authorization codes, %d access tokens, %d subject identifiers removed\n",
		authCodesDeleted, tokensDeleted, subjectsDeleted)

	return nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
