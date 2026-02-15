// Package gpg provides GPG key parsing and signature verification functionality
package gpg

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

// ParseKeyInfo extracts key ID and fingerprint from ASCII armored GPG key
func ParseKeyInfo(asciiArmor string) (keyID, fingerprint string, err error) {
	// Decode ASCII armor
	block, err := armor.Decode(bytes.NewBufferString(asciiArmor))
	if err != nil {
		return "", "", fmt.Errorf("failed to decode ASCII armor: %w", err)
	}

	// Check if it's a public key block
	if block.Type != "PGP PUBLIC KEY BLOCK" {
		return "", "", fmt.Errorf("expected PGP PUBLIC KEY BLOCK, got %s", block.Type)
	}

	// Read the key
	reader := packet.NewReader(block.Body)
	pkt, err := reader.Next()
	if err != nil {
		return "", "", fmt.Errorf("failed to read packet: %w", err)
	}

	// Parse the primary key
	var primaryKey *packet.PublicKey
	switch pk := pkt.(type) {
	case *packet.PublicKey:
		primaryKey = pk
	case *packet.PrivateKey:
		primaryKey = &pk.PublicKey
	default:
		return "", "", fmt.Errorf("expected public key packet, got %T", pkt)
	}

	// Get fingerprint (it's a [20]byte array, not nil-able)
	fingerprintBytes := primaryKey.Fingerprint
	fingerprint = hex.EncodeToString(fingerprintBytes[:])

	// Get key ID (last 16 characters of fingerprint)
	if len(fingerprint) < 16 {
		return "", "", fmt.Errorf("fingerprint too short: %s", fingerprint)
	}
	keyID = fingerprint[len(fingerprint)-16:]

	return keyID, fingerprint, nil
}

// VerifySignature verifies data against a GPG signature using the provided public key
func VerifySignature(asciiArmor, signature, data []byte) (bool, error) {
	// Decode the public key
	keyBlock, err := armor.Decode(bytes.NewReader(asciiArmor))
	if err != nil {
		return false, fmt.Errorf("failed to decode public key armor: %w", err)
	}

	if keyBlock.Type != "PGP PUBLIC KEY BLOCK" {
		return false, fmt.Errorf("expected PGP PUBLIC KEY BLOCK, got %s", keyBlock.Type)
	}

	// Read the key ring
	keyring, err := openpgp.ReadKeyRing(keyBlock.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read key ring: %w", err)
	}

	if len(keyring) == 0 {
		return false, fmt.Errorf("no keys found in key ring")
	}

	// Decode the signature
	sigBlock, err := armor.Decode(bytes.NewReader(signature))
	if err != nil {
		return false, fmt.Errorf("failed to decode signature armor: %w", err)
	}

	if sigBlock.Type != "PGP SIGNATURE" {
		return false, fmt.Errorf("expected PGP SIGNATURE, got %s", sigBlock.Type)
	}

	// Verify the signature
	_, err = openpgp.CheckDetachedSignature(keyring, bytes.NewReader(data), sigBlock.Body)
	if err != nil {
		// Signature verification failed
		return false, nil
	}

	// Signature is valid
	return true, nil
}

// ValidateKeyStructure validates that the GPG key has proper structure
func ValidateKeyStructure(asciiArmor string) error {
	// Decode ASCII armor
	block, err := armor.Decode(bytes.NewBufferString(asciiArmor))
	if err != nil {
		return fmt.Errorf("failed to decode ASCII armor: %w", err)
	}

	// Check block type
	if block.Type != "PGP PUBLIC KEY BLOCK" {
		return fmt.Errorf("invalid GPG key type: %s (expected PGP PUBLIC KEY BLOCK)", block.Type)
	}

	// Parse all entities to validate structure
	entities, err := openpgp.ReadKeyRing(block.Body)
	if err != nil {
		return fmt.Errorf("failed to parse key: %w", err)
	}

	// Check that we have exactly one key
	if len(entities) == 0 {
		return fmt.Errorf("GPG key contains no entities")
	}

	if len(entities) > 1 {
		// Python implementation also checks this
		return fmt.Errorf("GPG key must contain exactly one key, found %d", len(entities))
	}

	return nil
}

// ReadAllKeys reads all keys from ASCII armored data (for validation)
func ReadAllKeys(asciiArmor string) (int, error) {
	// Decode ASCII armor
	block, err := armor.Decode(bytes.NewBufferString(asciiArmor))
	if err != nil {
		return 0, fmt.Errorf("failed to decode ASCII armor: %w", err)
	}

	// Check block type
	if block.Type != "PGP PUBLIC KEY BLOCK" {
		return 0, fmt.Errorf("invalid GPG key type: %s", block.Type)
	}

	// Read all entities
	entities, err := openpgp.ReadKeyRing(block.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read key ring: %w", err)
	}

	return len(entities), nil
}

// VerifySignatureWithFiles verifies a signature from file paths (convenience function)
// This is useful when working with SHA256SUMS and SHA256SUMS.sig files
func VerifySignatureWithFiles(publicKeyASCII, dataContent, signatureContent []byte) (bool, error) {
	return VerifySignature(publicKeyASCII, signatureContent, dataContent)
}

// GetKeyDetails returns detailed information about a GPG key
type KeyDetails struct {
	KeyID        string
	Fingerprint  string
	CreationTime uint64 // Unix timestamp
}

// ParseKeyDetails extracts detailed information from a GPG key
func ParseKeyDetails(asciiArmor string) (*KeyDetails, error) {
	// Decode ASCII armor
	block, err := armor.Decode(bytes.NewBufferString(asciiArmor))
	if err != nil {
		return nil, fmt.Errorf("failed to decode ASCII armor: %w", err)
	}

	// Read the key
	reader := packet.NewReader(block.Body)
	pkt, err := reader.Next()
	if err != nil {
		return nil, fmt.Errorf("failed to read packet: %w", err)
	}

	var primaryKey *packet.PublicKey
	switch pk := pkt.(type) {
	case *packet.PublicKey:
		primaryKey = pk
	case *packet.PrivateKey:
		primaryKey = &pk.PublicKey
	default:
		return nil, fmt.Errorf("expected public key packet, got %T", pkt)
	}

	// Get fingerprint (it's a [20]byte array, not nil-able)
	fingerprintBytes := primaryKey.Fingerprint
	fingerprint := hex.EncodeToString(fingerprintBytes[:])

	// Get key ID (last 16 characters of fingerprint)
	if len(fingerprint) < 16 {
		return nil, fmt.Errorf("fingerprint too short: %s", fingerprint)
	}
	keyID := fingerprint[len(fingerprint)-16:]

	return &KeyDetails{
		KeyID:        keyID,
		Fingerprint:  fingerprint,
		CreationTime: uint64(primaryKey.CreationTime.Unix()),
	}, nil
}

// DecodeSignatureBlock decodes an armored signature block
func DecodeSignatureBlock(signature string) ([]byte, error) {
	block, err := armor.Decode(bytes.NewBufferString(signature))
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature armor: %w", err)
	}

	if block.Type != "PGP SIGNATURE" {
		return nil, fmt.Errorf("expected PGP SIGNATURE, got %s", block.Type)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, block.Body); err != nil {
		return nil, fmt.Errorf("failed to read signature: %w", err)
	}

	return buf.Bytes(), nil
}
