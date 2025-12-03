package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

// UserGroupRepository defines the interface for user group data access
type UserGroupRepository interface {
	// Basic CRUD operations
	Save(ctx context.Context, userGroup *auth.UserGroup) error
	FindByID(ctx context.Context, id int) (*auth.UserGroup, error)
	FindByName(ctx context.Context, name string) (*auth.UserGroup, error)
	Update(ctx context.Context, userGroup *auth.UserGroup) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, offset, limit int) ([]*auth.UserGroup, error)
	Count(ctx context.Context) (int64, error)

	// Permission operations
	AddNamespacePermission(ctx context.Context, userGroupID, namespaceID int, permissionType auth.PermissionType) error
	RemoveNamespacePermission(ctx context.Context, userGroupID, namespaceID int) error
	HasNamespacePermission(ctx context.Context, userGroupID, namespaceID int, permissionType auth.PermissionType) (bool, error)
	GetNamespacePermissions(ctx context.Context, userGroupID int) ([]auth.NamespacePermission, error)
	GetHighestNamespacePermission(ctx context.Context, userGroupID, namespaceID int) (auth.PermissionType, error)

	// Query operations
	FindSiteAdminGroups(ctx context.Context) ([]*auth.UserGroup, error)
	FindGroupsByNamespace(ctx context.Context, namespaceID int) ([]*auth.UserGroup, error)
	SearchByName(ctx context.Context, query string, offset, limit int) ([]*auth.UserGroup, error)
}

// UserGroupRepositoryTx extends UserGroupRepository with transaction support
type UserGroupRepositoryTx interface {
	UserGroupRepository
	WithTransaction(tx interface{}) UserGroupRepository
}