package service

import (
	"context"
	"fmt"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
)

// SessionDatabaseConfig contains session database configuration
type SessionDatabaseConfig struct {
	DefaultTTL      time.Duration `json:"default_ttl"`
	MaxTTL          time.Duration `json:"max_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// SessionService handles session management - pure database operations
type SessionService struct {
	sessionRepo repository.SessionRepository
	config      *SessionDatabaseConfig
}

// NewSessionService creates a new session service
func NewSessionService(sessionRepo repository.SessionRepository, config *SessionDatabaseConfig) *SessionService {
	if config == nil {
		config = DefaultSessionDatabaseConfig()
	}

	return &SessionService{
		sessionRepo: sessionRepo,
		config:      config,
	}
}

// DefaultSessionDatabaseConfig returns default session database configuration
func DefaultSessionDatabaseConfig() *SessionDatabaseConfig {
	return &SessionDatabaseConfig{
		DefaultTTL:      1 * time.Hour,
		MaxTTL:          24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
	}
}

// CreateSession creates a new session with provider-specific auth data
func (ss *SessionService) CreateSession(ctx context.Context, authMethod string, providerData []byte, ttl *time.Duration) (*auth.Session, error) {
	sessionID := ss.generateSessionID()
	expiry := time.Now().Add(ss.getEffectiveTTL(ttl))

	session := &auth.Session{
		ID:                 sessionID,
		Expiry:             expiry,
		ProviderSourceAuth: providerData,
		CreatedAt:          &now,
		LastAccessedAt:     &now,
	}

	if err := ss.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// ValidateSession validates a session by checking if it exists and hasn't expired
func (ss *SessionService) ValidateSession(ctx context.Context, sessionID string) (*auth.Session, error) {
	session, err := ss.sessionRepo.FindByID(ctx, sessionID)
	if err != nil || session == nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.IsExpired() {
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// GetSession retrieves a session by ID without checking expiry
func (ss *SessionService) GetSession(ctx context.Context, sessionID string) (*auth.Session, error) {
	return ss.sessionRepo.FindByID(ctx, sessionID)
}

// DeleteSession deletes a session by ID
func (ss *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	return ss.sessionRepo.Delete(ctx, sessionID)
}

// CleanupExpiredSessions removes all expired sessions from the database
func (ss *SessionService) CleanupExpiredSessions(ctx context.Context) error {
	return ss.sessionRepo.CleanupExpired(ctx)
}

// RefreshSession updates the expiry time of a session
func (ss *SessionService) RefreshSession(ctx context.Context, sessionID string, ttl time.Duration) (*auth.Session, error) {
	session, err := ss.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	newExpiry := time.Now().Add(ss.getEffectiveTTL(&ttl))
	session.Expiry = newExpiry
	session.LastAccessedAt = &now

	if err := ss.sessionRepo.UpdateProviderSourceAuth(ctx, sessionID, session.ProviderSourceAuth); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

// generateSessionID generates a unique session ID
func (ss *SessionService) generateSessionID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getEffectiveTTL returns the effective TTL, respecting min/max limits
func (ss *SessionService) getEffectiveTTL(ttl *time.Duration) time.Duration {
	if ttl == nil {
		return ss.config.DefaultTTL
	}

	effectiveTTL := *ttl
	if effectiveTTL > ss.config.MaxTTL {
		effectiveTTL = ss.config.MaxTTL
	}

	return effectiveTTL
}

var now = time.Now()
