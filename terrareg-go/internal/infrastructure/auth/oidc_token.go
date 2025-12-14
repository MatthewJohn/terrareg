package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// OIDCToken represents an OpenID Connect ID token
type OIDCToken struct {
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	Audience string `json:"aud"`
	Expiry   int64  `json:"exp"`
	IssuedAt int64  `json:"iat"`
	Nonce    string `json:"nonce,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
}

// OIDCClaims represents standard OIDC claims
type OIDCClaims struct {
	Issuer   string `json:"iss"`
	Subject  string `json:"sub"`
	Audience string `json:"aud"`
	Expiry   int64  `json:"exp"`
	IssuedAt int64  `json:"iat"`
	Nonce    string `json:"nonce,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`

	// Terraform-specific claims
	TerraformWorkspaceID string `json:"terraform_workspace_id,omitempty"`
	TerraformRunID       string `json:"terraform_run_id,omitempty"`
}

// TokenSigner handles OIDC token signing and verification
type TokenSigner struct {
	keyManager *OIDCKeyManager
	issuer     string
}

// NewTokenSigner creates a new token signer
func NewTokenSigner(keyManager *OIDCKeyManager, issuer string) *TokenSigner {
	return &TokenSigner{
		keyManager: keyManager,
		issuer:     issuer,
	}
}

// CreateIDToken creates a signed ID token with the specified claims
func (s *TokenSigner) CreateIDToken(claims OIDCClaims) (string, error) {
	// Set standard claims
	claims.Issuer = s.issuer
	claims.IssuedAt = time.Now().Unix()
	claims.Expiry = time.Now().Add(time.Hour).Unix() // 1 hour expiry

	// Get signing key
	privateKey, keyID := s.keyManager.GetSigningKey()
	if privateKey == nil {
		return "", fmt.Errorf("no signing key available")
	}

	// Create JWT header
	header := map[string]interface{}{
		"alg": "RS256",
		"kid": keyID,
		"typ": "JWT",
	}

	// Encode header
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Encode claims
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsBytes)

	// Create signature input
	signingInput := headerEncoded + "." + claimsEncoded

	// Sign the token
	signature, err := s.signWithRSA(privateKey, signingInput)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	// Combine parts
	token := signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)

	return token, nil
}

// VerifyIDToken verifies a signed ID token and returns the claims
func (s *TokenSigner) VerifyIDToken(token string) (*OIDCClaims, error) {
	// Split token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode header (for key ID)
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	var header map[string]interface{}
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	// Get key ID from header
	keyID, ok := header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing key ID in header")
	}

	// Get the public key for verification
	privateKey, exists := s.keyManager.GetSigningKey()
	if !exists {
		return nil, fmt.Errorf("signing key not found")
	}

	// Verify signature
	signingInput := parts[0] + "." + parts[1]
	signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	err = s.verifyWithRSA(&privateKey.PublicKey, signingInput, signatureBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: %w", err)
	}

	// Decode claims
	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	var claims OIDCClaims
	err = json.Unmarshal(claimsBytes, &claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	// Validate standard claims
	if claims.Issuer != s.issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	if time.Now().Unix() > claims.Expiry {
		return nil, fmt.Errorf("token expired")
	}

	if time.Now().Unix() < claims.IssuedAt {
		return nil, fmt.Errorf("token not yet valid")
	}

	return &claims, nil
}

// signWithRSA signs data using RSA-PKCS1v1.5 with SHA-256
func (s *TokenSigner) signWithRSA(privateKey *rsa.PrivateKey, data string) ([]byte, error) {
	hash := sha256.Sum256([]byte(data))
	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
}

// verifyWithRSA verifies RSA signature
func (s *TokenSigner) verifyWithRSA(publicKey *rsa.PublicKey, data string, signature []byte) error {
	hash := sha256.Sum256([]byte(data))
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
}

// CreateTerraformToken creates a token specifically for Terraform workflows
func (s *TokenSigner) CreateTerraformToken(workspaceID, runID, subject, audience string) (string, error) {
	claims := OIDCClaims{
		Subject:              subject,
		Audience:             audience,
		Name:                 "Terraform Service Account",
		Email:                "terraform@terrareg.local",
		TerraformWorkspaceID: workspaceID,
		TerraformRunID:       runID,
	}

	return s.CreateIDToken(claims)
}

// CreateUserInfoToken creates a token with user information
func (s *TokenSigner) CreateUserInfoToken(subject, name, email, audience string) (string, error) {
	claims := OIDCClaims{
		Subject:  subject,
		Audience: audience,
		Name:     name,
		Email:    email,
	}

	return s.CreateIDToken(claims)
}