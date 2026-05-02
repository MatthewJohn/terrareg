package service

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSessionCipher tests the constructor
func TestNewSessionCipher(t *testing.T) {
	tests := []struct {
		name      string
		secretKey string
		wantErr   string
	}{
		{
			name:      "valid secret key creates cipher",
			secretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			wantErr:   "",
		},
		{
			name:      "short secret key creates cipher (gets hashed)",
			secretKey: "short",
			wantErr:   "",
		},
		{
			name:      "empty secret key returns error",
			secretKey: "",
			wantErr:   "secret key cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cipher, err := NewSessionCipher(tt.secretKey)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Nil(t, cipher)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cipher)
			}
		})
	}
}

// TestSessionCipher_EncryptDecryptRoundTrip tests successful encryption and decryption
func TestSessionCipher_EncryptDecryptRoundTrip(t *testing.T) {
	secretKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	cipher, err := NewSessionCipher(secretKey)
	require.NoError(t, err)

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "simple string",
			data: []byte("hello world"),
		},
		{
			name: "session data JSON",
			data: []byte(`{"user_id": "123", "username": "testuser", "expires": "2024-01-01T00:00:00Z"}`),
		},
		{
			name: "binary data",
			data: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name: "large data (1KB)",
			data: make([]byte, 1024),
		},
		{
			name: "unicode data",
			data: []byte("Hello 世界 🌍🚀"),
		},
		{
			name: "special characters",
			data: []byte("!@#$%^&*()_+-=[]{}|;':\",./<>?\n\t\r"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill large data with a pattern
			if tt.name == "large data (1KB)" {
				for i := range tt.data {
					tt.data[i] = byte(i % 256)
				}
			}

			// Encrypt
			encrypted, err := cipher.Encrypt(tt.data)
			require.NoError(t, err)
			assert.NotEmpty(t, encrypted)
			assert.NotEqual(t, string(tt.data), encrypted) // Encrypted should differ from plaintext

			// Decrypt
			decrypted, err := cipher.Decrypt(encrypted)
			require.NoError(t, err)
			// Handle empty data case: gcm.Open returns nil for empty plaintext
			if len(tt.data) == 0 {
				assert.Nil(t, decrypted, "Empty data should decrypt to nil")
			} else {
				assert.Equal(t, tt.data, decrypted, "Decrypted data should match original")
			}
		})
	}
}

