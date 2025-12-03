package terraform

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/service"
)

// TerraformIDP implements Terrareg as an OIDC Identity Provider for Terraform Cloud/Enterprise
type TerraformIDP struct {
	config         TerraformIDPConfig
	jwkSet         *jwk.Set
	privateKey     *rsa.PrivateKey
	signingKey     jwk.Key
}

// TerraformIDPConfig holds the Terraform IDP configuration
type TerraformIDPConfig struct {
	IssuerURL            string
	ClientID             string
	RedirectURIs         []string
	TokenExpiration      time.Duration
	AllowUnsafeRedirects bool
}

// NewTerraformIDP creates a new Terraform IDP
func NewTerraformIDP(config TerraformIDPConfig, privateKey *rsa.PrivateKey) (*TerraformIDP, error) {
	// Create JWK from private key
	signingKey, err := jwk.New(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWK from private key: %w", err)
	}
	signingKey.Set(jwk.KeyIDKey, "terraform-signing-key")
	signingKey.Set(jwk.AlgorithmKey, jwa.RS256.String())

	jwkSet := jwk.NewSet()
	if err := jwkSet.AddKey(signingKey); err != nil {
		return nil, fmt.Errorf("failed to add key to JWK set: %w", err)
	}

	return &TerraformIDP{
		config:     config,
		jwkSet:     jwkSet,
		privateKey: privateKey,
		signingKey: signingKey,
	}, nil
}

// IsEnabled checks if Terraform IDP is enabled
func (idp *TerraformIDP) IsEnabled() bool {
	return idp.config.IssuerURL != ""
}

// GetAuthorizationURL generates authorization URL for Terraform OIDC flow
func (idp *TerraformIDP) GetAuthorizationURL(state string, redirectURI string) (string, error) {
	if !idp.IsEnabled() {
		return "", fmt.Errorf("Terraform IDP is not enabled")
	}

	// Validate redirect URI if unsafe redirects are not allowed
	if !idp.config.AllowUnsafeRedirects {
		validURI := false
		for _, uri := range idp.config.RedirectURIs {
			if uri == redirectURI {
				validURI = true
				break
			}
		}
		if !validURI {
			return "", fmt.Errorf("invalid redirect URI: %s", redirectURI)
		}
	}

	authURL, err := url.Parse(idp.config.IssuerURL + "/.well-known/openid-configuration")
	if err != nil {
		return "", fmt.Errorf("invalid issuer URL: %w", err)
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", idp.config.ClientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "openid email profile")
	params.Set("state", state)

	// For Terraform OIDC, use authorization endpoint
	authURL, err = url.Parse(idp.config.IssuerURL + "/oauth2/auth")
	if err != nil {
		return "", fmt.Errorf("failed to parse auth endpoint: %w", err)
	}

	authURL.RawQuery = params.Encode()
	return authURL.String(), nil
}

// HandleAuthorizationRequest handles OIDC authorization request
func (idp *TerraformIDP) HandleAuthorizationRequest(ctx context.Context, authRequest OIDCAuthRequest) (*OIDCAuthorizationResponse, error) {
	if !idp.IsEnabled() {
		return nil, fmt.Errorf("Terraform IDP is not enabled")
	}

	// Validate request parameters
	if authRequest.ResponseType != "code" {
		return nil, fmt.Errorf("unsupported response type: %s", authRequest.ResponseType)
	}

	if authRequest.ClientID != idp.config.ClientID {
		return nil, fmt.Errorf("invalid client ID: %s", authRequest.ClientID)
	}

	// Generate authorization code
	authCode, err := idp.generateAuthorizationCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate authorization code: %w", err)
	}

	return &OIDCAuthorizationResponse{
		Code:  authCode,
		State: authRequest.State,
	}, nil
}

// HandleTokenRequest handles OIDC token exchange
func (idp *TerraformIDP) HandleTokenRequest(ctx context.Context, tokenRequest OIDCTokenRequest) (*OIDCTokenResponse, error) {
	if !idp.IsEnabled() {
		return nil, fmt.Errorf("Terraform IDP is not enabled")
	}

	if tokenRequest.GrantType == "authorization_code" {
		return idp.handleAuthorizationCodeGrant(ctx, tokenRequest)
	} else if tokenRequest.GrantType == "client_credentials" {
		return idp.handleClientCredentialsGrant(ctx, tokenRequest)
	}

	return nil, fmt.Errorf("unsupported grant type: %s", tokenRequest.GrantType)
}

