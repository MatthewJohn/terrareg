package auth

import (
	"time"
)

// Session represents a user session (matching Python Session model)
type Session struct {
	id                 string
	expiry             time.Time
	providerSourceAuth []byte
}

// NewSession creates a new session
func NewSession(id string, expiry time.Time) *Session {
	return &Session{
		id:     id,
		expiry: expiry,
	}
}

// Getters
func (s *Session) ID() string              { return s.id }
func (s *Session) Expiry() time.Time       { return s.expiry }
func (s *Session) ProviderSourceAuth() []byte { return s.providerSourceAuth }

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.expiry)
}

// SetProviderSourceAuth updates provider source auth data
func (s *Session) SetProviderSourceAuth(data []byte) {
	s.providerSourceAuth = data
}
