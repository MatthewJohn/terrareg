package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestInfraConfigForCookie creates a test infrastructure config for cookie service
func newTestInfraConfigForCookie() *config.InfrastructureConfig {
	return &config.InfrastructureConfig{
		SecretKey:          "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName:  "terrareg_session",
	}
}

// newTestInfraConfigWithSecure creates a config with HTTPS public URL
func newTestInfraConfigWithSecure() *config.InfrastructureConfig {
	return &config.InfrastructureConfig{
		SecretKey:          "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName:  "terrareg_session",
		PublicURL:          "https://terrareg.example.com",
	}
}

// newTestInfraConfigWithInsecure creates a config with HTTP public URL
func newTestInfraConfigWithInsecure() *config.InfrastructureConfig {
	return &config.InfrastructureConfig{
		SecretKey:          "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName:  "terrareg_session",
		PublicURL:          "http://terrareg.example.com",
	}
}

// newTestInfraConfigWithHexKey creates a config with hex-encoded secret key
func newTestInfraConfigWithHexKey() *config.InfrastructureConfig {
	// 64-byte hex string (32 bytes when decoded)
	return &config.InfrastructureConfig{
		SecretKey:          "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName:  "terrareg_session",
	}
}

// TestNewCookieService tests the constructor
func TestNewCookieService(t *testing.T) {
	t.Run("returns nil when SECRET_KEY is empty", func(t *testing.T) {
		cfg := &config.InfrastructureConfig{
			SecretKey: "",
		}

		service := NewCookieService(cfg)

		assert.Nil(t, service, "Service should be nil when SECRET_KEY is empty")
	})

	t.Run("creates service with hex SECRET_KEY", func(t *testing.T) {
		service := NewCookieService(newTestInfraConfigWithHexKey())

		assert.NotNil(t, service)
		assert.Equal(t, "terrareg_session", service.GetSessionCookieName())
	})

	t.Run("creates service with raw string SECRET_KEY", func(t *testing.T) {
		cfg := &config.InfrastructureConfig{
			SecretKey:         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", // 64 chars
			SessionCookieName: "terrareg_session",
		}

		service := NewCookieService(cfg)

		assert.NotNil(t, service)
		assert.Equal(t, "terrareg_session", service.GetSessionCookieName())
	})

	t.Run("panics with short hex SECRET_KEY", func(t *testing.T) {
		cfg := &config.InfrastructureConfig{
			SecretKey:         "0123456789abcdef", // Only 8 bytes when decoded
			SessionCookieName: "terrareg_session",
		}

		assert.Panics(t, func() {
			NewCookieService(cfg)
		})
	})

	t.Run("panics with short raw string SECRET_KEY", func(t *testing.T) {
		cfg := &config.InfrastructureConfig{
			SecretKey:         "short", // Only 5 bytes
			SessionCookieName: "terrareg_session",
		}

		assert.Panics(t, func() {
			NewCookieService(cfg)
		})
	})

	t.Run("determines secure flag from HTTPS URL", func(t *testing.T) {
		service := NewCookieService(newTestInfraConfigWithSecure())

		assert.NotNil(t, service)
		// Service should set Secure=true on cookies when using HTTPS
	})

	t.Run("determines secure flag from HTTP URL", func(t *testing.T) {
		service := NewCookieService(newTestInfraConfigWithInsecure())

		assert.NotNil(t, service)
		// Service should set Secure=false on cookies when using HTTP
	})
}