// TestSessionCipher_Decrypt_InvalidInputs tests decrypt with various invalid inputs
func TestSessionCipher_Decrypt_InvalidInputs(t *testing.T) {
	secretKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	cipher, err := NewSessionCipher(secretKey)
	require.NoError(t, err)

	// First, encrypt some valid data
	validData := []byte("test session data")
	validEncrypted, err := cipher.Encrypt(validData)
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       string
		wantErr     error
		errContains string
	}{
		{
			name:        "empty string",
			input:       "",
			wantErr:     ErrInvalidCookie,
			errContains: "ciphertext too short",
		},
		{
			name:        "invalid base64",
			input:       "not-valid-base64!!!",
			wantErr:     ErrInvalidCookie,
			errContains: "failed to decode base64",
		},
		{
			name:        "valid base64 but too short (less than nonce size)",
			input:       base64.StdEncoding.EncodeToString([]byte{1, 2, 3}),
			wantErr:     ErrInvalidCookie,
			errContains: "ciphertext too short",
		},
		{
			name:        "tampered data (modified ciphertext)",
			input:       tamperWithCiphertext(validEncrypted),
			wantErr:     ErrDecryptFailed,
			errContains: "",
		},
		{
			name:        "different secret key",
			input:       validEncrypted,
			wantErr:     ErrDecryptFailed,
			errContains: "", // Error message varies, just check for decrypt failed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var differentKeyCipher *SessionCipher
			if tt.name == "different secret key" {
				differentKeyCipher, _ = NewSessionCipher("different-secret-key-0123456789abcdef")
			} else {
				differentKeyCipher = cipher
			}

			_, err := differentKeyCipher.Decrypt(tt.input)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

// TestSessionCipher_Encrypt_DifferentResults tests that encryption produces different results each time
func TestSessionCipher_Encrypt_DifferentResults(t *testing.T) {
	secretKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	cipher, err := NewSessionCipher(secretKey)
	require.NoError(t, err)

	data := []byte("session data")

	encrypted1, err := cipher.Encrypt(data)
	require.NoError(t, err)

	encrypted2, err := cipher.Encrypt(data)
	require.NoError(t, err)

	// Encrypted data should be different each time due to random nonce
	assert.NotEqual(t, encrypted1, encrypted2, "Encryption should produce different results due to random nonce")

	// But both should decrypt to the same original data
	decrypted1, err := cipher.Decrypt(encrypted1)
	require.NoError(t, err)

	decrypted2, err := cipher.Decrypt(encrypted2)
	require.NoError(t, err)

	assert.Equal(t, data, decrypted1)
	assert.Equal(t, data, decrypted2)
}

// TestSessionCipher_Encrypt_OutputFormat tests that encrypted output is valid base64
func TestSessionCipher_Encrypt_OutputFormat(t *testing.T) {
	secretKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	cipher, err := NewSessionCipher(secretKey)
	require.NoError(t, err)

	data := []byte("test")

	encrypted, err := cipher.Encrypt(data)
	require.NoError(t, err)

	// Verify output is valid base64
	_, err = base64.StdEncoding.DecodeString(encrypted)
	assert.NoError(t, err, "Encrypted output should be valid base64")

	// Verify output is URL-safe (no special characters that would break cookies)
	assert.NotContains(t, encrypted, "\n")
	assert.NotContains(t, encrypted, "\r")
	assert.NotContains(t, encrypted, " ")
}

// TestSessionCipher_KeyDerivation tests that different secret keys produce different derived keys
func TestSessionCipher_KeyDerivation(t *testing.T) {
	key1 := "secret-key-1"
	key2 := "secret-key-2"

	cipher1, err := NewSessionCipher(key1)
	require.NoError(t, err)

	cipher2, err := NewSessionCipher(key2)
	require.NoError(t, err)

	data := []byte("test data")

	encrypted1, err := cipher1.Encrypt(data)
	require.NoError(t, err)

	encrypted2, err := cipher2.Encrypt(data)
	require.NoError(t, err)

	// Encrypted data should be different with different keys
	assert.NotEqual(t, encrypted1, encrypted2)

	// Cipher1 cannot decrypt cipher2's data and vice versa
	_, err = cipher2.Decrypt(encrypted1)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDecryptFailed)
}

// TestSessionCipher_EdgeCases tests various edge cases
func TestSessionCipher_EdgeCases(t *testing.T) {
	secretKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	cipher, err := NewSessionCipher(secretKey)
	require.NoError(t, err)

	t.Run("very large data (100KB)", func(t *testing.T) {
		data := make([]byte, 100*1024)
		for i := range data {
			data[i] = byte(i % 256)
		}

		encrypted, err := cipher.Encrypt(data)
		require.NoError(t, err)

		decrypted, err := cipher.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, data, decrypted)
	})

	t.Run("data with null bytes", func(t *testing.T) {
		data := []byte{0x00, 0x01, 0x00, 0x02, 0x00, 0x00}

		encrypted, err := cipher.Encrypt(data)
		require.NoError(t, err)

		decrypted, err := cipher.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, data, decrypted)
	})

	t.Run("all possible byte values", func(t *testing.T) {
		data := make([]byte, 256)
		for i := range data {
			data[i] = byte(i)
		}

		encrypted, err := cipher.Encrypt(data)
		require.NoError(t, err)

		decrypted, err := cipher.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, data, decrypted)
	})
}

// TestSessionCipher_ConcurrentAccess tests thread safety of encrypt/decrypt operations
func TestSessionCipher_ConcurrentAccess(t *testing.T) {
	secretKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	cipher, err := NewSessionCipher(secretKey)
	require.NoError(t, err)

	const numGoroutines = 50
	const iterations = 10

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			data := []byte(strings.Repeat("test", id))
			for j := 0; j < iterations; j++ {
				encrypted, err := cipher.Encrypt(data)
				if err != nil {
					t.Errorf("Encrypt failed: %v", err)
					return
				}

				decrypted, err := cipher.Decrypt(encrypted)
				if err != nil {
					t.Errorf("Decrypt failed: %v", err)
					return
				}

				if string(decrypted) != string(data) {
					t.Errorf("Round trip failed: got %q, want %q", string(decrypted), string(data))
					return
				}
			}
		}(i)
	}
}

// tamperWithCiphertext modifies the encrypted data to simulate tampering
func tamperWithCiphertext(encoded string) string {
	// Decode base64
	ciphertext, _ := base64.StdEncoding.DecodeString(encoded)

	// Modify a byte in the middle (avoiding the nonce at the beginning)
	if len(ciphertext) > 20 {
		ciphertext[len(ciphertext)/2] ^= 0xFF // Flip all bits
	}

	// Re-encode
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// TestSessionCipher_Errors tests the exported error variables
func TestSessionCipher_Errors(t *testing.T) {
	assert.NotNil(t, ErrInvalidCookie)
	assert.NotNil(t, ErrDecryptFailed)

	assert.Equal(t, "invalid session cookie", ErrInvalidCookie.Error())
	assert.Equal(t, "failed to decrypt session data", ErrDecryptFailed.Error())
}
