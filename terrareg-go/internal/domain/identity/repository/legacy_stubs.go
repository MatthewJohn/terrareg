package repository

import (
	"context"
	"errors"
	"time"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
)

// Temporary stub repository interfaces for legacy user-based services
// These should be removed as part of the migration to group-based auth

// UserRepository represents user repository (legacy)
// TODO: Remove this as part of migration to group-based authentication
type UserRepository interface {
	FindByID(ctx context.Context, id int) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByExternalID(ctx context.Context, externalID string) (*model.User, error)
	FindByAccessToken(ctx context.Context, token string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Save(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*model.User, error)
	Count(ctx context.Context) (int, error)
}

// UserGroupRepository represents user group repository (legacy)
// TODO: Remove this as part of migration to auth domain user groups
type UserGroupRepository interface {
	FindByID(ctx context.Context, id int) (*model.UserGroup, error)
	FindByName(ctx context.Context, name string) (*model.UserGroup, error)
	Create(ctx context.Context, userGroup *model.UserGroup) error
	Update(ctx context.Context, userGroup *model.UserGroup) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*model.UserGroup, error)
	Count(ctx context.Context) (int, error)
}

// SessionRepository alias for legacy compatibility
// TODO: Remove this as part of migration to auth domain sessions
type SessionRepository = UserSessionRepository

// UserSessionRepository represents user session repository (legacy)
// TODO: Remove this as part of migration to auth domain sessions
type UserSessionRepository interface {
	Save(ctx context.Context, session *Session) error
	FindByID(ctx context.Context, id string) (*Session, error)
	FindByToken(ctx context.Context, token string) (*Session, error)
	Delete(ctx context.Context, id string) error
	DeleteByToken(ctx context.Context, token string) error
	FindByUserID(ctx context.Context, userID string) ([]*Session, error)
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) (int, error)
	List(ctx context.Context, offset, limit int) ([]*Session, error)
	Count(ctx context.Context) (int, error)
	CountByUserID(ctx context.Context, userID string) (int, error)
	WithTransaction(tx interface{}) UserSessionRepository
}

// Legacy Session model for compatibility
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// IdentityRepository interface for legacy compatibility
type IdentityRepository interface {
	Save(ctx context.Context, identity interface{}) error
	FindByID(ctx context.Context, id string) (interface{}, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, identity interface{}) error
}

// Add ErrNotFound constant at package level
var ErrNotFound = errors.New("resource not found")