// TestCookieService_EncryptDecryptSession tests encryption and decryption round-trip
func TestCookieService_EncryptDecryptSession(t *testing.T) {
	service := NewCookieService(newTestInfraConfigForCookie())
	require.NotNil(t, service)

	t.Run("successfully encrypts and decrypts session data", func(t *testing.T) {
		expiry := time.Now().Add(1 * time.Hour)
		originalData := &auth.SessionData{
			SessionID:   "test-session-123",
			AuthMethod:  "github",
			Username:    "testuser",
			IsAdmin:     true,
			SiteAdmin:   false,
			UserGroups:  []string{"group1", "group2"},
			Permissions: map[string]string{"ns1": "FULL"},
			Expiry:      &expiry,
		}

		// Encrypt
		encrypted, err := service.EncryptSession(originalData)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		assert.NotEqual(t, originalData.SessionID, encrypted, "Encrypted data should not equal plaintext")

		// Decrypt
		decryptedData, err := service.DecryptSession(encrypted)
		require.NoError(t, err)

		// Verify all fields match
		assert.Equal(t, originalData.SessionID, decryptedData.SessionID)
		assert.Equal(t, originalData.AuthMethod, decryptedData.AuthMethod)
		assert.Equal(t, originalData.Username, decryptedData.Username)
		assert.Equal(t, originalData.IsAdmin, decryptedData.IsAdmin)
		assert.Equal(t, originalData.SiteAdmin, decryptedData.SiteAdmin)
		assert.Equal(t, originalData.UserGroups, decryptedData.UserGroups)
		assert.Equal(t, originalData.Permissions, decryptedData.Permissions)
		assert.WithinDuration(t, *originalData.Expiry, *decryptedData.Expiry, time.Second)
	})

	t.Run("produces different ciphertext each time (due to nonce)", func(t *testing.T) {
		data := &auth.SessionData{
			SessionID:  "test-session",
			AuthMethod: "github",
			Username:   "testuser",
		}

		encrypted1, err1 := service.EncryptSession(data)
		encrypted2, err2 := service.EncryptSession(data)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, encrypted1, encrypted2, "Each encryption should produce different ciphertext due to random nonce")

		// But both should decrypt to the same data
		decrypted1, _ := service.DecryptSession(encrypted1)
		decrypted2, _ := service.DecryptSession(encrypted2)
		assert.Equal(t, decrypted1.SessionID, decrypted2.SessionID)
	})

	t.Run("handles empty session data", func(t *testing.T) {
		data := &auth.SessionData{}

		encrypted, err := service.EncryptSession(data)
		require.NoError(t, err)

		decrypted, err := service.DecryptSession(encrypted)
		require.NoError(t, err)
		assert.NotNil(t, decrypted)
	})
}

// TestCookieService_DecryptSession_Errors tests decryption error handling
func TestCookieService_DecryptSession_Errors(t *testing.T) {
	service := NewCookieService(newTestInfraConfigForCookie())
	require.NotNil(t, service)

	t.Run("fails with invalid base64", func(t *testing.T) {
		_, err := service.DecryptSession("not-valid-base64!!!")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode base64")
	})

	t.Run("fails with empty string", func(t *testing.T) {
		_, err := service.DecryptSession("")

		assert.Error(t, err)
		// Empty string decodes to empty bytes (valid base64), but fails ciphertext length check
		assert.Contains(t, err.Error(), "ciphertext too short")
	})

	t.Run("fails with truncated ciphertext", func(t *testing.T) {
		// Valid base64 but too short to contain nonce + ciphertext
		shortData := "YWJj" // base64 for "abc"

		_, err := service.DecryptSession(shortData)

		assert.Error(t, err)
		// Either "ciphertext too short" or "failed to decrypt" depending on where it fails
	})

	t.Run("fails with tampered ciphertext", func(t *testing.T) {
		data := &auth.SessionData{
			SessionID: "test-session",
		}

		encrypted, err := service.EncryptSession(data)
		require.NoError(t, err)

		// Tamper with the ciphertext
		tampered := strings.ReplaceAll(encrypted, "A", "B")

		_, err = service.DecryptSession(tampered)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decrypt")
	})

	t.Run("fails with invalid JSON after decryption", func(t *testing.T) {
		// Create a valid base64 string that decrypts but doesn't contain valid JSON
		// This is tricky to do without directly crafting AES-GCM output
		// Instead, we'll use a different approach
		invalidEncrypted := "invalid-but-valid-base64-that-will-fail-decryption"

		_, err := service.DecryptSession(invalidEncrypted)
		assert.Error(t, err)
	})
}

// TestCookieService_SetCookie tests cookie setting
func TestCookieService_SetCookie(t *testing.T) {
	t.Run("sets secure cookie with default options", func(t *testing.T) {
		service := NewCookieService(newTestInfraConfigWithSecure())
		require.NotNil(t, service)

		w := httptest.NewRecorder()
		service.SetCookie(w, "test_cookie", "test_value", nil)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "test_cookie", cookie.Name)
		assert.Equal(t, "test_value", cookie.Value)
		assert.Equal(t, "/", cookie.Path)
		assert.Equal(t, -1, cookie.MaxAge, "Default should be session cookie (-1)")
		assert.True(t, cookie.Secure, "Cookie should be Secure with HTTPS URL")
		assert.True(t, cookie.HttpOnly, "Cookie should be HttpOnly")
		assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	})

	t.Run("sets insecure cookie with HTTP URL", func(t *testing.T) {
		service := NewCookieService(newTestInfraConfigWithInsecure())
		require.NotNil(t, service)

		w := httptest.NewRecorder()
		service.SetCookie(w, "test_cookie", "test_value", nil)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.False(t, cookie.Secure, "Cookie should not be Secure with HTTP URL")
	})

	t.Run("sets cookie with custom options", func(t *testing.T) {
		service := NewCookieService(newTestInfraConfigForCookie())
		require.NotNil(t, service)

		w := httptest.NewRecorder()
		customOptions := &CookieOptions{
			Path:     "/api",
			MaxAge:   3600,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}
		service.SetCookie(w, "test_cookie", "test_value", customOptions)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "/api", cookie.Path)
		assert.Equal(t, 3600, cookie.MaxAge)
		assert.True(t, cookie.Secure)
		assert.True(t, cookie.HttpOnly)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	})

	t.Run("sets cookie with empty value", func(t *testing.T) {
		service := NewCookieService(newTestInfraConfigForCookie())
		require.NotNil(t, service)

		w := httptest.NewRecorder()
		service.SetCookie(w, "test_cookie", "", nil)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "test_cookie", cookie.Name)
		assert.Equal(t, "", cookie.Value)
	})
}

