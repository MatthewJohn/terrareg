package auth

import "context"

// UserInfo represents user information from Terraform IDP validation and OIDC
type UserInfo struct {
	Sub           string                 `json:"sub"`
	Name          string                 `json:"name"`
	Email         string                 `json:"email"`
	EmailVerified bool                   `json:"email_verified"`
	Issuer        string                 `json:"iss"`
	Audience      string                 `json:"aud"`
	Groups        []string               `json:"groups"`
	RawClaims     map[string]interface{} `json:"-"` // Raw claims from ID token
}

// TokenValidator interface for validating tokens
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (interface{}, error)
}

// TerraformIdpValidator interface for Terraform IDP operations
type TerraformIdpValidator interface {
	ValidateToken(ctx context.Context, token string) (*UserInfo, error)
}

// OIDCValidator interface for OpenID Connect operations
type OIDCValidator interface {
	// VerifyIDToken verifies an ID token signature and returns user info
	VerifyIDToken(ctx context.Context, idToken string) (*UserInfo, error)
}