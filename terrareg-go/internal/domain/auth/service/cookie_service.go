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

	terraregAppConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/config"
)

// CookieService handles HTTP cookie operations - pure cookie management
type CookieService struct {
	secretKey     []byte
	sessionCookie string
	isSecure      bool
}

// NewCookieService creates a new cookie service
func NewCookieService(config *terraregAppConfig.Config) *CookieService {
	// Default to secure cookies - in production this should be configurable
	isSecure := true

	var secretKey []byte
	keyStr := config.SecretKey

	// Check if key is hex-encoded (common for SECRET_KEY)
	if len(keyStr) == 64 || len(keyStr) == 96 {
		// Check if it's all hex characters
		_, err := hex.DecodeString(keyStr)
		if err == nil {
			// It's valid hex, decode it
			secretKey, _ = hex.DecodeString(keyStr)
			fmt.Printf("CookieService: Detected hex SECRET_KEY, decoded to %d bytes\n", len(secretKey))
		} else {
			// Not valid hex, use as raw bytes
			secretKey = []byte(keyStr)
			fmt.Printf("CookieService: Not hex SECRET_KEY, using raw %d bytes\n", len(secretKey))
		}
	} else {
		// Use as raw bytes
		secretKey = []byte(keyStr)
		fmt.Printf("CookieService: Using raw SECRET_KEY with %d bytes\n", len(secretKey))
	}

	// Ensure the secret key is at least 32 bytes for AES-256
	if len(secretKey) < 32 {
		fmt.Printf("CookieService: SECRET_KEY too short (%d bytes), padding to 32\n", len(secretKey))
		// Pad the key to 32 bytes
		paddedKey := make([]byte, 32)
		copy(paddedKey, secretKey)
		secretKey = paddedKey
	}

	fmt.Printf("CookieService: Final secretKey length=%d\n", len(secretKey))

	return &CookieService{
		secretKey:     secretKey,
		sessionCookie: config.SessionCookieName,
		isSecure:      isSecure,
	}
}

// SessionData represents the data stored in the new session cookie
// This is a simplified version for the refactored architecture
type SessionData struct {
	SessionID   string            `json:"session_id"`
	Username    string            `json:"username"`
	AuthMethod  string            `json:"auth_method"`
	IsAdmin     bool              `json:"is_admin"`
	Permissions map[string]string `json:"permissions,omitempty"`
	Expiry      *time.Time        `json:"expiry,omitempty"`
}

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

// Cookie errors
var (
	ErrNoSessionCookie      = fmt.Errorf("no session cookie found")
	ErrInvalidSessionCookie = fmt.Errorf("invalid session cookie")
	ErrSessionExpired       = fmt.Errorf("session cookie expired")
)
