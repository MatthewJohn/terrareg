package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// TestNewTerraformSession tests creating a new TerraformSession
func TestNewTerraformSession(t *testing.T) {
	tests := []struct {
		name                string
		authorizationCode   string
		clientID            string
		redirectURI         string
		scope               string
		state               string
		codeChallenge       *string
		codeChallengeMethod *string
		nonce               *string
		expectError         bool
	}{
		{
			name:                "valid session with all fields",
			authorizationCode:   "auth-code-123",
			clientID:            "terraform-cli",
			redirectURI:         "http://localhost:10000/login",
			scope:               "openid",
			state:               "state-123",
			codeChallenge:       strPtr("challenge-123"),
			codeChallengeMethod: strPtr("S256"),
			nonce:               strPtr("nonce-123"),
			expectError:         false,
		},
		{
			name:                "valid session with optional fields nil",
			authorizationCode:   "auth-code-456",
			clientID:            "terraform-cli",
			redirectURI:         "http://localhost:10001/login",
			scope:               "",
			state:               "",
			codeChallenge:       nil,
			codeChallengeMethod: nil,
			nonce:               nil,
			expectError:         false,
		},
		{
			name:              "empty authorization code",
			authorizationCode: "",
			clientID:          "terraform-cli",
			redirectURI:       "http://localhost:10000/login",
			expectError:       true,
		},
		{
			name:              "empty client ID",
			authorizationCode: "auth-code-123",
			clientID:          "",
			redirectURI:       "http://localhost:10000/login",
			expectError:       true,
		},
		{
			name:              "empty redirect URI",
			authorizationCode: "auth-code-123",
			clientID:          "terraform-cli",
			redirectURI:       "",
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := NewTerraformSession(
				tt.authorizationCode,
				tt.clientID,
				tt.redirectURI,
				tt.scope,
				tt.state,
				tt.codeChallenge,
				tt.codeChallengeMethod,
				tt.nonce,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, session)
				assert.Equal(t, shared.ErrInvalidInput, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
				assert.NotEmpty(t, session.ID())
				assert.Equal(t, tt.authorizationCode, session.AuthorizationCode())
				assert.Equal(t, tt.clientID, session.ClientID())
				assert.Equal(t, tt.redirectURI, session.RedirectURI())
				assert.Equal(t, tt.scope, session.Scope())
				assert.Equal(t, tt.state, session.State())
				assert.Equal(t, tt.codeChallenge, session.CodeChallenge())
				assert.Equal(t, tt.codeChallengeMethod, session.CodeChallengeMethod())
				assert.Equal(t, tt.nonce, session.Nonce())
				assert.False(t, session.ExpiresAt().IsZero())
				assert.False(t, session.CreatedAt().IsZero())
				assert.Nil(t, session.ExchangedAt())
				assert.Empty(t, session.SubjectIdentifier())
			}
		})
	}
}

// TestTerraformSession_IsExpired tests the IsExpired method
func TestTerraformSession_IsExpired(t *testing.T) {
	t.Run("non-expired session", func(t *testing.T) {
		session, err := NewTerraformSession("auth-code", "client-id", "http://localhost", "", "", nil, nil, nil)
		require.NoError(t, err)

		// Session should not be expired immediately after creation
		assert.False(t, session.IsExpired())

		// Check expiration time is about 10 minutes from now
		expectedExpiry := time.Now().Add(10 * time.Minute)
		assert.WithinDuration(t, expectedExpiry, session.ExpiresAt(), time.Second)
	})

	t.Run("expired session using reconstruction", func(t *testing.T) {
		// Create an expired session by setting expiry in the past
		pastTime := time.Now().Add(-1 * time.Hour)
		session := ReconstructTerraformSession(
			"session-id",
			"auth-code",
			"client-id",
			"http://localhost",
			"openid",
			"state",
			nil,
			nil,
			nil,
			pastTime,
			time.Now().Add(-2*time.Hour),
			nil,
			nil,
			"",
		)

		assert.True(t, session.IsExpired())
	})
}