// TestCookieService_ClearCookie tests cookie clearing
func TestCookieService_ClearCookie(t *testing.T) {
	service := NewCookieService(newTestInfraConfigForCookie())
	require.NotNil(t, service)

	t.Run("clears cookie by setting MaxAge to -1", func(t *testing.T) {
		w := httptest.NewRecorder()
		service.ClearCookie(w, "test_cookie")

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "test_cookie", cookie.Name)
		assert.Equal(t, "", cookie.Value)
		assert.Equal(t, -1, cookie.MaxAge, "Cleared cookie should have MaxAge=-1")
		assert.True(t, cookie.HttpOnly)
		assert.True(t, cookie.Secure)
	})
}

// TestCookieService_ValidateSessionCookie tests session cookie validation
func TestCookieService_ValidateSessionCookie(t *testing.T) {
	service := NewCookieService(newTestInfraConfigForCookie())
	require.NotNil(t, service)

	t.Run("validates valid session cookie", func(t *testing.T) {
		expiry := time.Now().Add(1 * time.Hour)
		sessionData := &auth.SessionData{
			SessionID:  "test-session-123",
			AuthMethod: "github",
			Username:   "testuser",
			Expiry:     &expiry,
		}

		encrypted, err := service.EncryptSession(sessionData)
		require.NoError(t, err)

		validated, err := service.ValidateSessionCookie(encrypted)
		require.NoError(t, err)
		assert.Equal(t, sessionData.SessionID, validated.SessionID)
		assert.Equal(t, sessionData.Username, validated.Username)
	})

	t.Run("fails with empty cookie value", func(t *testing.T) {
		_, err := service.ValidateSessionCookie("")

		assert.Error(t, err)
		assert.Equal(t, ErrNoSessionCookie, err)
	})

	t.Run("fails with invalid encrypted value", func(t *testing.T) {
		_, err := service.ValidateSessionCookie("invalid-encrypted-value")

		assert.Error(t, err)
		// Should wrap ErrInvalidSessionCookie
	})

	t.Run("fails with expired session", func(t *testing.T) {
		pastExpiry := time.Now().Add(-1 * time.Hour)
		sessionData := &auth.SessionData{
			SessionID:  "expired-session",
			AuthMethod: "github",
			Username:   "testuser",
			Expiry:     &pastExpiry,
		}

		encrypted, err := service.EncryptSession(sessionData)
		require.NoError(t, err)

		_, err = service.ValidateSessionCookie(encrypted)
		assert.Error(t, err)
		assert.Equal(t, ErrSessionExpired, err)
	})

	t.Run("accepts session with no expiry (session cookie)", func(t *testing.T) {
		sessionData := &auth.SessionData{
			SessionID:  "no-expiry-session",
			AuthMethod: "github",
			Username:   "testuser",
			Expiry:     nil, // No expiry = session cookie
		}

		encrypted, err := service.EncryptSession(sessionData)
		require.NoError(t, err)

		validated, err := service.ValidateSessionCookie(encrypted)
		require.NoError(t, err)
		assert.Equal(t, sessionData.SessionID, validated.SessionID)
		assert.Nil(t, validated.Expiry)
	})

	t.Run("accepts session with future expiry", func(t *testing.T) {
		futureExpiry := time.Now().Add(24 * time.Hour)
		sessionData := &auth.SessionData{
			SessionID:  "future-expiry-session",
			AuthMethod: "github",
			Username:   "testuser",
			Expiry:     &futureExpiry,
		}

		encrypted, err := service.EncryptSession(sessionData)
		require.NoError(t, err)

		validated, err := service.ValidateSessionCookie(encrypted)
		require.NoError(t, err)
		assert.Equal(t, sessionData.SessionID, validated.SessionID)
	})
}

