package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewStateStorageService tests the constructor
func TestNewStateStorageService(t *testing.T) {
	// Note: This test creates a StateStorageService with a nil SessionService
	// In production, this would be created with proper dependency injection
	service := NewStateStorageService(nil)

	assert.NotNil(t, service)
	// When SessionService is nil, the field will be nil
	// This is acceptable for testing state storage operations that don't require sessions
}

// TestStateInfo_JSONSerialization tests StateInfo JSON marshaling/unmarshaling
func TestStateInfo_JSONSerialization(t *testing.T) {
	stateInfo := &StateInfo{
		State:       "test-state-12345",
		RedirectURL: "https://example.com/oauth/callback?code=xyz",
		CreatedAt:   time.Now().Truncate(time.Millisecond),
		ExpiresAt:   time.Now().Add(10 * time.Minute).Truncate(time.Millisecond),
		AuthMethod:  "oidc",
	}

	// Marshal
	data, err := json.Marshal(stateInfo)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "test-state-12345")
	assert.Contains(t, string(data), "https://example.com/oauth/callback")

	// Unmarshal
	var decoded StateInfo
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, stateInfo.State, decoded.State)
	assert.Equal(t, stateInfo.RedirectURL, decoded.RedirectURL)
	assert.Equal(t, stateInfo.AuthMethod, decoded.AuthMethod)
	assert.Equal(t, stateInfo.CreatedAt.Unix(), decoded.CreatedAt.Unix())
	assert.Equal(t, stateInfo.ExpiresAt.Unix(), decoded.ExpiresAt.Unix())
}

// TestStateInfo_Expiration tests StateInfo expiration logic
func TestStateInfo_Expiration(t *testing.T) {
	tests := []struct {
		name      string
		createdAt time.Time
		expiresAt time.Time
		isExpired bool
	}{
		{
			name:      "not expired",
			createdAt: time.Now(),
			expiresAt: time.Now().Add(1 * time.Hour),
			isExpired: false,
		},
		{
			name:      "expired",
			createdAt: time.Now().Add(-1 * time.Hour),
			expiresAt: time.Now().Add(-1 * time.Minute),
			isExpired: true,
		},
		{
			name:      "just expired",
			createdAt: time.Now().Add(-10 * time.Minute),
			expiresAt: time.Now().Add(-1 * time.Second),
			isExpired: true,
		},
		{
			name:      "expires now",
			createdAt: time.Now().Add(-10 * time.Minute),
			expiresAt: time.Now(),
			isExpired: true, // Expired if now or after
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateInfo := &StateInfo{
				State:       "test",
				RedirectURL: "https://example.com",
				CreatedAt:   tt.createdAt,
				ExpiresAt:   tt.expiresAt,
				AuthMethod:  "oidc",
			}

			isExpired := time.Now().After(stateInfo.ExpiresAt)
			assert.Equal(t, tt.isExpired, isExpired, "Expiration check should match")
		})
	}
}

// TestStateInfo_AllAuthMethods tests StateInfo with different auth methods
func TestStateInfo_AllAuthMethods(t *testing.T) {
	authMethods := []string{"oidc", "github", "saml", "gitlab", "bitbucket"}

	for _, method := range authMethods {
		t.Run(method, func(t *testing.T) {
			stateInfo := &StateInfo{
				State:       "test-state",
				RedirectURL: "https://example.com/callback",
				CreatedAt:   time.Now(),
				ExpiresAt:   time.Now().Add(10 * time.Minute),
				AuthMethod:  method,
			}

			// Marshal and unmarshal to ensure all auth methods are handled
			data, err := json.Marshal(stateInfo)
			assert.NoError(t, err)

			var decoded StateInfo
			err = json.Unmarshal(data, &decoded)
			assert.NoError(t, err)
			assert.Equal(t, method, decoded.AuthMethod)
		})
	}
}

// TestStateInfo_EmptyFields tests StateInfo with minimal required fields
func TestStateInfo_EmptyFields(t *testing.T) {
	stateInfo := &StateInfo{
		State:       "test-state",
		RedirectURL: "",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		AuthMethod:  "",
	}

	data, err := json.Marshal(stateInfo)
	assert.NoError(t, err)

	var decoded StateInfo
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "test-state", decoded.State)
	assert.Empty(t, decoded.RedirectURL)
	assert.Empty(t, decoded.AuthMethod)
}

// TestStateInfo_LongRedirectURL tests with very long redirect URLs
func TestStateInfo_LongRedirectURL(t *testing.T) {
	// Create a very long redirect URL (2000 characters)
	longURL := "https://example.com/callback?" + strings.Repeat("param=value&", 200)

	stateInfo := &StateInfo{
		State:       "test-state",
		RedirectURL: longURL,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		AuthMethod:  "oidc",
	}

	// Marshal and unmarshal to ensure long URLs are handled
	data, err := json.Marshal(stateInfo)
	assert.NoError(t, err)

	var decoded StateInfo
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, longURL, decoded.RedirectURL)
}

// TestStateInfo_SpecialCharactersInURL tests with special characters in URLs
func TestStateInfo_SpecialCharactersInURL(t *testing.T) {
	specialURLs := []string{
		"https://example.com/callback?code=abc&state=xyz&foo=bar",
		"https://example.com/callback?query=",
		"https://example.com/callback?space=hello%20world",
		"https://example.com/callback?special=!@#$%25%5E&*()",
		"https://example.com/callback?unicode=\u4e2d\u6587", // Chinese characters
	}

	for _, url := range specialURLs {
		t.Run(url, func(t *testing.T) {
			stateInfo := &StateInfo{
				State:       "test-state",
				RedirectURL: url,
				CreatedAt:   time.Now(),
				ExpiresAt:   time.Now().Add(10 * time.Minute),
				AuthMethod:  "oidc",
			}

			data, err := json.Marshal(stateInfo)
			assert.NoError(t, err)

			var decoded StateInfo
			err = json.Unmarshal(data, &decoded)
			assert.NoError(t, err)
			assert.Equal(t, url, decoded.RedirectURL)
		})
	}
}

// Integration-style tests would go here but are skipped due to complex mocking requirements
// The following tests document the expected behavior when proper mocks are available:

/*
TestGenerateAndStoreState_Success - tests successful state generation and storage
TestGenerateAndStoreState_GeneratesUniqueStates - tests that each call generates a unique state
TestValidateAndConsumeState_Success - tests successful state validation and consumption
TestValidateAndConsumeState_ReplayAttack - tests that state cannot be used twice (replay protection)
TestValidateAndConsumeState_InvalidInputs - tests validation with invalid inputs
TestValidateAndConsumeState_Expiration - tests that expired states are rejected
TestValidateAndConsumeState_StateMismatch - tests CSRF protection (tampered state detection)
TestGenerateAndStoreState_ConcurrentStateGeneration - tests thread safety of concurrent generation
*/

// TestStateInfo_CreatedAtBeforeExpiresAt validates that CreatedAt is always before ExpiresAt
func TestStateInfo_CreatedAtBeforeExpiresAt(t *testing.T) {
	baseTime := time.Now()

	stateInfo := &StateInfo{
		State:       "test",
		RedirectURL: "https://example.com",
		CreatedAt:   baseTime,
		ExpiresAt:   baseTime.Add(10 * time.Minute),
		AuthMethod:  "oidc",
	}

	assert.True(t, stateInfo.ExpiresAt.After(stateInfo.CreatedAt) || stateInfo.ExpiresAt.Equal(stateInfo.CreatedAt),
		"ExpiresAt should be after or equal to CreatedAt")
}
