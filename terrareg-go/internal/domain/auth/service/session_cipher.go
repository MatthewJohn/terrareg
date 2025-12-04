package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidCookie = errors.New("invalid session cookie")
	ErrDecryptFailed = errors.New("failed to decrypt session data")
)

// SessionCipher handles encryption and decryption of session cookie data
type SessionCipher struct {
	key []byte
}

// NewSessionCipher creates a new session cipher using the provided secret key
func NewSessionCipher(secretKey string) (*SessionCipher, error) {
	if secretKey == "" {
		return nil, errors.New("secret key cannot be empty")
	}

	// Derive encryption key using SHA256 of the secret key
	hash := sha256.Sum256([]byte(secretKey))

	return &SessionCipher{
		key: hash[:],
	}, nil
}

// Encrypt encrypts session data into a cookie-safe format
// Returns base64-encoded encrypted data with IV
func (c *SessionCipher) Encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce (12 bytes is recommended for GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Base64 encode the result for cookie safety
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts session data from a cookie-safe format
// Expects base64-encoded encrypted data with IV
func (c *SessionCipher) Decrypt(encodedData string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode base64", ErrInvalidCookie)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("%w: ciphertext too short", ErrInvalidCookie)
	}

	// Split nonce and actual ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and authenticate
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}

	return plaintext, nil
}