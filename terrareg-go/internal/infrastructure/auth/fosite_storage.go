package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
)

// FositeStorage implements Fosite's storage interface using our existing repositories
type FositeStorage struct {
	authCodeRepo  TerraformIdpAuthorizationCodeRepository
	accessTokenRepo TerraformIdpAccessTokenRepository
	subjectRepo   TerraformIdpSubjectIdentifierRepository
}

// NewFositeStorage creates a new Fosite storage implementation
func NewFositeStorage(
	authCodeRepo TerraformIdpAuthorizationCodeRepository,
	accessTokenRepo TerraformIdpAccessTokenRepository,
	subjectRepo TerraformIdpSubjectIdentifierRepository,
) *FositeStorage {
	return &FositeStorage{
		authCodeRepo:   authCodeRepo,
		accessTokenRepo: accessTokenRepo,
		subjectRepo:    subjectRepo,
	}
}

// Client represents an OAuth2 client in our system
type Client struct {
	ID           string   `json:"client_id"`
	Secret       string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	Audiences    []string `json:"audiences"`
	Public       bool     `json:"public"`
}

// GetClient retrieves an OAuth2 client by ID
func (s *FositeStorage) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	// For Terraform CLI, we typically use predefined clients
	// In a real implementation, you would look up clients from your database

	switch id {
	case "terraform-cli":
		return &Client{
			ID:           "terraform-cli",
			Secret:       "terraform-secret", // In production, use secure secrets
			RedirectURIs: []string{"http://localhost:3000/callback"},
			Scopes:       []string{"openid", "profile", "email", "terraform"},
			Audiences:    []string{"terraform.workspaces"},
			Public:       false,
		}, nil
	default:
		return nil, fosite.ErrNotFound
	}
}

// ClientAssertionJWTValid returns an error if the JWT is invalid
func (s *FositeStorage) ClientAssertionJWTValid(ctx context.Context, jti string, exp time.Time, iss string) error {
	// TODO: Implement JWT validation if needed for client assertions
	return nil
}

// CreateAuthorizeCodeSession creates an authorization code session
func (s *FositeStorage) CreateAuthorizeCodeSession(ctx context.Context, code string, request fosite.Requester) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal authorize request: %w", err)
	}

	expiry := request.GetRequestedAt().Add(10 * time.Minute) // 10 minutes
	return s.authCodeRepo.Create(ctx, code, data, expiry)
}

// GetAuthorizeCodeSession retrieves an authorization code session
func (s *FositeStorage) GetAuthorizeCodeSession(ctx context.Context, code string, session fosite.Session) (fosite.Requester, error) {
	record, err := s.authCodeRepo.FindByKey(ctx, code)
	if err != nil {
		return nil, fosite.ErrNotFound
	}

	var requester fosite.Requester
	err = json.Unmarshal(record.Data, &requester)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorize request: %w", err)
	}

	return requester, nil
}

// DeleteAuthorizeCodeSession deletes an authorization code session
func (s *FositeStorage) DeleteAuthorizeCodeSession(ctx context.Context, code string) error {
	return s.authCodeRepo.DeleteByKey(ctx, code)
}

// CreateAccessTokenSession creates an access token session
func (s *FositeStorage) CreateAccessTokenSession(ctx context.Context, signature string, request fosite.Requester) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal access token request: %w", err)
	}

	expiry := request.GetRequestedAt().Add(time.Hour) // 1 hour
	return s.accessTokenRepo.Create(ctx, signature, data, expiry)
}

// GetAccessTokenSession retrieves an access token session
func (s *FositeStorage) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	record, err := s.accessTokenRepo.FindByKey(ctx, signature)
	if err != nil {
		return nil, fosite.ErrNotFound
	}

	var requester fosite.Requester
	err = json.Unmarshal(record.Data, &requester)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal access token request: %w", err)
	}

	return requester, nil
}

// DeleteAccessTokenSession deletes an access token session
func (s *FositeStorage) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	return s.accessTokenRepo.DeleteByKey(ctx, signature)
}

// CreateRefreshTokenSession creates a refresh token session
func (s *FositeStorage) CreateRefreshTokenSession(ctx context.Context, signature string, request fosite.Requester) error {
	// For Terraform, we typically don't use refresh tokens
	// But we implement the interface for completeness
	return nil
}

// GetRefreshTokenSession retrieves a refresh token session
func (s *FositeStorage) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	// Not implemented for Terraform
	return nil, fosite.ErrNotFound
}

// DeleteRefreshTokenSession deletes a refresh token session
func (s *FositeStorage) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	// Not implemented for Terraform
	return nil
}

// RevokeAccessToken revokes an access token
func (s *FositeStorage) RevokeAccessToken(ctx context.Context, token string) error {
	return s.accessTokenRepo.DeleteByKey(ctx, token)
}

