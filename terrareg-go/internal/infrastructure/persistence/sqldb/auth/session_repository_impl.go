package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

const SessionIDLength = 32

// SessionRepositoryImpl implements the session repository using GORM
type SessionRepositoryImpl struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) *SessionRepositoryImpl {
	return &SessionRepositoryImpl{db: db}
}

// Create creates a new session
func (r *SessionRepositoryImpl) Create(ctx context.Context, session *auth.Session) error {
	sessionDB := sqldb.SessionDB{
		ID:                  session.ID(),
		Expiry:              session.Expiry(),
		ProviderSourceAuth: session.ProviderSourceAuth(),
	}

	return r.db.WithContext(ctx).Create(&sessionDB).Error
}

// FindByID retrieves a session by ID if it hasn't expired
func (r *SessionRepositoryImpl) FindByID(ctx context.Context, sessionID string) (*auth.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is empty")
	}

	var sessionDB sqldb.SessionDB
	err := r.db.WithContext(ctx).
		Where("id = ? AND expiry >= ?", sessionID, time.Now()).
		First(&sessionDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	session := auth.NewSession(sessionDB.ID, sessionDB.Expiry)
	session.SetProviderSourceAuth(sessionDB.ProviderSourceAuth)
	return session, nil
}

// Delete deletes a session
func (r *SessionRepositoryImpl) Delete(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", sessionID).
		Delete(&sqldb.SessionDB{}).Error
}

// CleanupExpired removes all expired sessions
func (r *SessionRepositoryImpl) CleanupExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expiry < ?", time.Now()).
		Delete(&sqldb.SessionDB{}).Error
}

// UpdateProviderSourceAuth updates provider source auth data for a session
func (r *SessionRepositoryImpl) UpdateProviderSourceAuth(ctx context.Context, sessionID string, data []byte) error {
	return r.db.WithContext(ctx).
		Model(&sqldb.SessionDB{}).
		Where("id = ?", sessionID).
		Update("provider_source_auth", data).Error
}

// GenerateSessionID generates a random session ID (matching Python's secrets.token_urlsafe)
func GenerateSessionID() (string, error) {
	bytes := make([]byte, SessionIDLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:SessionIDLength], nil
}
