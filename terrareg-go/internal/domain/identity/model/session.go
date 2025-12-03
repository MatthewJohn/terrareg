package model

import (
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// Session represents a user authentication session
type Session struct {
	id         string
	userID     string
	authMethod  AuthMethod
	expiresAt   time.Time
	createdAt   time.Time
	accessToken string
	ipAddress   string
	userAgent   string
}

// NewSession creates a new session
func NewSession(userID string, authMethod AuthMethod, accessToken string, ttl time.Duration) (*Session, error) {
	if userID == "" {
		return nil, ErrUserIDRequired
	}
	if accessToken == "" {
		return nil, ErrAccessTokenRequired
	}
	if ttl <= 0 {
		return nil, ErrInvalidInput
	}

	now := time.Now()
	return &Session{
		id:         shared.GenerateID(),
		userID:     userID,
		authMethod:  authMethod,
		accessToken: accessToken,
		createdAt:   now,
		expiresAt:   now.Add(ttl),
	}, nil
}

// Getters
func (s *Session) ID() string              { return s.id }
func (s *Session) UserID() string           { return s.userID }
func (s *Session) AuthMethod() AuthMethod    { return s.authMethod }
func (s *Session) AccessToken() string       { return s.accessToken }
func (s *Session) ExpiresAt() time.Time     { return s.expiresAt }
func (s *Session) CreatedAt() time.Time     { return s.createdAt }
func (s *Session) IPAddress() string        { return s.ipAddress }
func (s *Session) UserAgent() string        { return s.userAgent }

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}

// IsValid checks if the session is still valid
func (s *Session) IsValid() bool {
	return !s.IsExpired() && s.accessToken != ""
}

// Extend extends the session expiration
func (s *Session) Extend(ttl time.Duration) {
	s.expiresAt = time.Now().Add(ttl)
}

// SetMetadata sets session metadata
func (s *Session) SetMetadata(ipAddress, userAgent string) {
	s.ipAddress = ipAddress
	s.userAgent = userAgent
}