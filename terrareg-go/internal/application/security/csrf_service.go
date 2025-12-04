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
	tokenGenerator csrf.TokenGenerator
	tokenValidator csrf.TokenValidator
	sessionManager SessionManager
}

// NewCSRFService creates a new CSRF service
func NewCSRFService(
	tokenGenerator csrf.TokenGenerator,
	tokenValidator csrf.TokenValidator,
	sessionManager SessionManager,
) *CSRFService {
	return &CSRFService{
		tokenGenerator: tokenGenerator,
		tokenValidator: tokenValidator,
		sessionManager: sessionManager,
	}
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