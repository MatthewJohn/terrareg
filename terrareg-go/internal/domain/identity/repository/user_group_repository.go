package repository

import (
	"context"
	"terrareg/internal/domain/identity/model"
)

// UserGroupRepository defines the interface for user group data access
type UserGroupRepository interface {
	// Basic CRUD operations
	Save(ctx context.Context, userGroup *model.UserGroup) error
	FindByID(ctx context.Context, id string) (*model.UserGroup, error)
	FindByName(ctx context.Context, name string) (*model.UserGroup, error)
	Update(ctx context.Context, userGroup *model.UserGroup) error
	Delete(ctx context.Context, id string) error

	// Membership operations
	AddMember(ctx context.Context, userGroupID, userID string) error
	RemoveMember(ctx context.Context, userGroupID, userID string) error
	IsMember(ctx context.Context, userGroupID, userID string) (bool, error)
	GetMembers(ctx context.Context, userGroupID string) ([]*model.User, error)
	GetActiveMembers(ctx context.Context, userGroupID string) ([]*model.User, error)

	// Permission operations
	AddPermission(ctx context.Context, userGroupID string, resourceType model.ResourceType, resourceID string, action model.Action, grantedBy string) error
	RemovePermission(ctx context.Context, userGroupID string, permissionID int) error
	HasPermission(ctx context.Context, userGroupID string, resourceType model.ResourceType, resourceID string, action model.Action) (bool, error)
	GetPermissions(ctx context.Context, userGroupID string) ([]model.GroupPermission, error)

	// Query operations
	List(ctx context.Context, offset, limit int) ([]*model.UserGroup, error)
	Count(ctx context.Context) (int, error)
	Search(ctx context.Context, query string, offset, limit int) ([]*model.UserGroup, error)

	// Find by user
	FindByUserID(ctx context.Context, userID string) ([]*model.UserGroup, error)

	// Transaction support
	WithTransaction(tx interface{}) UserGroupRepository
}