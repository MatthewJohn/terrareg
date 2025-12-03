package repository

import (
	"context"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Basic CRUD operations
	Save(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error

	// Auth-specific operations
	FindByAuthProviderID(ctx context.Context, authMethod model.AuthMethod, providerID string) (*model.User, error)
	FindByExternalID(ctx context.Context, authMethod model.AuthMethod, externalID string) (*model.User, error)
	FindByAccessToken(ctx context.Context, accessToken string) (*model.User, error)

	// User management
	FindActive(ctx context.Context) ([]*model.User, error)
	FindInactive(ctx context.Context) ([]*model.User, error)
	FindByUserGroupID(ctx context.Context, userGroupID string) ([]*model.User, error)

	// Query operations
	List(ctx context.Context, offset, limit int) ([]*model.User, error)
	Count(ctx context.Context) (int, error)
	Search(ctx context.Context, query string, offset, limit int) ([]*model.User, error)

	// Transaction support
	WithTransaction(tx interface{}) UserRepository
}