// TestTerraformSession_IsExchanged tests the IsExchanged method
func TestTerraformSession_IsExchanged(t *testing.T) {
	t.Run("newly created session", func(t *testing.T) {
		session, err := NewTerraformSession("auth-code", "client-id", "http://localhost", "", "", nil, nil, nil)
		require.NoError(t, err)

		assert.False(t, session.IsExchanged())
		assert.Nil(t, session.ExchangedAt())
		assert.Nil(t, session.AccessTokenID())
	})

	t.Run("exchanged session", func(t *testing.T) {
		session, err := NewTerraformSession("auth-code", "client-id", "http://localhost", "", "", nil, nil, nil)
		require.NoError(t, err)

		now := time.Now()
		accessTokenID := "token-123"

		// Manually set exchanged state for testing
		reconstructed := ReconstructTerraformSession(
			session.ID(),
			session.AuthorizationCode(),
			session.ClientID(),
			session.RedirectURI(),
			session.Scope(),
			session.State(),
			session.CodeChallenge(),
			session.CodeChallengeMethod(),
			session.Nonce(),
			session.ExpiresAt(),
			session.CreatedAt(),
			&now,
			&accessTokenID,
			"",
		)

		assert.True(t, reconstructed.IsExchanged())
		assert.NotNil(t, reconstructed.ExchangedAt())
		assert.Equal(t, &accessTokenID, reconstructed.AccessTokenID())
	})
}

// TestTerraformSession_Exchange tests the Exchange method
func TestTerraformSession_Exchange(t *testing.T) {
	t.Run("successful exchange", func(t *testing.T) {
		session, err := NewTerraformSession("auth-code", "client-id", "http://localhost", "", "", nil, nil, nil)
		require.NoError(t, err)

		accessTokenID := "token-123"
		err = session.Exchange(accessTokenID)

		assert.NoError(t, err)
		assert.True(t, session.IsExchanged())
		assert.NotNil(t, session.ExchangedAt())
		assert.Equal(t, &accessTokenID, session.AccessTokenID())
	})

	t.Run("exchange already exchanged session", func(t *testing.T) {
		session, err := NewTerraformSession("auth-code", "client-id", "http://localhost", "", "", nil, nil, nil)
		require.NoError(t, err)

		accessTokenID := "token-123"
		err = session.Exchange(accessTokenID)
		require.NoError(t, err)

		// Try to exchange again
		err = session.Exchange("token-456")
		assert.Error(t, err)
		assert.Equal(t, ErrAuthorizationCodeAlreadyExchanged, err)
	})

	t.Run("exchange expired session", func(t *testing.T) {
		// Create an expired session
		pastTime := time.Now().Add(-1 * time.Hour)
		session := ReconstructTerraformSession(
			"session-id",
			"auth-code",
			"client-id",
			"http://localhost",
			"openid",
			"state",
			nil,
			nil,
			nil,
			pastTime,
			time.Now().Add(-2*time.Hour),
			nil,
			nil,
			"",
		)

		err := session.Exchange("token-123")
		assert.Error(t, err)
		assert.Equal(t, ErrAuthorizationCodeExpired, err)
	})
}

// TestTerraformSession_SetSubjectIdentifier tests setting subject identifier
func TestTerraformSession_SetSubjectIdentifier(t *testing.T) {
	session, err := NewTerraformSession("auth-code", "client-id", "http://localhost", "", "", nil, nil, nil)
	require.NoError(t, err)

	assert.Empty(t, session.SubjectIdentifier())

	session.SetSubjectIdentifier("user-123")
	assert.Equal(t, "user-123", session.SubjectIdentifier())

	session.SetSubjectIdentifier("user-456")
	assert.Equal(t, "user-456", session.SubjectIdentifier())
}

