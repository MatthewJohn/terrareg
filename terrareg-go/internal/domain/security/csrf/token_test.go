package csrf

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCSRFToken(t *testing.T) {
	t.Run("generates valid token", func(t *testing.T) {
		token, err := NewCSRFToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("generates 64-character hex string", func(t *testing.T) {
		token, err := NewCSRFToken()
		require.NoError(t, err)
		assert.Len(t, token, 64, "Token should be 64 characters (SHA256 hex)")
	})

	t.Run("generates valid hex format", func(t *testing.T) {
		token, err := NewCSRFToken()
		require.NoError(t, err)

		// Should only contain lowercase hex characters
		hexRegex := regexp.MustCompile(`^[a-f0-9]{64}$`)
		assert.True(t, hexRegex.MatchString(token.String()),
			"Token should be lowercase hex: %s", token)
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		token1, err1 := NewCSRFToken()
		token2, err2 := NewCSRFToken()
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, token1, token2,
			"Each token should be unique")
	})
}

func TestCSRFToken_String(t *testing.T) {
	t.Run("converts to string correctly", func(t *testing.T) {
		token := CSRFToken("test-token")
		assert.Equal(t, "test-token", token.String())
	})

	t.Run("handles empty token", func(t *testing.T) {
		token := CSRFToken("")
		assert.Equal(t, "", token.String())
	})
}

func TestCSRFToken_IsEmpty(t *testing.T) {
	t.Run("returns true for empty token", func(t *testing.T) {
		token := CSRFToken("")
		assert.True(t, token.IsEmpty())
	})

	t.Run("returns false for non-empty token", func(t *testing.T) {
		token := CSRFToken("abc123")
		assert.False(t, token.IsEmpty())
	})

	t.Run("returns false for whitespace-only token", func(t *testing.T) {
		token := CSRFToken("   ")
		assert.False(t, token.IsEmpty())
	})
}

func TestCSRFToken_Equals(t *testing.T) {
	t.Run("returns true for equal tokens", func(t *testing.T) {
		token1 := CSRFToken("abc123")
		token2 := CSRFToken("abc123")
		assert.True(t, token1.Equals(token2))
	})

	t.Run("returns false for different tokens", func(t *testing.T) {
		token1 := CSRFToken("abc123")
		token2 := CSRFToken("xyz789")
		assert.False(t, token1.Equals(token2))
	})

	t.Run("returns false for empty vs non-empty", func(t *testing.T) {
		token1 := CSRFToken("")
		token2 := CSRFToken("abc123")
		assert.False(t, token1.Equals(token2))
	})

	t.Run("is case-sensitive", func(t *testing.T) {
		token1 := CSRFToken("ABC123")
		token2 := CSRFToken("abc123")
		assert.False(t, token1.Equals(token2))
	})
}
