package service

import (
	"context"
)

// OIDCServiceInterface defines the interface for OIDC authentication
type OIDCServiceInterface interface {
	// GetAuthURL generates the OIDC authorization URL with proper security parameters
	GetAuthURL(ctx context.Context, state, redirectURL string) (string, *OIDCSession, error)

	// ExchangeCode exchanges the authorization code for tokens and extracts user info
	ExchangeCode(ctx context.Context, session *OIDCSession, code, state string) (*OIDCUserInfo, error)

	// IsConfigured checks if OIDC is properly configured
	IsConfigured() bool
}

// SAMLServiceInterface defines the interface for SAML authentication
type SAMLServiceInterface interface {
	// CreateAuthRequest creates a SAML authentication request
	CreateAuthRequest(ctx context.Context, relayState string) (*SAMLAuthRequest, error)

	// ProcessResponse processes a SAML response and extracts user info
	ProcessResponse(ctx context.Context, samlResponse, relayState string) (*SAMLUserInfo, error)

	// GetMetadata generates SAML metadata
	GetMetadata(ctx context.Context) (string, error)

	// IsConfigured checks if SAML is properly configured
	IsConfigured() bool
}

// Ensure concrete types implement the interfaces
var (
	_ OIDCServiceInterface = (*OIDCService)(nil)
	_ SAMLServiceInterface = (*SAMLService)(nil)
)