// TestNewTerraformAccessToken tests creating a new TerraformAccessToken
func TestNewTerraformAccessToken(t *testing.T) {
	tests := []struct {
		name                string
		tokenValue          string
		tokenType           string
		expiresAt           time.Time
		scopes              []string
		subjectIdentifier   string
		authorizationCodeID *string
		expectError         bool
	}{
		{
			name:                "valid access token",
			tokenValue:          "access-token-123",
			tokenType:           "Bearer",
			expiresAt:           time.Now().Add(1 * time.Hour),
			scopes:              []string{"openid", "profile"},
			subjectIdentifier:   "user-123",
			authorizationCodeID: strPtr("auth-code-123"),
			expectError:         false,
		},
		{
			name:                "valid token without authorization code ID",
			tokenValue:          "access-token-456",
			tokenType:           "Bearer",
			expiresAt:           time.Now().Add(30 * time.Minute),
			scopes:              []string{"openid"},
			subjectIdentifier:   "user-456",
			authorizationCodeID: nil,
			expectError:         false,
		},
		{
			name:              "empty token value",
			tokenValue:        "",
			tokenType:         "Bearer",
			expiresAt:         time.Now().Add(1 * time.Hour),
			scopes:            []string{"openid"},
			subjectIdentifier: "user-123",
			expectError:       true,
		},
		{
			name:              "empty token type",
			tokenValue:        "access-token-123",
			tokenType:         "",
			expiresAt:         time.Now().Add(1 * time.Hour),
			scopes:            []string{"openid"},
			subjectIdentifier: "user-123",
			expectError:       true,
		},
		{
			name:              "empty subject identifier",
			tokenValue:        "access-token-123",
			tokenType:         "Bearer",
			expiresAt:         time.Now().Add(1 * time.Hour),
			scopes:            []string{"openid"},
			subjectIdentifier: "",
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewTerraformAccessToken(
				tt.tokenValue,
				tt.tokenType,
				tt.expiresAt,
				tt.scopes,
				tt.subjectIdentifier,
				tt.authorizationCodeID,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, token)
				assert.Equal(t, shared.ErrInvalidInput, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.NotEmpty(t, token.ID())
				assert.Equal(t, tt.tokenValue, token.TokenValue())
				assert.Equal(t, tt.tokenType, token.TokenType())
				assert.Equal(t, tt.expiresAt, token.ExpiresAt())
				assert.Equal(t, tt.scopes, token.Scopes())
				assert.Equal(t, tt.subjectIdentifier, token.SubjectIdentifier())
				assert.Equal(t, tt.authorizationCodeID, token.AuthorizationCodeID())
				assert.False(t, token.CreatedAt().IsZero())
			}
		})
	}
}

// TestTerraformAccessToken_IsExpired tests the IsExpired method
func TestTerraformAccessToken_IsExpired(t *testing.T) {
	t.Run("non-expired token", func(t *testing.T) {
		expiresAt := time.Now().Add(1 * time.Hour)
		token, err := NewTerraformAccessToken("token-value", "Bearer", expiresAt, []string{"openid"}, "user-123", nil)
		require.NoError(t, err)

		assert.False(t, token.IsExpired())
		assert.True(t, token.IsValid())
	})

	t.Run("expired token", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)
		token, err := NewTerraformAccessToken("token-value", "Bearer", expiresAt, []string{"openid"}, "user-123", nil)
		require.NoError(t, err)

		assert.True(t, token.IsExpired())
		assert.False(t, token.IsValid())
	})

	t.Run("token expiring now", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Second)
		token, err := NewTerraformAccessToken("token-value", "Bearer", expiresAt, []string{"openid"}, "user-123", nil)
		require.NoError(t, err)

		// Token should be considered expired
		assert.True(t, token.IsExpired())
	})
}

// TestTerraformAccessToken_HasScope tests the HasScope method
func TestTerraformAccessToken_HasScope(t *testing.T) {
	scopes := []string{"openid", "profile", "email"}
	token, err := NewTerraformAccessToken("token-value", "Bearer", time.Now().Add(1*time.Hour), scopes, "user-123", nil)
	require.NoError(t, err)

	assert.True(t, token.HasScope("openid"))
	assert.True(t, token.HasScope("profile"))
	assert.True(t, token.HasScope("email"))
	assert.False(t, token.HasScope("address"))
	assert.False(t, token.HasScope("phone"))
	assert.False(t, token.HasScope(""))
}

