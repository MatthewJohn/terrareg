package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// TerraformIDPValidator interface for token validation (implemented by domain service)
type TerraformIDPValidator interface {
	ValidateToken(ctx context.Context, token string) (interface{}, error)
}

// TerraformUserInfo represents the user information for OIDC (auth package)
type TerraformUserInfo struct {
	Sub      string `json:"sub"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Issuer   string `json:"iss"`
	Audience string `json:"aud"`
}

// convertUserInfo converts domain UserInfo to auth package format
func convertUserInfo(userInfo interface{}) (*TerraformUserInfo, error) {
	// Handle domain service UserInfo conversion
	if user, ok := userInfo.(map[string]interface{}); ok {
		return &TerraformUserInfo{
			Sub:      getString(user, "Sub"),
			Name:     getString(user, "Name"),
			Email:    getString(user, "Email"),
			Issuer:   getString(user, "Issuer"),
			Audience: getString(user, "Audience"),
		}, nil
	}

	// Handle struct conversion with reflection-like approach
	if user, ok := userInfo.(struct{Sub, Name, Email, Issuer, Audience string}); ok {
		return &TerraformUserInfo{
			Sub:      user.Sub,
			Name:     user.Name,
			Email:    user.Email,
			Issuer:   user.Issuer,
			Audience: user.Audience,
		}, nil
	}

	return nil, fmt.Errorf("unsupported UserInfo type")
}

// getString safely extracts string from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}


// TerraformOIDCIDP is a complete implementation of TerraformIDP interface
type TerraformOIDCIDP struct {
	terraformIDPService TerraformIDPValidator
	logger              *zerolog.Logger
	enabled             bool
}

// NewTerraformOIDCIDP creates a Terraform OIDC IDP implementation
func NewTerraformOIDCIDP(terraformIDPService TerraformIDPValidator, logger *zerolog.Logger, enabled bool) *TerraformOIDCIDP {
	return &TerraformOIDCIDP{
		terraformIDPService: terraformIDPService,
		logger:              logger,
		enabled:             enabled,
	}
}

// IsEnabled returns whether the IDP is enabled
func (t *TerraformOIDCIDP) IsEnabled() bool {
	return t.enabled
}

// HandleUserinfoRequest handles a userinfo request
func (t *TerraformOIDCIDP) HandleUserinfoRequest(data []byte, headers map[string]string) (*TerraformUserinfoResponse, error) {
	if !t.enabled {
		return nil, fmt.Errorf("Terraform IDP is not enabled")
	}

	// Extract access token from headers
	authHeader := headers["Authorization"]
	if authHeader == "" {
		return nil, fmt.Errorf("missing Authorization header")
	}

	// Remove "Bearer " prefix
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" || token == authHeader {
		return nil, fmt.Errorf("invalid Authorization header format")
	}

	// Validate access token and get user info
	domainUserInfo, err := t.terraformIDPService.ValidateToken(context.Background(), token)
	if err != nil {
		t.logger.Debug().Err(err).Str("token", token[:min(8, len(token))]).Msg("Failed to validate Terraform token")
		return nil, fmt.Errorf("failed to validate access token: %w", err)
	}

	// Convert domain UserInfo to auth package format
	userInfo, err := convertUserInfo(domainUserInfo)
	if err != nil {
		t.logger.Debug().Err(err).Msg("Failed to convert UserInfo")
		return nil, fmt.Errorf("failed to convert user info: %w", err)
	}

	// Convert to TerraformUserinfoResponse
	response := &TerraformUserinfoResponse{
		Subject:  userInfo.Sub,
		Name:     userInfo.Name,
		Username: userInfo.Email, // Use email as username for Terraform
		Email:    userInfo.Email,
		Groups:   []string{}, // Terraform users don't typically have groups
		Metadata: map[string]interface{}{
			"issuer":   userInfo.Issuer,
			"audience": userInfo.Audience,
		},
	}

	return response, nil
}

// ValidateAccessToken validates an access token
func (t *TerraformOIDCIDP) ValidateAccessToken(token string) (*TerraformTokenValidation, error) {
	if !t.enabled {
		return &TerraformTokenValidation{
			Valid:    false,
			Subject:  "",
			Username: "",
		}, nil
	}

	// Validate the access token using the service
	domainUserInfo, err := t.terraformIDPService.ValidateToken(context.Background(), token)
	if err != nil {
		t.logger.Debug().Err(err).Str("token", token[:min(8, len(token))]).Msg("Token validation failed")
		return &TerraformTokenValidation{
			Valid:    false,
			Subject:  "",
			Username: "",
		}, nil
	}

	// Convert domain UserInfo to auth package format
	userInfo, err := convertUserInfo(domainUserInfo)
	if err != nil {
		t.logger.Debug().Err(err).Msg("Failed to convert UserInfo")
		return &TerraformTokenValidation{
			Valid:    false,
			Subject:  "",
			Username: "",
		}, nil
	}

	// Convert to validation response
	validation := &TerraformTokenValidation{
		Valid:    true,
		Subject:  userInfo.Sub,
		Username: userInfo.Email, // Use email as username
	}

	// Parse expiry from user info if available
	if expiry, ok := parseExpiryFromSubject(userInfo.Sub); ok {
		validation.Expiry = expiry
	}

	return validation, nil
}

// parseExpiryFromSubject extracts expiry from subject if encoded
func parseExpiryFromSubject(subject string) (int64, bool) {
	// In the current implementation, we don't encode expiry in the subject
	// This could be enhanced if needed
	return 0, false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SetEnabled enables or disables the IDP
func (t *TerraformOIDCIDP) SetEnabled(enabled bool) {
	t.enabled = enabled
	t.logger.Info().Bool("enabled", enabled).Msg("Terraform IDP status changed")
}