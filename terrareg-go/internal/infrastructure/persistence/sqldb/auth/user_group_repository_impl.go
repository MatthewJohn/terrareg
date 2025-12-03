package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"gorm.io/gorm"
)

// UserGroupRepositoryImpl implements the UserGroupRepository interface using SQL database
type UserGroupRepositoryImpl struct {
	db *gorm.DB
}

// NewUserGroupRepository creates a new UserGroupRepository implementation
func NewUserGroupRepository(db *gorm.DB) repository.UserGroupRepository {
	return &UserGroupRepositoryImpl{
		db: db,
	}
}

// Create creates a new user group
func (r *UserGroupRepositoryImpl) Create(ctx context.Context, userGroup *auth.UserGroup) error {
	dbUserGroup := &sqldb.UserGroupDB{
		Name:        userGroup.Name,
		SiteAdmin:   userGroup.SiteAdmin,
		Description: userGroup.Description,
	}

	result := r.db.WithContext(ctx).Create(dbUserGroup)
	if result.Error != nil {
		return result.Error
	}

	userGroup.ID = dbUserGroup.ID
	return nil
}

// FindByID finds a user group by ID
func (r *UserGroupRepositoryImpl) FindByID(ctx context.Context, id int) (*auth.UserGroup, error) {
	var dbUserGroup sqldb.UserGroupDB
	err := r.db.WithContext(ctx).First(&dbUserGroup, id).Error
	if err != nil {
		return nil, err
	}

	return &auth.UserGroup{
		ID:          dbUserGroup.ID,
		Name:        dbUserGroup.Name,
		SiteAdmin:   dbUserGroup.SiteAdmin,
		Description: dbUserGroup.Description,
	}, nil
}

// FindByName finds a user group by name
func (r *UserGroupRepositoryImpl) FindByName(ctx context.Context, name string) (*auth.UserGroup, error) {
	var dbUserGroup sqldb.UserGroupDB
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&dbUserGroup).Error
	if err != nil {
		return nil, err
	}

	return &auth.UserGroup{
		ID:          dbUserGroup.ID,
		Name:        dbUserGroup.Name,
		SiteAdmin:   dbUserGroup.SiteAdmin,
		Description: dbUserGroup.Description,
	}, nil
}

// Update updates a user group
func (r *UserGroupRepositoryImpl) Update(ctx context.Context, userGroup *auth.UserGroup) error {
	dbUserGroup := &sqldb.UserGroupDB{
		ID:          userGroup.ID,
		Name:        userGroup.Name,
		SiteAdmin:   userGroup.SiteAdmin,
		Description: userGroup.Description,
	}

	return r.db.WithContext(ctx).Save(dbUserGroup).Error
}

// Delete deletes a user group by ID
func (r *UserGroupRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&sqldb.UserGroupDB{}, id).Error
}

// List returns all user groups
func (r *UserGroupRepositoryImpl) List(ctx context.Context) ([]*auth.UserGroup, error) {
	var dbUserGroups []sqldb.UserGroupDB
	err := r.db.WithContext(ctx).Find(&dbUserGroups).Error
	if err != nil {
		return nil, err
	}

	userGroups := make([]*auth.UserGroup, len(dbUserGroups))
	for i, dbUserGroup := range dbUserGroups {
		userGroups[i] = &auth.UserGroup{
			ID:          dbUserGroup.ID,
			Name:        dbUserGroup.Name,
			SiteAdmin:   dbUserGroup.SiteAdmin,
			Description: dbUserGroup.Description,
		}
	}

	return userGroups, nil
}

// FindSiteAdminGroups returns all site admin user groups
func (r *UserGroupRepositoryImpl) FindSiteAdminGroups(ctx context.Context) ([]*auth.UserGroup, error) {
	var dbUserGroups []sqldb.UserGroupDB
	err := r.db.WithContext(ctx).Where("site_admin = ?", true).Find(&dbUserGroups).Error
	if err != nil {
		return nil, err
	}

	userGroups := make([]*auth.UserGroup, len(dbUserGroups))
	for i, dbUserGroup := range dbUserGroups {
		userGroups[i] = &auth.UserGroup{
			ID:          dbUserGroup.ID,
			Name:        dbUserGroup.Name,
			SiteAdmin:   dbUserGroup.SiteAdmin,
			Description: dbUserGroup.Description,
		}
	}

	return userGroups, nil
}

// GetNamespacePermissions returns all namespace permissions for a user group
func (r *UserGroupRepositoryImpl) GetNamespacePermissions(ctx context.Context, userGroupID int) ([]auth.NamespacePermission, error) {
	var dbPermissions []sqldb.UserGroupNamespacePermissionDB
	err := r.db.WithContext(ctx).Where("user_group_id = ?", userGroupID).Find(&dbPermissions).Error
	if err != nil {
		return nil, err
	}

	permissions := make([]auth.NamespacePermission, len(dbPermissions))
	for i, dbPerm := range dbPermissions {
		permissions[i] = &UserGroupNamespacePermission{
			ID:             dbPerm.ID,
			UserGroupID:    dbPerm.UserGroupID,
			NamespaceID:    dbPerm.NamespaceID,
			PermissionType: auth.PermissionType(dbPerm.PermissionType),
		}
	}

	return permissions, nil
}

// GetGroupsForUser returns groups for a user (placeholder implementation)
func (r *UserGroupRepositoryImpl) GetGroupsForUser(ctx context.Context, userID string) ([]*auth.UserGroup, error) {
	// This is a placeholder implementation
	// In a real system, you would have a user_groups table mapping users to groups
	// For now, return empty list
	return []*auth.UserGroup{}, nil
}

// CreateNamespacePermission creates a new namespace permission for a user group
func (r *UserGroupRepositoryImpl) CreateNamespacePermission(ctx context.Context, userGroupID, namespaceID int, permissionType auth.PermissionType) error {
	dbPermission := &sqldb.UserGroupNamespacePermissionDB{
		UserGroupID:    userGroupID,
		NamespaceID:    namespaceID,
		PermissionType: string(permissionType),
	}

	return r.db.WithContext(ctx).Create(dbPermission).Error
}

// DeleteNamespacePermission deletes a namespace permission for a user group
func (r *UserGroupRepositoryImpl) DeleteNamespacePermission(ctx context.Context, userGroupID, namespaceID int) error {
	return r.db.WithContext(ctx).
		Where("user_group_id = ? AND namespace_id = ?", userGroupID, namespaceID).
		Delete(&sqldb.UserGroupNamespacePermissionDB{}).Error
}

// WithTransaction returns a new repository instance that works within the given transaction
func (r *UserGroupRepositoryImpl) WithTransaction(tx interface{}) repository.UserGroupRepository {
	// Implementation would depend on the transaction interface used
	// This is a placeholder for transaction support
	return r
}

// UserGroupNamespacePermission implements the NamespacePermission interface
type UserGroupNamespacePermission struct {
	ID             int
	UserGroupID    int
	NamespaceID    int
	PermissionType auth.PermissionType
}

// GetID returns the permission ID
func (p *UserGroupNamespacePermission) GetID() int {
	return p.ID
}

// GetUserGroupID returns the user group ID
func (p *UserGroupNamespacePermission) GetUserGroupID() int {
	return p.UserGroupID
}

// GetNamespaceID returns the namespace ID
func (p *UserGroupNamespacePermission) GetNamespaceID() int {
	return p.NamespaceID
}

// GetPermissionType returns the permission type
func (p *UserGroupNamespacePermission) GetPermissionType() auth.PermissionType {
	return p.PermissionType
}