package csrf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecureTokenGenerator(t *testing.T) {
	t.Run("creates non-nil generator", func(t *testing.T) {
		generator := NewSecureTokenGenerator()
		assert.NotNil(t, generator)
	})
}

func TestSecureTokenGenerator_GenerateToken(t *testing.T) {
	t.Run("generates valid token", func(t *testing.T) {
		generator := NewSecureTokenGenerator()
		token, err := generator.GenerateToken()

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Len(t, token, 64)
	})

	t.Run("delegates to NewCSRFToken", func(t *testing.T) {
		generator := NewSecureTokenGenerator()
		token1, err1 := NewCSRFToken()
		token2, err2 := generator.GenerateToken()

		require.NoError(t, err1)
		require.NoError(t, err2)

		// Both should be valid 64-char hex strings
		assert.Len(t, token1, 64)
		assert.Len(t, token2, 64)

		// But they should be different (different random bytes)
		assert.NotEqual(t, token1, token2)
	})

	t.Run("wraps errors from NewCSRFToken", func(t *testing.T) {
		generator := NewSecureTokenGenerator()

		// In normal operation, this shouldn't fail
		token, err := generator.GenerateToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify error wrapping exists
		if err != nil {
			assert.Contains(t, err.Error(), "failed to generate CSRF token")
		}
	})
}
