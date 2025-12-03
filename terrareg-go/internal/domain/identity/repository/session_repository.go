package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
)

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	// Basic CRUD operations
	Save(ctx context.Context, session *model.Session) error
	FindByID(ctx context.Context, id string) (*model.Session, error)
	FindByToken(ctx context.Context, token string) (*model.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteByToken(ctx context.Context, token string) error

	// User session operations
	FindByUserID(ctx context.Context, userID string) ([]*model.Session, error)
	DeleteByUserID(ctx context.Context, userID string) error

	// Cleanup operations
	DeleteExpired(ctx context.Context) (int, error)

	// Query operations
	List(ctx context.Context, offset, limit int) ([]*model.Session, error)
	Count(ctx context.Context) (int, error)
	CountByUserID(ctx context.Context, userID string) (int, error)

	// Transaction support
	WithTransaction(tx interface{}) SessionRepository
}

// SessionMetadata represents additional session metadata
type SessionMetadata struct {
	IPAddress string
	UserAgent string
	Remember  bool
}