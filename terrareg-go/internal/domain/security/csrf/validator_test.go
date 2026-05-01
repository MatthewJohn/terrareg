package csrf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecureTokenValidator(t *testing.T) {
	t.Run("creates non-nil validator", func(t *testing.T) {
		validator := NewSecureTokenValidator()
		assert.NotNil(t, validator)
	})
}

func TestSecureTokenValidator_ValidateToken(t *testing.T) {
	tests := []struct {
		name          string
		expectedToken CSRFToken
		providedToken CSRFToken
		required      bool
		wantErr       error
	}{
		// Success cases
		{
			name:          "valid token",
			expectedToken: "abc123def456",
			providedToken: "abc123def456",
			required:      true,
			wantErr:       nil,
		},
		{
			name:          "valid token when not required",
			expectedToken: "abc123def456",
			providedToken: "",
			required:      false,
			wantErr:       nil,
		},
		{
			name:          "both empty when not required",
			expectedToken: "",
			providedToken: "",
			required:      false,
			wantErr:       nil,
		},

		// Error cases
		{
			name:          "missing token",
			expectedToken: "abc123def456",
			providedToken: "",
			required:      true,
			wantErr:       ErrMissingToken,
		},
		{
			name:          "invalid token",
			expectedToken: "abc123def456",
			providedToken: "xyz789uvw123",
			required:      true,
			wantErr:       ErrInvalidToken,
		},
		{
			name:          "no session",
			expectedToken: "",
			providedToken: "abc123",
			required:      true,
			wantErr:       ErrNoSession,
		},
		{
			name:          "no session when required",
			expectedToken: "",
			providedToken: "",
			required:      true,
			wantErr:       ErrNoSession,
		},
		{
			name:          "case-sensitive comparison",
			expectedToken: "ABC123",
			providedToken: "abc123",
			required:      true,
			wantErr:       ErrInvalidToken,
		},
		{
			name:          "whitespace differences",
			expectedToken: "abc 123",
			providedToken: "abc123",
			required:      true,
			wantErr:       ErrInvalidToken,
		},
		{
			name:          "special characters in token",
			expectedToken: "abc123!@#",
			providedToken: "abc123!@#",
			required:      true,
			wantErr:       nil,
		},
	}

	validator := NewSecureTokenValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateToken(tt.expectedToken, tt.providedToken, tt.required)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr,
					"Expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
