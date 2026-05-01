package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// SessionRepository defines the interface for session persistence
type SessionRepository interface {
	// Create creates a new session
	Create(ctx context.Context, session *auth.Session) error

	// FindByID retrieves a session by ID if it hasn't expired
	FindByID(ctx context.Context, sessionID string) (*auth.Session, error)

	// Delete deletes a session
	Delete(ctx context.Context, sessionID string) error

	// CleanupExpired removes all expired sessions
	CleanupExpired(ctx context.Context) error

	// UpdateProviderSourceAuth updates provider source auth data for a session
	UpdateProviderSourceAuth(ctx context.Context, sessionID string, data []byte) error
}
