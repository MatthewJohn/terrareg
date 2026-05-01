package auth

import (
	"context"
	"time"
)

// Session represents a user session (matching Python Session model)
type Session struct {
	ID                 string     `json:"id"`
	Expiry             time.Time  `json:"expiry"`
	ProviderSourceAuth []byte     `json:"provider_source_auth"`
	CreatedAt          *time.Time `json:"created_at,omitempty"`
	LastAccessedAt     *time.Time `json:"last_accessed_at,omitempty"`
}

// NewSession creates a new session
func NewSession(id string, expiry time.Time) *Session {
	return &Session{
		ID:     id,
		Expiry: expiry,
	}
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.Expiry)
}

// SetProviderSourceAuth updates provider source auth data
func (s *Session) SetProviderSourceAuth(data []byte) {
	s.ProviderSourceAuth = data
}

// SessionData represents the data stored in an encrypted session cookie
// This is the data that gets encrypted and stored in the HTTP cookie
type SessionData struct {
	SessionID    string            `json:"session_id"`
	UserID       string            `json:"user_id"`
	Username     string            `json:"username"`
	AuthMethod   string            `json:"auth_method"`
	IsAdmin      bool              `json:"is_admin"`
	SiteAdmin    bool              `json:"site_admin"`
	UserGroups   []string          `json:"user_groups"`
	Permissions  map[string]string `json:"permissions,omitempty"`
	Expiry       *time.Time        `json:"expiry"`
	LastAccessed *time.Time        `json:"last_accessed,omitempty"`
}

// SessionManager is an interface for session validation operations
// This interface allows infrastructure layer to use session management without import cycle
type SessionManager interface {
	// ValidateSessionCookie validates a session cookie by decrypting and checking the database
	// Returns the session if valid, nil if invalid/expired
	ValidateSessionCookie(ctx context.Context, cookieValue string) (*Session, error)

	// IsAvailable returns true if session management is available (SECRET_KEY configured)
	IsAvailable() bool
}
