package csrf

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// CSRFToken represents a CSRF token value
type CSRFToken string

// NewCSRFToken generates a new CSRF token using secure random bytes
func NewCSRFToken() (CSRFToken, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Hash the random bytes using SHA256
	hash := sha256.Sum256(bytes)
	return CSRFToken(hex.EncodeToString(hash[:])), nil
}

// String returns the string representation of the CSRF token
func (t CSRFToken) String() string {
	return string(t)
}

// IsEmpty checks if the token is empty
func (t CSRFToken) IsEmpty() bool {
	return len(t) == 0
}

// Equals compares two CSRF tokens for equality
func (t CSRFToken) Equals(other CSRFToken) bool {
	return t == other
}
