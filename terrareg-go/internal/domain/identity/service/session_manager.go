package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
)

// Using identity model errors instead of redefining here

// SessionInfo represents simplified session information for identity management
type SessionInfo struct {
	ID         string
	UserID     string
	AuthMethod  model.AuthMethod
	ExpiresAt   time.Time
	CreatedAt   time.Time
	IPAddress   string
	UserAgent   string
}

// SessionManager manages user sessions
type SessionManager struct {
	userRepo   repository.UserRepository
	config     SessionConfig
	sessions   map[string]*SessionInfo // In-memory session store for Phase 4
}

// SessionConfig holds session configuration
type SessionConfig struct {
	DefaultTTL       time.Duration
	MaxTTL           time.Duration
	CleanupInterval   time.Duration
	RequireReauth     bool
	SessionCookieName string
	Domain           string
	Secure           bool
	HTTPOnly         bool
	SameSite         string
}

// NewSessionManager creates a new session manager
func NewSessionManager(userRepo repository.UserRepository, config SessionConfig) *SessionManager {
	return &SessionManager{
		userRepo: userRepo,
		config:   config,
		sessions: make(map[string]*SessionInfo),
	}
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(ctx context.Context, userID string, authMethod model.AuthMethod, metadata SessionMetadata) (*SessionInfo, error) {
	// Verify user exists and is active
	user, err := sm.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, model.ErrUserNotFound
	}
	if !user.Active() {
		return nil, model.ErrUserInactive
	}

	// Generate session token
	sessionToken, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	// Determine TTL based on auth method
	ttl := sm.config.DefaultTTL
	if authMethod == model.AuthMethodAPIKey {
		ttl = sm.config.MaxTTL // API keys can have longer sessions
	}

	// Create session
	now := time.Now()
	session := &SessionInfo{
		ID:        sessionToken,
		UserID:    userID,
		AuthMethod: authMethod,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
		IPAddress: metadata.IPAddress,
		UserAgent: metadata.UserAgent,
	}

	// Store session in memory for Phase 4
	// In a full implementation, this would be stored in the database
	sm.sessions[sessionToken] = session

	return session, nil
}

// ValidateSession validates a session token and returns the user
func (sm *SessionManager) ValidateSession(ctx context.Context, sessionToken string) (*SessionInfo, *model.User, error) {
	if sessionToken == "" {
		return nil, nil, model.ErrSessionInvalid
	}

	// Find session
	session, exists := sm.sessions[sessionToken]
	if !exists {
		return nil, nil, model.ErrSessionNotFound
	}

	// Check if session is valid
	if time.Now().After(session.ExpiresAt) {
		// Remove expired session
		delete(sm.sessions, sessionToken)
		return nil, nil, model.ErrSessionExpired
	}

	// Get user
	user, err := sm.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	// Check if user is active
	if !user.Active() {
		return nil, nil, model.ErrUserInactive
	}

	return session, user, nil
}

// RefreshSession extends a session's expiration
func (sm *SessionManager) RefreshSession(ctx context.Context, sessionToken string) (*SessionInfo, error) {
	if sessionToken == "" {
		return nil, model.ErrSessionInvalid
	}

	// Find session
	session, exists := sm.sessions[sessionToken]
	if !exists {
		return nil, model.ErrSessionNotFound
	}

	// Check if session is still valid
	if time.Now().After(session.ExpiresAt) {
		delete(sm.sessions, sessionToken)
		return nil, model.ErrSessionExpired
	}

	// Extend session
	session.ExpiresAt = time.Now().Add(sm.config.DefaultTTL)

	return session, nil
}

// InvalidateSession invalidates a session
func (sm *SessionManager) InvalidateSession(ctx context.Context, sessionToken string) error {
	if sessionToken == "" {
		return model.ErrSessionInvalid
	}

	_, exists := sm.sessions[sessionToken]
	if !exists {
		return model.ErrSessionNotFound
	}

	delete(sm.sessions, sessionToken)
	return nil
}

// InvalidateAllUserSessions invalidates all sessions for a user
func (sm *SessionManager) InvalidateAllUserSessions(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrUserNotFound
	}

	// Remove all sessions for this user
	for token, session := range sm.sessions {
		if session.UserID == userID {
			delete(sm.sessions, token)
		}
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (sm *SessionManager) CleanupExpiredSessions(ctx context.Context) (int, error) {
	count := 0
	now := time.Now()

	for token, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, token)
			count++
		}
	}

	return count, nil
}

// GetActiveUserSessions returns all active sessions for a user
func (sm *SessionManager) GetActiveUserSessions(ctx context.Context, userID string) ([]*SessionInfo, error) {
	if userID == "" {
		return nil, ErrUserNotFound
	}

	var userSessions []*SessionInfo
	now := time.Now()

	for _, session := range sm.sessions {
		if session.UserID == userID && !now.After(session.ExpiresAt) {
			userSessions = append(userSessions, session)
		}
	}

	return userSessions, nil
}

// SessionMetadata holds session creation metadata
type SessionMetadata struct {
	IPAddress string
	UserAgent string
	Remember  bool
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}