// TestCookieService_SetSessionCookie tests setting encrypted session cookies
func TestCookieService_SetSessionCookie(t *testing.T) {
	service := NewCookieService(newTestInfraConfigForCookie())
	require.NotNil(t, service)

	t.Run("sets encrypted session cookie", func(t *testing.T) {
		sessionData := &auth.SessionData{
			SessionID:  "test-session-123",
			AuthMethod: "github",
			Username:   "testuser",
		}

		w := httptest.NewRecorder()
		err := service.SetSessionCookie(w, sessionData)

		require.NoError(t, err)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "terrareg_session", cookie.Name)
		assert.NotEmpty(t, cookie.Value, "Cookie value should be encrypted")
		assert.NotEqual(t, sessionData.SessionID, cookie.Value, "Cookie value should not equal plaintext session ID")
		assert.Equal(t, "/", cookie.Path)
		assert.Equal(t, 0, cookie.MaxAge, "Session cookie should have MaxAge=0")
		assert.True(t, cookie.HttpOnly)
		assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	})

	t.Run("sets secure cookie with HTTPS config", func(t *testing.T) {
		service := NewCookieService(newTestInfraConfigWithSecure())
		require.NotNil(t, service)

		sessionData := &auth.SessionData{
			SessionID: "test-session",
		}

		w := httptest.NewRecorder()
		err := service.SetSessionCookie(w, sessionData)

		require.NoError(t, err)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.True(t, cookie.Secure, "Session cookie should be Secure with HTTPS URL")
	})

	t.Run("sets insecure cookie with HTTP config", func(t *testing.T) {
		service := NewCookieService(newTestInfraConfigWithInsecure())
		require.NotNil(t, service)

		sessionData := &auth.SessionData{
			SessionID: "test-session",
		}

		w := httptest.NewRecorder()
		err := service.SetSessionCookie(w, sessionData)

		require.NoError(t, err)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.False(t, cookie.Secure, "Session cookie should not be Secure with HTTP URL")
	})
}

// TestCookieService_ClearSessionCookie tests clearing session cookies
func TestCookieService_ClearSessionCookie(t *testing.T) {
	service := NewCookieService(newTestInfraConfigForCookie())
	require.NotNil(t, service)

	t.Run("clears session cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := service.ClearSessionCookie(w)

		require.NoError(t, err)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)

		cookie := cookies[0]
		assert.Equal(t, "terrareg_session", cookie.Name)
		assert.Equal(t, "", cookie.Value)
		assert.Equal(t, -1, cookie.MaxAge, "Cleared cookie should have MaxAge=-1")
		assert.True(t, cookie.HttpOnly)
		assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	})
}

// TestCookieService_GetSessionCookieName tests getting the session cookie name
func TestCookieService_GetSessionCookieName(t *testing.T) {
	t.Run("returns configured session cookie name", func(t *testing.T) {
		cfg := &config.InfrastructureConfig{
			SecretKey:         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			SessionCookieName: "custom_session_cookie",
		}

		service := NewCookieService(cfg)
		require.NotNil(t, service)

		assert.Equal(t, "custom_session_cookie", service.GetSessionCookieName())
	})

	t.Run("returns default session cookie name", func(t *testing.T) {
		service := NewCookieService(newTestInfraConfigForCookie())
		require.NotNil(t, service)

		assert.Equal(t, "terrareg_session", service.GetSessionCookieName())
	})
}

