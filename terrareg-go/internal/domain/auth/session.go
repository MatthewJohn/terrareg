package auth

import (
	"time"
)

// Session represents a user session (matching Python Session model)
type Session struct {
	ID                 string      `json:"id"`
	Expiry             time.Time   `json:"expiry"`
	ProviderSourceAuth []byte      `json:"provider_source_auth"`
	CreatedAt          *time.Time  `json:"created_at,omitempty"`
	LastAccessedAt     *time.Time  `json:"last_accessed_at,omitempty"`
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
