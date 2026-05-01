package security

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/security/csrf"
)

// SessionManager interface for managing user sessions
type SessionManager interface {
	GetCSRFToken(ctx context.Context, sessionID string) (csrf.CSRFToken, error)
	CreateSession(ctx context.Context) (string, csrf.CSRFToken, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

// CSRFService handles CSRF token operations
type CSRFService struct {
	// tokenGenerator generates CSRF tokens (required)
	tokenGenerator csrf.TokenGenerator
	// tokenValidator validates CSRF tokens (required)
	tokenValidator csrf.TokenValidator
	// sessionManager manages user sessions (required)
	sessionManager SessionManager
}

// NewCSRFService creates a new CSRF service
// Returns an error if any required dependency is nil
func NewCSRFService(
	tokenGenerator csrf.TokenGenerator,
	tokenValidator csrf.TokenValidator,
	sessionManager SessionManager,
) (*CSRFService, error) {
	if tokenGenerator == nil {
		return nil, fmt.Errorf("tokenGenerator cannot be nil")
	}
	if tokenValidator == nil {
		return nil, fmt.Errorf("tokenValidator cannot be nil")
	}
	if sessionManager == nil {
		return nil, fmt.Errorf("sessionManager cannot be nil")
	}

	return &CSRFService{
		tokenGenerator: tokenGenerator,
		tokenValidator: tokenValidator,
		sessionManager: sessionManager,
	}, nil
}

// GetOrCreateSessionToken gets existing CSRF token from session or creates new session
func (s *CSRFService) GetOrCreateSessionToken(ctx context.Context, sessionID string) (csrf.CSRFToken, error) {
	if sessionID != "" {
		// Try to get existing token
		token, err := s.sessionManager.GetCSRFToken(ctx, sessionID)
		if err == nil && !token.IsEmpty() {
			return token, nil
		}
	}

	// Create new session with token
	newSessionID, token, err := s.sessionManager.CreateSession(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	// TODO: Return new session ID to be set in cookie
	_ = newSessionID
	return token, nil
}

// ValidateRequestToken validates a CSRF token from a request
func (s *CSRFService) ValidateRequestToken(ctx context.Context, sessionID string, providedToken csrf.CSRFToken, required bool) error {
	if !required {
		return nil
	}

	expectedToken, err := s.sessionManager.GetCSRFToken(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session CSRF token: %w", err)
	}

	return s.tokenValidator.ValidateToken(expectedToken, providedToken, required)
}