// TestNewTerraformSubjectIdentifier tests creating a new TerraformSubjectIdentifier
func TestNewTerraformSubjectIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		subject     string
		issuer      string
		authMethod  string
		expectError bool
	}{
		{
			name:        "valid subject identifier",
			subject:     "user-123",
			issuer:      "https://example.com",
			authMethod:  "oidc",
			expectError: false,
		},
		{
			name:        "valid with complex subject",
			subject:     "auth0|user-abc-123",
			issuer:      "https://auth0.com",
			authMethod:  "oauth2",
			expectError: false,
		},
		{
			name:        "empty subject",
			subject:     "",
			issuer:      "https://example.com",
			authMethod:  "oidc",
			expectError: true,
		},
		{
			name:        "empty issuer",
			subject:     "user-123",
			issuer:      "",
			authMethod:  "oidc",
			expectError: true,
		},
		{
			name:        "empty auth method",
			subject:     "user-123",
			issuer:      "https://example.com",
			authMethod:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subjectID, err := NewTerraformSubjectIdentifier(tt.subject, tt.issuer, tt.authMethod)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, subjectID)
				assert.Equal(t, shared.ErrInvalidInput, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, subjectID)
				assert.NotEmpty(t, subjectID.ID())
				assert.Equal(t, tt.subject, subjectID.Subject())
				assert.Equal(t, tt.issuer, subjectID.Issuer())
				assert.Equal(t, tt.authMethod, subjectID.AuthMethod())
				assert.False(t, subjectID.CreatedAt().IsZero())
				assert.False(t, subjectID.LastSeenAt().IsZero())
				assert.NotNil(t, subjectID.Metadata())
			}
		})
	}
}

// TestTerraformSubjectIdentifier_UpdateLastSeen tests updating last seen timestamp
func TestTerraformSubjectIdentifier_UpdateLastSeen(t *testing.T) {
	subjectID, err := NewTerraformSubjectIdentifier("user-123", "https://example.com", "oidc")
	require.NoError(t, err)

	originalLastSeen := subjectID.LastSeenAt()

	// Wait a tiny bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	subjectID.UpdateLastSeen()

	assert.True(t, subjectID.LastSeenAt().After(originalLastSeen))
}

// TestTerraformSubjectIdentifier_Metadata tests metadata operations
func TestTerraformSubjectIdentifier_Metadata(t *testing.T) {
	subjectID, err := NewTerraformSubjectIdentifier("user-123", "https://example.com", "oidc")
	require.NoError(t, err)

	t.Run("set and get metadata", func(t *testing.T) {
		subjectID.SetMetadata("key1", "value1")
		value, exists := subjectID.GetMetadata("key1")

		assert.True(t, exists)
		assert.Equal(t, "value1", value)
	})

	t.Run("get non-existent metadata", func(t *testing.T) {
		value, exists := subjectID.GetMetadata("nonexistent")

		assert.False(t, exists)
		assert.Nil(t, value)
	})

	t.Run("update existing metadata", func(t *testing.T) {
		subjectID.SetMetadata("key2", "original")
		subjectID.SetMetadata("key2", "updated")

		value, exists := subjectID.GetMetadata("key2")
		assert.True(t, exists)
		assert.Equal(t, "updated", value)
	})

	t.Run("set different types", func(t *testing.T) {
		subjectID.SetMetadata("string", "string-value")
		subjectID.SetMetadata("int", 123)
		subjectID.SetMetadata("bool", true)
		subjectID.SetMetadata("slice", []string{"a", "b", "c"})

		assert.Equal(t, "string-value", subjectID.Metadata()["string"])
		assert.Equal(t, 123, subjectID.Metadata()["int"])
		assert.Equal(t, true, subjectID.Metadata()["bool"])
		assert.Equal(t, []string{"a", "b", "c"}, subjectID.Metadata()["slice"])
	})
}

