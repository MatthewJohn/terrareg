package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// CookieService handles HTTP cookie operations - pure cookie management
type CookieService struct {
	secretKey     []byte
	sessionCookie string
	isSecure      bool
}

// prepareSecretKey prepares the secret key for encryption
func prepareSecretKey(keyStr string) ([]byte, error) {
	// Try hex decoding first (recommended format)
	if keyBytes, err := hex.DecodeString(keyStr); err == nil {
		if len(keyBytes) < 32 {
			return nil, fmt.Errorf("hex-decoded SECRET_KEY is too short: %d bytes (minimum 32)", len(keyBytes))
		}
		fmt.Printf("CookieService: Using hex SECRET_KEY (%d bytes)\n", len(keyBytes))
		return keyBytes, nil
	}

	// Fall back to raw string if not valid hex
	keyBytes := []byte(keyStr)
	if len(keyBytes) < 32 {
		return nil, fmt.Errorf("SECRET_KEY is too short: %d characters (minimum 32)", len(keyBytes))
	}

	// Warning for raw keys (hex is preferred)
	fmt.Printf("CookieService: Using raw string SECRET_KEY - hex format is recommended (%d bytes)\n", len(keyBytes))
	return keyBytes, nil
}

// NewCookieService creates a new cookie service
func NewCookieService(config *infraConfig.InfrastructureConfig) *CookieService {
	// Default to secure cookies - in production this should be configurable
	isSecure := true

	// Prepare secret key with explicit validation
	secretKey, err := prepareSecretKey(config.SecretKey)
	if err != nil {
		panic(fmt.Sprintf("Invalid SECRET_KEY: %v. Generate with: python -c 'import secrets; print(secrets.token_hex())'", err))
	}

	return &CookieService{
		secretKey:     secretKey,
		sessionCookie: config.SessionCookieName,
		isSecure:      isSecure,
	}
}

// SessionData represents the data stored in the new session cookie
// This type is defined in authentication_service.go to avoid duplication

// EncryptSession encrypts session data for storage in cookie
func (cs *CookieService) EncryptSession(data *SessionData) (string, error) {
	// Serialize session data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to serialize session data: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(cs.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM cipher for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to create nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, jsonData, nil)

	// Return base64 encoded ciphertext
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSession decrypts session data from cookie
func (cs *CookieService) DecryptSession(encrypted string) (*SessionData, error) {
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(cs.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and authenticate
	jsonData, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	// Deserialize session data
	var data SessionData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to deserialize session data: %w", err)
	}

	return &data, nil
}

// SetCookie sets an HTTP cookie with options
func (cs *CookieService) SetCookie(w http.ResponseWriter, name, value string, options *CookieOptions) {
	if options == nil {
		options = &CookieOptions{
			Path:     "/",
			MaxAge:   -1, // Session cookie
			Secure:   cs.isSecure,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     options.Path,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: options.SameSite,
	})
}

// ClearCookie removes a cookie by setting MaxAge to -1
func (cs *CookieService) ClearCookie(w http.ResponseWriter, name string) {
	cs.SetCookie(w, name, "", &CookieOptions{
		Path:     "/",
		MaxAge:   -1,
		Secure:   cs.isSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetSessionCookieName returns the configured session cookie name
func (cs *CookieService) GetSessionCookieName() string {
	return cs.sessionCookie
}

// CookieOptions defines options for setting cookies
type CookieOptions struct {
	Path     string
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

// ValidateSessionCookie validates the session cookie format and expiry
func (cs *CookieService) ValidateSessionCookie(cookieValue string) (*SessionData, error) {
	if cookieValue == "" {
		return nil, ErrNoSessionCookie
	}

	// Decrypt the cookie value
	sessionData, err := cs.DecryptSession(cookieValue)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSessionCookie, err)
	}

	// Check if session has expired
	if sessionData.Expiry != nil && time.Now().After(*sessionData.Expiry) {
		return nil, ErrSessionExpired
	}

	return sessionData, nil
}

// SetSessionCookie sets an encrypted session cookie
func (cs *CookieService) SetSessionCookie(w http.ResponseWriter, sessionData *SessionData) error {
	// Encrypt session data
	encryptedSession, err := cs.EncryptSession(sessionData)
	if err != nil {
		return err
	}

	// Set the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cs.sessionCookie,
		Value:    encryptedSession,
		Path:     "/",
		MaxAge:   0, // Session cookie
		Secure:   cs.isSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// ClearSessionCookie clears the session cookie
func (cs *CookieService) ClearSessionCookie(w http.ResponseWriter) error {
	// Clear the cookie by setting MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:     cs.sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   cs.isSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// Cookie errors
var (
	ErrNoSessionCookie      = fmt.Errorf("no session cookie found")
	ErrInvalidSessionCookie = fmt.Errorf("invalid session cookie")
	ErrSessionExpired       = fmt.Errorf("session cookie expired")
)
