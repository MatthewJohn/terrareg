package terraform

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// TerraformIDPConfig holds the Terraform IDP configuration
type TerraformIDPConfig struct {
	IssuerURL            string
	ClientID             string
	RedirectURIs         []string
	TokenExpiration      time.Duration
	AllowUnsafeRedirects bool
}

// TerraformIDP implements Terrareg as an OIDC Identity Provider for Terraform Cloud/Enterprise
type TerraformIDP struct {
	config    TerraformIDPConfig
	privateKey *rsa.PrivateKey
	keyID      string
}

// NewTerraformIDP creates a new Terraform IDP
func NewTerraformIDP(config TerraformIDPConfig, privateKey *rsa.PrivateKey) (*TerraformIDP, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is required for JWT signing")
	}

	if config.IssuerURL == "" {
		return nil, fmt.Errorf("issuer URL is required for Terraform IDP")
	}

	// Generate a key ID for JWKS
	keyID := uuid.New().String()

	return &TerraformIDP{
		config:     config,
		privateKey: privateKey,
		keyID:      keyID,
	}, nil
}

// IsEnabled checks if the IDP is enabled
func (idp *TerraformIDP) IsEnabled() bool {
	return idp.config.IssuerURL != ""
}

// GetOpenIDConfiguration returns the OpenID configuration
func (idp *TerraformIDP) GetOpenIDConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"issuer":                                idp.config.IssuerURL,
		"subject_types_supported":               []string{"public"},
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code"},
		"jwks_uri":                              fmt.Sprintf("%s/.well-known/jwks.json", idp.config.IssuerURL),
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"userinfo_endpoint":                     fmt.Sprintf("%s/userinfo", idp.config.IssuerURL),
		"token_endpoint":                        fmt.Sprintf("%s/token", idp.config.IssuerURL),
		"authorization_endpoint":                fmt.Sprintf("%s/authorize", idp.config.IssuerURL),
	}
}

// GetJWKS returns the JSON Web Key Set
func (idp *TerraformIDP) GetJWKS() (map[string]interface{}, error) {
	// Convert RSA public key to JWK format
	publicKey := &idp.privateKey.PublicKey

	// Create JWK representation
	jwk := map[string]interface{}{
		"kty": "RSA",
		"kid": idp.keyID,
		"alg": "RS256",
		"use": "sig",
		"n":   jwt.EncodeSegment(publicKey.N.Bytes()),
		"e":   jwt.EncodeSegment([]byte{1, 0, 1}), // Standard exponent for RSA keys
	}

	return map[string]interface{}{
		"keys": []interface{}{jwk},
	}, nil
}

// HandleTokenRequest handles token requests
func (idp *TerraformIDP) HandleTokenRequest(ctx context.Context, tokenRequest map[string]interface{}) (map[string]interface{}, error) {
	// Extract subject (user identifier) from token request
	subject, ok := tokenRequest["subject"].(string)
	if !ok || subject == "" {
		subject = "terraform-user" // Default fallback
	}

	// Token expiration
	expiresIn := int64(3600) // 1 hour default
	if idp.config.TokenExpiration > 0 {
		expiresIn = int64(idp.config.TokenExpiration.Seconds())
	}

	now := time.Now()

	// Create JWT claims with additional user information
	claims := jwt.MapClaims{
		"iss": idp.config.IssuerURL,
		"sub": subject,
		"aud": "terraform",
		"exp": now.Add(time.Duration(expiresIn) * time.Second).Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
	}

	// Add additional user claims from tokenRequest
	for key, value := range tokenRequest {
		if key != "subject" {
			claims[key] = value
		}
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = idp.keyID

	// Sign the token
	signedToken, err := token.SignedString(idp.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT: %w", err)
	}

	return map[string]interface{}{
		"access_token": signedToken,
		"token_type":   "Bearer",
		"expires_in":   expiresIn,
		"scope":        "openid",
	}, nil
}

// HandleUserInfoRequest handles user info requests
func (idp *TerraformIDP) HandleUserInfoRequest(ctx context.Context, token string) (map[string]interface{}, error) {
	// Parse and validate the JWT token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return idp.privateKey.Public(), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	// Extract claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		// Build user info from claims
		userInfo := map[string]interface{}{
			"sub": claims["sub"],
		}

		// Add optional claims if they exist
		if name, ok := claims["name"]; ok {
			userInfo["name"] = name
		} else {
			userInfo["name"] = "Terraform User"
		}

		if email, ok := claims["email"]; ok {
			userInfo["email"] = email
		} else {
			userInfo["email"] = "terraform@example.com"
		}

		// Add any additional claims
		for key, value := range claims {
			if key != "iss" && key != "aud" && key != "exp" && key != "iat" && key != "nbf" {
				if _, exists := userInfo[key]; !exists {
					userInfo[key] = value
				}
			}
		}

		return userInfo, nil
	}

	return nil, fmt.Errorf("invalid JWT token")
}

// GenerateIDToken generates an ID token for the given subject and claims
func (idp *TerraformIDP) GenerateIDToken(ctx context.Context, subject string, additionalClaims map[string]interface{}) (string, error) {
	now := time.Now()

	// Standard OIDC claims
	claims := jwt.MapClaims{
		"iss": idp.config.IssuerURL,
		"sub": subject,
		"aud": "terraform",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
	}

	// Add additional claims
	for key, value := range additionalClaims {
		claims[key] = value
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = idp.keyID

	// Sign the token
	return token.SignedString(idp.privateKey)
}

// ValidateToken validates a JWT token and returns the claims
func (idp *TerraformIDP) ValidateToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return idp.privateKey.Public(), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid JWT token")
}

// GetKeyID returns the key ID for this IDP instance
func (idp *TerraformIDP) GetKeyID() string {
	return idp.keyID
}