// TestReconstructTerraformSession tests reconstructing a session from persistence
func TestReconstructTerraformSession(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(10 * time.Minute)
	createdAt := now.Add(-1 * time.Minute)
	exchangedAt := now.Add(-30 * time.Second)
	accessTokenID := "token-123"

	session := ReconstructTerraformSession(
		"session-id",
		"auth-code",
		"client-id",
		"http://localhost",
		"openid",
		"state-123",
		strPtr("challenge"),
		strPtr("S256"),
		strPtr("nonce"),
		expiresAt,
		createdAt,
		&exchangedAt,
		&accessTokenID,
		"user-123",
	)

	assert.Equal(t, "session-id", session.ID())
	assert.Equal(t, "auth-code", session.AuthorizationCode())
	assert.Equal(t, "client-id", session.ClientID())
	assert.Equal(t, "http://localhost", session.RedirectURI())
	assert.Equal(t, "openid", session.Scope())
	assert.Equal(t, "state-123", session.State())
	assert.Equal(t, "user-123", session.SubjectIdentifier())
	assert.Equal(t, &exchangedAt, session.ExchangedAt())
	assert.Equal(t, &accessTokenID, session.AccessTokenID())
}

// TestReconstructTerraformAccessToken tests reconstructing a token from persistence
func TestReconstructTerraformAccessToken(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)
	createdAt := now.Add(-30 * time.Minute)
	authorizationCodeID := "auth-code-123"

	token := ReconstructTerraformAccessToken(
		"token-id",
		"token-value",
		"Bearer",
		expiresAt,
		createdAt,
		[]string{"openid", "profile"},
		"user-123",
		&authorizationCodeID,
	)

	assert.Equal(t, "token-id", token.ID())
	assert.Equal(t, "token-value", token.TokenValue())
	assert.Equal(t, "Bearer", token.TokenType())
	assert.Equal(t, expiresAt, token.ExpiresAt())
	assert.Equal(t, createdAt, token.CreatedAt())
	assert.Equal(t, []string{"openid", "profile"}, token.Scopes())
	assert.Equal(t, "user-123", token.SubjectIdentifier())
	assert.Equal(t, &authorizationCodeID, token.AuthorizationCodeID())
}

// TestReconstructTerraformSubjectIdentifier tests reconstructing from persistence
func TestReconstructTerraformSubjectIdentifier(t *testing.T) {
	now := time.Now()
	createdAt := now.Add(-1 * time.Hour)
	lastSeenAt := now.Add(-30 * time.Minute)
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	subjectID := ReconstructTerraformSubjectIdentifier(
		"subject-id",
		"user-123",
		"https://example.com",
		"oidc",
		createdAt,
		lastSeenAt,
		metadata,
	)

	assert.Equal(t, "subject-id", subjectID.ID())
	assert.Equal(t, "user-123", subjectID.Subject())
	assert.Equal(t, "https://example.com", subjectID.Issuer())
	assert.Equal(t, "oidc", subjectID.AuthMethod())
	assert.Equal(t, createdAt, subjectID.CreatedAt())
	assert.Equal(t, lastSeenAt, subjectID.LastSeenAt())
	assert.Equal(t, "value1", subjectID.Metadata()["key1"])
	assert.Equal(t, 123, subjectID.Metadata()["key2"])
}

// TestReconstructTerraformSubjectIdentifier_NilMetadata tests reconstruction with nil metadata
func TestReconstructTerraformSubjectIdentifier_NilMetadata(t *testing.T) {
	now := time.Now()

	subjectID := ReconstructTerraformSubjectIdentifier(
		"subject-id",
		"user-123",
		"https://example.com",
		"oidc",
		now,
		now,
		nil,
	)

	// Metadata should be initialized to empty map
	assert.NotNil(t, subjectID.Metadata())
	assert.Empty(t, subjectID.Metadata())
}

// Helper function
func strPtr(s string) *string {
	return &s
}
