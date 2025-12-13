package model

import (
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationTokenType_String(t *testing.T) {
	tests := []struct {
		name     string
		tokenType AuthenticationTokenType
		expected string
	}{
		{"Admin", AuthenticationTokenTypeAdmin, "admin"},
		{"Upload", AuthenticationTokenTypeUpload, "upload"},
		{"Publish", AuthenticationTokenTypePublish, "publish"},
		{"Unknown", AuthenticationTokenType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.tokenType.String())
		})
	}
}

func TestAuthenticationTokenType_ToAuthMethodType(t *testing.T) {
	tests := []struct {
		name     string
		tokenType AuthenticationTokenType
		expected auth.AuthMethodType
	}{
		{"Admin", AuthenticationTokenTypeAdmin, auth.AuthMethodAdminApiKey},
		{"Upload", AuthenticationTokenTypeUpload, auth.AuthMethodUploadApiKey},
		{"Publish", AuthenticationTokenTypePublish, auth.AuthMethodPublishApiKey},
		{"Unknown", AuthenticationTokenType(999), auth.AuthMethodNotAuthenticated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.tokenType.ToAuthMethodType())
		})
	}
}

func TestAuthenticationTokenTypeFromString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  AuthenticationTokenType
		expectErr bool
	}{
		{"Admin lowercase", "admin", AuthenticationTokenTypeAdmin, false},
		{"Admin uppercase", "ADMIN", AuthenticationTokenTypeAdmin, false},
		{"Upload lowercase", "upload", AuthenticationTokenTypeUpload, false},
		{"Publish lowercase", "publish", AuthenticationTokenTypePublish, false},
		{"Invalid type", "invalid", AuthenticationTokenTypeAdmin, true},
		{"Empty string", "", AuthenticationTokenTypeAdmin, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AuthenticationTokenTypeFromString(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid token type")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNewAuthenticationToken(t *testing.T) {
	// Create a test namespace
	namespace, err := model.NewNamespace("testns", nil, model.NamespaceTypeGithubOrg)
	require.NoError(t, err)

	t.Run("Valid admin token", func(t *testing.T) {
		token, err := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"Test admin token",
			nil,
			nil,
			"testuser",
		)

		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, AuthenticationTokenTypeAdmin, token.TokenType())
		assert.Equal(t, "Test admin token", token.Description())
		assert.Nil(t, token.Namespace())
		assert.True(t, token.IsActive())
		assert.Equal(t, "testuser", token.CreatedBy())
		assert.NotEmpty(t, token.TokenValue())
		assert.True(t, token.IsAdmin())
		assert.False(t, token.IsUpload())
		assert.False(t, token.IsPublish())
		assert.False(t, token.HasNamespace())
	})

	t.Run("Valid upload token", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		token, err := NewAuthenticationToken(
			AuthenticationTokenTypeUpload,
			"Test upload token",
			nil,
			&expiresAt,
			"testuser",
		)

		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, AuthenticationTokenTypeUpload, token.TokenType())
		assert.Equal(t, "Test upload token", token.Description())
		assert.NotNil(t, token.ExpiresAt())
		assert.True(t, token.IsActive())
		assert.True(t, token.IsUpload())
		assert.False(t, token.IsAdmin())
		assert.False(t, token.IsPublish())
	})

	t.Run("Valid publish token with namespace", func(t *testing.T) {
		token, err := NewAuthenticationToken(
			AuthenticationTokenTypePublish,
			"Test publish token",
			namespace,
			nil,
			"testuser",
		)

		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, AuthenticationTokenTypePublish, token.TokenType())
		assert.Equal(t, "Test publish token", token.Description())
		assert.Equal(t, namespace, token.Namespace())
		assert.True(t, token.IsActive())
		assert.True(t, token.IsPublish())
		assert.False(t, token.IsAdmin())
		assert.False(t, token.IsUpload())
		assert.True(t, token.HasNamespace())
	})

	t.Run("Publish token without namespace", func(t *testing.T) {
		token, err := NewAuthenticationToken(
			AuthenticationTokenTypePublish,
			"Test publish token",
			nil,
			nil,
			"testuser",
		)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "namespace is required for publish tokens")
	})

	t.Run("Admin token with namespace", func(t *testing.T) {
		token, err := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"Test admin token",
			namespace,
			nil,
			"testuser",
		)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "namespace is only allowed for publish tokens")
	})

	t.Run("Empty description", func(t *testing.T) {
		token, err := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"",
			nil,
			nil,
			"testuser",
		)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Equal(t, ErrDescriptionRequired, err)
	})

	t.Run("Empty created by", func(t *testing.T) {
		token, err := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"Test token",
			nil,
			nil,
			"",
		)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Equal(t, ErrCreatedByRequired, err)
	})

	t.Run("Description too long", func(t *testing.T) {
		longDesc := ""
		for i := 0; i < 256; i++ {
			longDesc += "a"
		}

		token, err := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			longDesc,
			nil,
			nil,
			"testuser",
		)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Equal(t, ErrDescriptionTooLong, err)
	})

	t.Run("Expiration in past", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		token, err := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"Test token",
			nil,
			&past,
			"testuser",
		)

		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Equal(t, ErrExpirationInPast, err)
	})
}

