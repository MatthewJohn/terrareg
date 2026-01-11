package csrf

import (
	"fmt"
)

// TokenGenerator interface for generating CSRF tokens
type TokenGenerator interface {
	GenerateToken() (CSRFToken, error)
}

// SecureTokenGenerator implements secure CSRF token generation
type SecureTokenGenerator struct{}

// NewSecureTokenGenerator creates a new secure token generator
func NewSecureTokenGenerator() *SecureTokenGenerator {
	return &SecureTokenGenerator{}
}

// GenerateToken generates a new secure CSRF token
func (g *SecureTokenGenerator) GenerateToken() (CSRFToken, error) {
	token, err := NewCSRFToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return token, nil
}
