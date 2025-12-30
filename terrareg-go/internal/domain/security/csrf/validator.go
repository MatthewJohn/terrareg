package csrf

// TokenValidator interface for validating CSRF tokens
type TokenValidator interface {
	ValidateToken(expectedToken, providedToken CSRFToken, required bool) error
}

// SecureTokenValidator implements secure CSRF token validation
type SecureTokenValidator struct{}

// NewSecureTokenValidator creates a new secure token validator
func NewSecureTokenValidator() *SecureTokenValidator {
	return &SecureTokenValidator{}
}

// ValidateToken validates a provided CSRF token against the expected token
func (v *SecureTokenValidator) ValidateToken(expectedToken, providedToken CSRFToken, required bool) error {
	// If CSRF protection is not required for this request, skip validation
	if !required {
		return nil
	}

	// Check if the expected token is empty (no session)
	if expectedToken.IsEmpty() {
		return ErrNoSession
	}

	// Check if the provided token is empty
	if providedToken.IsEmpty() {
		return ErrMissingToken
	}

	// Check if tokens match
	if !expectedToken.Equals(providedToken) {
		return ErrInvalidToken
	}

	return nil
}