func TestAuthenticationToken_Validate(t *testing.T) {
	t.Run("Valid active token", func(t *testing.T) {
		token, _ := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"Test token",
			nil,
			nil,
			"testuser",
		)

		err := token.Validate()
		assert.NoError(t, err)
	})

	t.Run("Inactive token", func(t *testing.T) {
		token := ReconstructAuthenticationToken(
			1,
			AuthenticationTokenTypeAdmin,
			"tokenvalue",
			nil,
			"Test token",
			time.Now(),
			nil,
			false,
			"testuser",
		)

		err := token.Validate()
		assert.Equal(t, ErrTokenInactive, err)
	})

	t.Run("Expired token", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		token := ReconstructAuthenticationToken(
			1,
			AuthenticationTokenTypeAdmin,
			"tokenvalue",
			nil,
			"Test token",
			time.Now(),
			&past,
			true,
			"testuser",
		)

		err := token.Validate()
		assert.Equal(t, ErrTokenExpired, err)
	})
}

func TestAuthenticationToken_CanAccessNamespace(t *testing.T) {
	t.Run("Admin token can access any namespace", func(t *testing.T) {
		token, _ := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"Admin token",
			nil,
			nil,
			"admin",
		)

		assert.True(t, token.CanAccessNamespace("anyns"))
		assert.True(t, token.CanAccessNamespace("otherns"))
	})

	t.Run("Upload token can access any namespace", func(t *testing.T) {
		token, _ := NewAuthenticationToken(
			AuthenticationTokenTypeUpload,
			"Upload token",
			nil,
			nil,
			"uploader",
		)

		assert.True(t, token.CanAccessNamespace("anyns"))
		assert.True(t, token.CanAccessNamespace("otherns"))
	})

	t.Run("Publish token can only access its namespace", func(t *testing.T) {
		namespace, _ := model.NewNamespace("testns", nil, model.NamespaceTypeGithubOrg)
		token, _ := NewAuthenticationToken(
			AuthenticationTokenTypePublish,
			"Publish token",
			namespace,
			nil,
			"publisher",
		)

		assert.True(t, token.CanAccessNamespace("testns"))
		assert.False(t, token.CanAccessNamespace("otherns"))
	})
}

func TestAuthenticationToken_Revoke(t *testing.T) {
	token, _ := NewAuthenticationToken(
		AuthenticationTokenTypeAdmin,
		"Test token",
		nil,
		nil,
		"testuser",
	)

	t.Run("Revoke active token", func(t *testing.T) {
		err := token.Revoke()
		assert.NoError(t, err)
		assert.False(t, token.IsActive())
	})

	t.Run("Revoke already revoked token", func(t *testing.T) {
		err := token.Revoke()
		assert.Equal(t, ErrTokenAlreadyRevoked, err)
	})
}