// HandleUserInfoRequest handles OIDC userinfo request
func (idp *TerraformIDP) HandleUserInfoRequest(ctx context.Context, accessToken string) (*TerraformUserInfo, error) {
	if !idp.IsEnabled() {
		return nil, fmt.Errorf("Terraform IDP is not enabled")
	}

	// Decode and validate access token
	parsedToken, err := jwt.ParseString(accessToken, jwt.WithVerify(jwt.WithKeySet(idp.jwkSet)))
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	// Extract claims
	claims, err := parsedToken.AsMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract token claims: %w", err)
	}

	// For Terraform, we only need the subject claim
	subject, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("missing subject claim in access token")
	}

	return &TerraformUserInfo{
		Subject: subject,
	}, nil
}

// GetOpenIDConfiguration returns OIDC discovery document
func (idp *TerraformIDP) GetOpenIDConfiguration() map[string]interface{} {
	if !idp.IsEnabled() {
		return nil
	}

	return map[string]interface{}{
		"issuer":                 idp.config.IssuerURL,
		"authorization_endpoint": idp.config.IssuerURL + "/oauth2/auth",
		"token_endpoint":         idp.config.IssuerURL + "/oauth2/token",
		"userinfo_endpoint":      idp.config.IssuerURL + "/oauth2/userinfo",
		"jwks_uri":              idp.config.IssuerURL + "/.well-known/jwks.json",
		"response_types_supported": []string{"code"},
		"grant_types_supported": []string{"authorization_code", "client_credentials"},
		"subject_types_supported": []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported": []string{"openid", "email", "profile"},
	}
}

// GetJWKS returns JSON Web Key Set
func (idp *TerraformIDP) GetJWKS() (string, error) {
	if !idp.IsEnabled() {
		return "", fmt.Errorf("Terraform IDP is not enabled")
	}

	jsonBytes, err := json.MarshalIndent(idp.jwkSet, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWK set: %w", err)
	}

	return string(jsonBytes), nil
}

// handleAuthorizationCodeGrant handles authorization code grant flow
func (idp *TerraformIDP) handleAuthorizationCodeGrant(ctx context.Context, tokenRequest OIDCTokenRequest) (*OIDCTokenResponse, error) {
	// In a real implementation, this would validate the authorization code
	// For now, we'll create a simple access token

	// Generate access token with subject from authorization code
	// For Terraform, the subject should be a unique identifier
	subject := "terraform-user" // This should come from the authorization code state

	accessToken, err := idp.generateAccessToken(subject)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &OIDCTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(idp.config.TokenExpiration.Seconds()),
		Scope:       tokenRequest.Scope,
	}, nil
}

// handleClientCredentialsGrant handles client credentials grant flow
func (idp *TerraformIDP) handleClientCredentialsGrant(ctx context.Context, tokenRequest OIDCTokenRequest) (*OIDCTokenResponse, error) {
	// For Terraform client credentials, the subject is the client ID
	subject := tokenRequest.ClientID

	accessToken, err := idp.generateAccessToken(subject)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &OIDCTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(idp.config.TokenExpiration.Seconds()),
		Scope:       tokenRequest.Scope,
	}, nil
}

// generateAuthorizationCode generates a secure authorization code
func (idp *TerraformIDP) generateAuthorizationCode() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// generateAccessToken generates a JWT access token
func (idp *TerraformIDP) generateAccessToken(subject string) (string, error) {
	now := time.Now()

	builder := jwt.NewBuilder().
		Issuer(idp.config.IssuerURL).
		Subject(subject).
		Audience([]string{idp.config.ClientID}).
		IssuedAt(now).
		Expiration(now.Add(idp.config.TokenExpiration)).
		JwtID(idp.generateTokenID())

	token, err := builder.Build()
	if err != nil {
		return "", fmt.Errorf("failed to build JWT: %w", err)
	}

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, idp.privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return string(signed), nil
}

// generateTokenID generates a unique token identifier
func (idp *TerraformIDP) generateTokenID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// OIDCAuthRequest represents an OIDC authorization request
type OIDCAuthRequest struct {
	ResponseType string `json:"response_type"`
	ClientID     string `json:"client_id"`
	RedirectURI  string `json:"redirect_uri"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
}

// OIDCTokenRequest represents an OIDC token request
type OIDCTokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code,omitempty"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	Scope        string `json:"scope"`
}

// OIDCAuthorizationResponse represents an OIDC authorization response
type OIDCAuthorizationResponse struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// OIDCTokenResponse represents an OIDC token response
type OIDCTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
	IDToken     string `json:"id_token,omitempty"`
}

// TerraformUserInfo represents Terraform user information
type TerraformUserInfo struct {
	Subject string `json:"sub"`
}