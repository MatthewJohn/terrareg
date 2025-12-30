package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateUsername tests username validation
func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		expectErr error
	}{
		{"valid username", "john_doe", nil},
		{"valid with hyphen", "john-doe", nil},
		{"valid with numbers", "user123", nil},
		{"valid mixed", "User_123-456", nil},
		{"empty username", "", ErrUsernameRequired},
		{"username too long", string(make([]byte, 51)), ErrUsernameTooLong},
		{"username with spaces", "john doe", ErrUsernameInvalid},
		{"username with special chars", "john@doe", ErrUsernameInvalid},
		{"username with dots", "john.doe", ErrUsernameInvalid},
		{"username only numbers", "123456", nil},
		{"single char", "a", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if tt.expectErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			}
		})
	}
}

// TestValidateEmail tests email validation
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expectErr error
	}{
		{"valid email", "user@example.com", nil},
		{"valid with subdomain", "user@mail.example.com", nil},
		{"valid with dots", "first.last@example.com", nil},
		{"valid with plus", "user+tag@example.com", nil},
		{"valid with numbers", "user123@example123.com", nil},
		{"empty email", "", ErrEmailRequired},
		{"missing @", "userexample.com", ErrEmailInvalid},
		{"missing domain", "user@", ErrEmailInvalid},
		{"missing user", "@example.com", ErrEmailInvalid},
		{"invalid format", "user@.com", ErrEmailInvalid},
		{"too long", string(make([]byte, 256)) + "@example.com", ErrEmailInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.expectErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			}
		})
	}
}

// TestValidateDisplayName tests display name validation
func TestValidateDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		expectErr   error
	}{
		{"valid display name", "John Doe", nil},
		{"valid with emoji", "John Doe 👋", nil},
		{"valid with special chars", "John (Developer)", nil},
		{"valid unicode", "日本語", nil},
		{"empty display name", "", ErrDisplayNameRequired},
		{"too long", string(make([]byte, 101)), ErrDisplayNameTooLong},
		{"single character", "a", nil},
		{"with spaces only", "   ", nil}, // spaces are valid characters
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDisplayName(tt.displayName)
			if tt.expectErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			}
		})
	}
}

// TestValidateGroupName tests group name validation
func TestValidateGroupName(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		expectErr error
	}{
		{"valid group name", "developers", nil},
		{"valid with underscore", "dev_team", nil},
		{"valid with hyphen", "dev-team", nil},
		{"valid with numbers", "team123", nil},
		{"valid mixed", "Dev_Team-123", nil},
		{"empty group name", "", ErrGroupNameRequired},
		{"group name too long", string(make([]byte, 51)), ErrGroupNameTooLong},
		{"group name with spaces", "dev team", ErrGroupNameInvalid},
		{"group name with special chars", "dev@team", ErrGroupNameInvalid},
		{"group name with dots", "dev.team", ErrGroupNameInvalid},
		{"single char", "a", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGroupName(tt.groupName)
			if tt.expectErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			}
		})
	}
}

// TestValidateAction tests action validation
func TestValidateAction(t *testing.T) {
	tests := []struct {
		name      string
		action    string
		expectErr bool
	}{
		{"READ action", "READ", false},
		{"MODIFY action", "MODIFY", false},
		{"FULL action", "FULL", false},
		{"invalid action", "DELETE", true},
		{"empty action", "", true},
		{"lowercase read", "read", true},
		{"mixed case", "Read", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAction(tt.action)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateResourceType tests resource type validation
func TestValidateResourceType(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expectErr    bool
	}{
		{"namespace resource", "namespace", false},
		{"module resource", "module", false},
		{"provider resource", "provider", false},
		{"invalid resource", "user", true},
		{"empty resource", "", true},
		{"capitalized", "Namespace", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceType(tt.resourceType)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateAuthMethod tests auth method validation
func TestValidateAuthMethod(t *testing.T) {
	validMethods := []string{
		"NOT_AUTHENTICATED",
		"SAML",
		"OPENID_CONNECT",
		"GITHUB",
		"ADMIN_API_KEY",
		"TERRAFORM_OIDC",
	}

	for _, method := range validMethods {
		t.Run("valid_"+method, func(t *testing.T) {
			err := ValidateAuthMethod(method)
			assert.NoError(t, err)
		})
	}

	t.Run("invalid method", func(t *testing.T) {
		err := ValidateAuthMethod("OAUTH2")
		assert.Error(t, err)
	})

	t.Run("empty method", func(t *testing.T) {
		err := ValidateAuthMethod("")
		assert.Error(t, err)
	})

	t.Run("lowercase", func(t *testing.T) {
		err := ValidateAuthMethod("github")
		assert.Error(t, err)
	})
}

// TestSanitizeInput tests input sanitization
func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no spaces", "test", "test"},
		{"leading spaces", "  test", "test"},
		{"trailing spaces", "test  ", "test"},
		{"both sides", "  test  ", "test"},
		{"tabs only", "\t\t", ""},
		{"spaces only", "   ", ""},
		{"mixed whitespace", "  \t test \t  ", "test"},
		{"internal spaces preserved", "test value", "test value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateScopes tests scope validation
func TestValidateScopes(t *testing.T) {
	tests := []struct {
		name      string
		scopes    []string
		expectErr error
	}{
		{"valid scopes", []string{"openid", "profile", "email"}, nil},
		{"single scope", []string{"openid"}, nil},
		{"empty scopes", []string{}, ErrScopesRequired},
		{"nil scopes", nil, ErrScopesRequired},
		{"scope with empty string", []string{"openid", ""}, assert.AnError},
		{"multiple empty strings", []string{"", ""}, assert.AnError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScopes(tt.scopes)
			if tt.expectErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