// RevokeRefreshToken revokes a refresh token
func (s *FositeStorage) RevokeRefreshToken(ctx context.Context, token string) error {
	// Not implemented for Terraform
	return nil
}

// GetPublicKey returns the public key for token verification
func (s *FositeStorage) GetPublicKey(ctx context.Context, keyID string) (*json.RawMessage, error) {
	// This would be implemented if you stored keys in the database
	// For now, we rely on the in-memory key management
	return nil, nil
}

// IsJWTUsed checks if a JWT has been used
func (s *FositeStorage) IsJWTUsed(ctx context.Context, jti string) (bool, error) {
	// TODO: Implement JWT usage tracking if needed
	return false, nil
}

// SetJWTUsed marks a JWT as used
func (s *FositeStorage) SetJWTUsed(ctx context.Context, jti string) error {
	// TODO: Implement JWT usage tracking if needed
	return nil
}

// AuthenticateClient authenticates an OAuth2 client
func (s *FositeStorage) AuthenticateClient(ctx context.Context, id string, secret string) (fosite.Client, error) {
	client, err := s.GetClient(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if client secret matches
	if client.GetSecret() != secret {
		return nil, fosite.ErrInvalidClient
	}

	return client, nil
}

// GetSubjectIdentifier gets the subject identifier for a user
func (s *FositeStorage) GetSubjectIdentifier(ctx context.Context, user string, clientID string) (string, error) {
	key := fmt.Sprintf("%s:%s", user, clientID)
	record, err := s.subjectRepo.FindByKey(ctx, key)
	if err != nil {
		return "", err
	}

	var data map[string]interface{}
	err = json.Unmarshal(record.Data, &data)
	if err != nil {
		return "", err
	}

	if sub, ok := data["subject"].(string); ok {
		return sub, nil
	}

	return "", fmt.Errorf("subject not found")
}

// SetSubjectIdentifier sets the subject identifier for a user
func (s *FositeStorage) SetSubjectIdentifier(ctx context.Context, user string, clientID string, sub string) error {
	key := fmt.Sprintf("%s:%s", user, clientID)
	data := map[string]interface{}{
		"subject":  sub,
		"user":     user,
		"clientID": clientID,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Store for 1 year
	expiry := time.Now().Add(365 * 24 * time.Hour)
	return s.subjectRepo.Create(ctx, key, dataBytes, expiry)
}

// GenerateUserCode generates a device code for device flow
func (s *FositeStorage) GenerateUserCode(ctx context.Context, deviceCode, userCode string, request fosite.Requester) error {
	// Not implemented for Terraform
	return fosite.ErrNotImplemented
}

// GetDeviceCode retrieves a device code
func (s *FositeStorage) GetDeviceCode(ctx context.Context, deviceCode string) (fosite.Requester, error) {
	// Not implemented for Terraform
	return nil, fosite.ErrNotFound
}

// AddDeviceUserCode associates a user code with a device code
func (s *FositeStorage) AddDeviceUserCode(ctx context.Context, deviceCode, userCode string) error {
	// Not implemented for Terraform
	return fosite.ErrNotImplemented
}

// GetUserCode retrieves a user code
func (s *FositeStorage) GetUserCode(ctx context.Context, userCode string) (string, error) {
	// Not implemented for Terraform
	return "", fosite.ErrNotFound
}

// AbandonDeviceCode abandons a device code
func (s *FositeStorage) AbandonDeviceCode(ctx context.Context, deviceCode string) error {
	// Not implemented for Terraform
	return fosite.ErrNotImplemented
}

// Additional methods needed for OpenID Connect

// GetOpenIDConnectSession retrieves an OpenID Connect session
func (s *FositeStorage) GetOpenIDConnectSession(ctx context.Context, authorizeCode string, request fosite.Requester) (openid.Session, error) {
	// For Terraform, we create a default OpenID Connect session
	session := &openid.DefaultSession{}
	session.Claims = &openid.IDTokenClaims{
		Issuer:   "http://localhost:3000",
		Subject:  fmt.Sprintf("terraform-user-%s", request.GetClient().GetID()),
		Audience: request.GetClient().GetID(),
	}

	return session, nil
}

// RevokeRefreshTokenMaybeSession revokes a refresh token session
func (s *FositeStorage) RevokeRefreshTokenMaybeSession(ctx context.Context, requestID string) error {
	// Not implemented for Terraform
	return nil
}

// GetPublicKeyByKid returns the public key by key ID
func (s *FositeStorage) GetPublicKeyByKid(ctx context.Context, keyID string) (*json.RawMessage, error) {
	return s.GetPublicKey(ctx, keyID)
}

// GetClientByGrantType returns a client by grant type
func (s *FositeStorage) GetClientByGrantType(ctx context.Context, grantType string, clientID string) (fosite.Client, error) {
	return s.GetClient(ctx, clientID)
}