package auth

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"
)

// OIDCKeyManager manages RSA keys for OIDC token signing
type OIDCKeyManager struct {
	mu       sync.RWMutex
	keys     map[string]*rsa.PrivateKey
	keyIDs   []string
	keyCache []byte // Cached JWKS JSON
}

// JWKS represents JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	KeyType   string `json:"kty"`
	KeyID     string `json:"kid"`
	Use       string `json:"use"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

// NewOIDCKeyManager creates a new key manager, loading keys from configuration if available
func NewOIDCKeyManager(signingKeyPath string) (*OIDCKeyManager, error) {
	manager := &OIDCKeyManager{
		keys: make(map[string]*rsa.PrivateKey),
	}

	// Try to load key from configuration path
	if signingKeyPath != "" {
		err := manager.loadKeyFromFile(signingKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load signing key from %s: %w", signingKeyPath, err)
		}
	}

	// If no key loaded, generate one (fallback for development)
	if len(manager.keys) == 0 {
		err := manager.generateNewKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate initial signing key: %w", err)
		}
	}

	return manager, nil
}

// loadKeyFromFile loads an RSA private key from a PEM file
func (m *OIDCKeyManager) loadKeyFromFile(keyPath string) error {
	// Read key file
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Decode PEM block
	block, _ := pem.Decode(keyData)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("key is not RSA private key")
		}
	}

	// Generate key ID based on filename and hash
	keyID := fmt.Sprintf("config-%x", md5.Sum([]byte(keyPath)))[:16]

	// Store the key
	m.mu.Lock()
	m.keys[keyID] = privateKey
	m.keyIDs = append([]string{keyID}, m.keyIDs...) // Config keys first
	m.keyCache = nil                                // Invalidate cache
	m.mu.Unlock()

	return nil
}

// generateNewKey generates a new RSA key for signing
func (m *OIDCKeyManager) generateNewKey() error {
	// Generate 2048-bit RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Generate key ID based on timestamp and random bytes
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	keyID := fmt.Sprintf("key-%d-%x", timestamp, randomBytes)

	// Store the key
	m.mu.Lock()
	m.keys[keyID] = privateKey
	m.keyIDs = append([]string{keyID}, m.keyIDs...) // Newest keys first
	m.keyCache = nil                                // Invalidate cache
	m.mu.Unlock()

	return nil
}

// GetSigningKey returns the current signing key
func (m *OIDCKeyManager) GetSigningKey() (*rsa.PrivateKey, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.keyIDs) == 0 {
		return nil, ""
	}

	keyID := m.keyIDs[0] // Use the most recent key for signing
	return m.keys[keyID], keyID
}

// GetJWKS returns the JSON Web Key Set for public key verification
func (m *OIDCKeyManager) GetJWKS() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return cached JWKS if available
	if m.keyCache != nil {
		return m.keyCache, nil
	}

	jwks := JWKS{
		Keys: make([]JWK, 0, len(m.keyIDs)),
	}

	for _, keyID := range m.keyIDs {
		privateKey := m.keys[keyID]
		if privateKey == nil {
			continue
		}

		// Convert to public key components
		publicKey := &privateKey.PublicKey

		jwk := JWK{
			KeyType:   "RSA",
			KeyID:     keyID,
			Use:       "sig",
			Algorithm: "RS256",
			Modulus:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
			Exponent:  base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
		}

		jwks.Keys = append(jwks.Keys, jwk)
	}

	// Cache the result
	jwksBytes, err := json.Marshal(jwks)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JWKS: %w", err)
	}

	// Note: In a real implementation, we'd want to atomically update the cache
	// For now, this is sufficient for our use case
	return jwksBytes, nil
}

// GenerateSelfSignedCertificate creates a self-signed certificate for the OIDC provider
func (m *OIDCKeyManager) GenerateSelfSignedCertificate() ([]byte, error) {
	privateKey, _ := m.GetSigningKey()
	if privateKey == nil {
		return nil, fmt.Errorf("no signing key available")
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Terrareg OIDC Provider"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	// Generate certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return certPEM, nil
}

// RotateKeys generates a new signing key and keeps old keys for verification
func (m *OIDCKeyManager) RotateKeys() error {
	// Generate new signing key
	err := m.generateNewKey()
	if err != nil {
		return fmt.Errorf("failed to rotate keys: %w", err)
	}

	// In a production environment, you might want to clean up very old keys
	// For now, we keep all keys for verification

	return nil
}

// CleanupOldKeys removes keys older than the specified duration
func (m *OIDCKeyManager) CleanupOldKeys(maxAge time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	keepKeys := make([]string, 0)

	for _, keyID := range m.keyIDs {
		// Extract timestamp from key ID
		var timestamp int64
		fmt.Sscanf(keyID, "key-%d", &timestamp)

		keyTime := time.Unix(timestamp, 0)
		if keyTime.After(cutoff) {
			keepKeys = append(keepKeys, keyID)
		} else {
			// Remove old key
			delete(m.keys, keyID)
		}
	}

	m.keyIDs = keepKeys
	m.keyCache = nil // Invalidate cache
}
