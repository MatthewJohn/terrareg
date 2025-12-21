package auth

import "context"

// UserInfo represents user information from Terraform IDP validation
type UserInfo struct {
	Sub      string `json:"sub"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Issuer   string `json:"iss"`
	Audience string `json:"aud"`
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