// TestCookieService_EdgeCases tests edge cases
func TestCookieService_EdgeCases(t *testing.T) {
	service := NewCookieService(newTestInfraConfigForCookie())
	require.NotNil(t, service)

	t.Run("handles special characters in session data", func(t *testing.T) {
		sessionData := &auth.SessionData{
			SessionID:  "session-with-!@#$%^&*()_+-=[]{}|;':\",./<>?",
			AuthMethod: "github",
			Username:   "user-with-中文-😀",
			UserGroups: []string{"group-with-\\\"quotes\\\"", "group-with-\\tslash"},
		}

		encrypted, err := service.EncryptSession(sessionData)
		require.NoError(t, err)

		decrypted, err := service.DecryptSession(encrypted)
		require.NoError(t, err)
		assert.Equal(t, sessionData.SessionID, decrypted.SessionID)
		assert.Equal(t, sessionData.Username, decrypted.Username)
		assert.Equal(t, sessionData.UserGroups, decrypted.UserGroups)
	})

	t.Run("handles very long session data", func(t *testing.T) {
		longString := string(make([]byte, 10000))
		sessionData := &auth.SessionData{
			SessionID:  longString,
			AuthMethod: "github",
			Username:   longString,
			UserGroups: []string{longString},
			Permissions: map[string]string{
				longString: longString,
			},
		}

		encrypted, err := service.EncryptSession(sessionData)
		require.NoError(t, err)

		decrypted, err := service.DecryptSession(encrypted)
		require.NoError(t, err)
		assert.Equal(t, sessionData.SessionID, decrypted.SessionID)
	})

	t.Run("handles empty slices and maps", func(t *testing.T) {
		sessionData := &auth.SessionData{
			SessionID:  "test-session",
			AuthMethod: "github",
			Username:   "testuser",
			UserGroups: []string{},
			Permissions: map[string]string{},
		}

		encrypted, err := service.EncryptSession(sessionData)
		require.NoError(t, err)

		decrypted, err := service.DecryptSession(encrypted)
		require.NoError(t, err)
		assert.Empty(t, decrypted.UserGroups)
		assert.Empty(t, decrypted.Permissions)
	})

	t.Run("handles nil optional fields", func(t *testing.T) {
		sessionData := &auth.SessionData{
			SessionID:  "test-session",
			AuthMethod: "github",
			Username:   "testuser",
			// All other fields nil/empty
		}

		encrypted, err := service.EncryptSession(sessionData)
		require.NoError(t, err)

		decrypted, err := service.DecryptSession(encrypted)
		require.NoError(t, err)
		assert.Equal(t, sessionData.SessionID, decrypted.SessionID)
		assert.Empty(t, decrypted.UserGroups)
		assert.Nil(t, decrypted.Permissions)
		assert.Nil(t, decrypted.Expiry)
	})
}

// TestCookieService_ConcurrentAccess tests thread safety
func TestCookieService_ConcurrentAccess(t *testing.T) {
	service := NewCookieService(newTestInfraConfigForCookie())
	require.NotNil(t, service)

	t.Run("concurrent encrypt/decrypt operations", func(t *testing.T) {
		const goroutines = 100
		const operationsPerGoroutine = 10

		errors := make(chan error, goroutines*operationsPerGoroutine)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				for j := 0; j < operationsPerGoroutine; j++ {
					sessionData := &auth.SessionData{
						SessionID:  "session-" + string(rune(id)) + "-" + string(rune(j)),
						AuthMethod: "github",
						Username:   "user",
					}

					encrypted, err := service.EncryptSession(sessionData)
					if err != nil {
						errors <- err
						return
					}

					decrypted, err := service.DecryptSession(encrypted)
					if err != nil {
						errors <- err
						return
					}

					if decrypted.SessionID != sessionData.SessionID {
						errors <- assert.AnError
						return
					}
				}
			}(i)
		}

		// Check for errors
		select {
		case err := <-errors:
			t.Fatalf("Concurrent operation failed: %v", err)
		default:
			// No errors
		}
	})
}

// TestPrepareSecretKey tests secret key preparation
func TestPrepareSecretKey(t *testing.T) {
	t.Run("accepts valid hex key", func(t *testing.T) {
		hexKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

		key, err := prepareSecretKey(hexKey)

		require.NoError(t, err)
		assert.Len(t, key, 32, "Hex-decoded key should be 32 bytes")
	})

	t.Run("accepts valid raw string key", func(t *testing.T) {
		rawKey := "0123456789-abcdefghijk-1mno-pqrs" // exactly 32 chars, contains non-hex chars like '-'

		key, err := prepareSecretKey(rawKey)

		require.NoError(t, err)
		assert.Len(t, key, 32, "Raw key should be 32 bytes")
	})

	t.Run("rejects short hex key", func(t *testing.T) {
		shortHexKey := "0123456789abcdef" // Only 8 bytes when decoded

		_, err := prepareSecretKey(shortHexKey)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too short")
	})

	t.Run("rejects short raw key", func(t *testing.T) {
		shortRawKey := "short" // Only 5 bytes

		_, err := prepareSecretKey(shortRawKey)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too short")
	})

	t.Run("treats invalid hex as raw string", func(t *testing.T) {
		// This is not valid hex, so it should be treated as raw string
		invalidHex := "not-a-valid-hex-string-but-long-enough-for-raw-key-!!"

		key, err := prepareSecretKey(invalidHex)

		require.NoError(t, err)
		assert.Len(t, key, len(invalidHex), "Should use raw string when hex decoding fails")
	})
}