func TestAuthenticationToken_IsExpired(t *testing.T) {
	t.Run("Token without expiration", func(t *testing.T) {
		token, _ := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"Test token",
			nil,
			nil,
			"testuser",
		)

		assert.False(t, token.IsExpired())
	})

	t.Run("Token with future expiration", func(t *testing.T) {
		future := time.Now().Add(1 * time.Hour)
		token, _ := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"Test token",
			nil,
			&future,
			"testuser",
		)

		assert.False(t, token.IsExpired())
	})

	t.Run("Token with past expiration", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		token := ReconstructAuthenticationToken(
			1,
			AuthenticationTokenTypeAdmin,
			"tokenvalue",
			nil,
			"Test token",
			time.Now(),
			&past,
			true,
			"testuser",
		)

		assert.True(t, token.IsExpired())
	})
}

func TestAuthenticationToken_Getters(t *testing.T) {
	namespace, _ := model.NewNamespace("testns", nil, model.NamespaceTypeGithubOrg)
	expiresAt := time.Now().Add(1 * time.Hour)
	token, _ := NewAuthenticationToken(
		AuthenticationTokenTypePublish,
		"Test token",
		namespace,
		&expiresAt,
		"testuser",
	)

	assert.Equal(t, AuthenticationTokenTypePublish, token.TokenType())
	assert.Equal(t, "Test token", token.Description())
	assert.Equal(t, namespace, token.Namespace())
	assert.Equal(t, "testuser", token.CreatedBy())
	assert.True(t, token.IsActive())
	assert.Equal(t, expiresAt, *token.ExpiresAt())
	assert.True(t, token.IsPublish())
	assert.True(t, token.HasNamespace())
	assert.Equal(t, auth.AuthMethodPublishApiKey, token.GetAuthMethod())
	assert.Equal(t, "Test token", token.GetDisplayName())
}

func TestAuthenticationToken_GetDisplayName(t *testing.T) {
	t.Run("With description", func(t *testing.T) {
		token, _ := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"My custom token",
			nil,
			nil,
			"testuser",
		)
		assert.Equal(t, "My custom token", token.GetDisplayName())
	})

	t.Run("Without description", func(t *testing.T) {
		token, _ := NewAuthenticationToken(
			AuthenticationTokenTypeUpload,
			"Upload Token",
			nil,
			nil,
			"testuser",
		)
		assert.Equal(t, "Upload Token", token.GetDisplayName())
	})
}

func TestAuthenticationToken_TimeToExpiration(t *testing.T) {
	t.Run("Token without expiration", func(t *testing.T) {
		token, _ := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"Test token",
			nil,
			nil,
			"testuser",
		)

		assert.Equal(t, time.Duration(0), token.TimeToExpiration())
	})

	t.Run("Token with expiration", func(t *testing.T) {
		expiresAt := time.Now().Add(1 * time.Hour)
		token, _ := NewAuthenticationToken(
			AuthenticationTokenTypeAdmin,
			"Test token",
			nil,
			&expiresAt,
			"testuser",
		)

		// Should be approximately 1 hour (allowing for small time differences)
		duration := token.TimeToExpiration()
		assert.True(t, duration > 59*time.Minute)
		assert.True(t, duration <= time.Hour)
	})
}

func TestAuthenticationToken_String(t *testing.T) {
	namespace, _ := model.NewNamespace("testns", nil, model.NamespaceTypeGithubOrg)
	token, _ := NewAuthenticationToken(
		AuthenticationTokenTypePublish,
		"Test token",
		namespace,
		nil,
		"testuser",
	)

	str := token.String()
	assert.Contains(t, str, "AuthenticationToken")
	assert.Contains(t, str, "publish")
	assert.Contains(t, str, "testns")
	assert.Contains(t, str, "true")
	// Should not contain the actual token value
	assert.NotContains(t, str, token.TokenValue())
}